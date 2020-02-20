/* Sample code for /dev/kvm API
 *
 * Copyright (c) 2015 Intel Corporation
 * Author: Josh Triplett <josh@joshtriplett.org>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */
#include <err.h>
#include <fcntl.h>
#include <linux/kvm.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <asm/types.h>
#include <linux/const.h>

// https://wiki.osdev.org/Global_Descriptor_Table

#define MMU_TABLES_ADDRESS  0x1000
#define GDT_ADDRESS         0x5000
#define CODE_GUEST_ADDRESS  0x8000

#define GDT_ENTRY(flags, base, limit)               \
    ((((base)  & _AC(0xff000000,ULL)) << (56-24)) | \
     (((flags) & _AC(0x0000f0ff,ULL)) << 40) |      \
     (((limit) & _AC(0x000f0000,ULL)) << (48-16)) | \
     (((base)  & _AC(0x00ffffff,ULL)) << 16) |      \
     (((limit) & _AC(0x0000ffff,ULL))))

void printfBinary(int v)
{
    unsigned int mask=1<<((sizeof(int)<<3)-1);
    while(mask) {
        printf("%d", (v&mask ? 1 : 0));
        mask >>= 1;
    }
    printf("\n");
}

void dumpRegisters(int vcpufd) {
    struct kvm_regs regs;
    struct kvm_sregs sregs;
    int ret;

    printf("registers:\n");

    ret = ioctl(vcpufd, KVM_GET_REGS, &regs);
    if (ret == -1)
        err(1, "KVM_GET_REGS");

    ret = ioctl(vcpufd, KVM_GET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_GET_SREGS");

    printf(" rax:%08llx, rdx:%08llx, rip:%08llx\n", regs.rax, regs.rdx, regs.rip);
    printf(" cr0:%08llx, cr2:%08llx, cr3:%08llx, cr4:%08llx, cr8:%08llx\n", sregs.cr0, sregs.cr2, sregs.cr3, sregs.cr4, sregs.cr8);
    printf(" es: %08llx\n", sregs.es.base);
    printf(" cr0:");
    printfBinary(sregs.cr0);
    printf(" cr4:");
    printfBinary(sregs.cr4);
}

char* loadBinary(char *fileName, int *bufferSize) {
    int fd = open(fileName, O_RDONLY);
    if (fd < 0)
        err(1, "can not open binary file\n");

    struct stat stat;
    fstat(fd, &stat);

    *bufferSize = stat.st_size;
    unsigned char* buffer = malloc(stat.st_size);

    int readden = read(fd, buffer, stat.st_size);
    printf("read size: %d\n", readden);

    printf("code bytes:\n");
    for(int i=0;i<stat.st_size; i++ ){
        printf(" %02x", buffer[i]);
        if(i%16==15)
            printf("\n");
    }
    printf( "\n");

    return buffer;
}

struct kvm_run* getKvmCpuRunData(int kvm, int vcpufd) {
    int mmap_size = ioctl(kvm, KVM_GET_VCPU_MMAP_SIZE, NULL);
    if (mmap_size == -1)
        err(1, "KVM_GET_VCPU_MMAP_SIZE");
    if (mmap_size < sizeof(struct kvm_run))
        errx(1, "KVM_GET_VCPU_MMAP_SIZE unexpectedly small");

    struct kvm_run *run = mmap(NULL, mmap_size, PROT_READ | PROT_WRITE, MAP_SHARED, vcpufd, 0);
    if (!run)
        err(1, "mmap vcpu");
    
    return run;
}

void* createMemoryRegion(int vmfd, int slot, __u64 guestPhysicalAddress, int size) {
    // align mmapped size to 4kB page size
    int mmapSize = size;
    if(mmapSize % 0x1000)
        mmapSize = mmapSize - (mmapSize % 0x1000) + 0x1000;

    printf("mmap size: %d\n", mmapSize);
    uint8_t *mem = mmap(NULL, mmapSize, PROT_READ | PROT_WRITE, MAP_SHARED | MAP_ANONYMOUS, -1, 0);
    if (!mem)
        err(1, "allocating guest memory");

    struct kvm_userspace_memory_region region = {
        .slot = slot,
        .guest_phys_addr = guestPhysicalAddress,
        .memory_size = mmapSize,
        .userspace_addr = (uint64_t)mem,
    };

    int ret = ioctl(vmfd, KVM_SET_USER_MEMORY_REGION, &region);
    if (ret == -1)
        err(1, "KVM_SET_USER_MEMORY_REGION");

    return mem;
}

// build a very simple one page (4kB) mmu identity map
// it does not use huge pages, and only one page is mapped
// mmuTablePhysicalAddress and physicalPointedPageAddress should be 0x1000 aligned
// after a call to that:
// - cr3 should be set to mmuTablePhysicalAddress
// - access to address 0x0000000000000000 will be mapped to physical address physicalPointedPageAddress
// - this is valid only for the first 0x1000 bytes, then memory is undefined
void buildMmuTables(uint8_t* mmuTable, __u64 mmuTablePhysicalAddress) {
    *(__u64*)&mmuTable[0x0000] = ((mmuTablePhysicalAddress + 0x1000) & 0x000ffffffffff000) | 0x3;
    *(__u64*)&mmuTable[0x1000] = ((mmuTablePhysicalAddress + 0x2000) & 0x000ffffffffff000) | 0x3;
    *(__u64*)&mmuTable[0x2000] = ((mmuTablePhysicalAddress + 0x3000) & 0x000ffffffffff000) | 0x3;

    // level 1 table begins at 0x3000 and takes 0x1000 bytes
    __u64* l1table = (__u64*)&mmuTable[0x3000];
    for(int i=0; i<0x1000/sizeof(__u64); i++)
        l1table[i] = ((i << 12) & 0x000ffffffffff000) | 0x3;
}




/*

GDT functions, from firecracker (and most probably rust-vmm)

*/

__u64 gdt_get_base(__u64 entry) {
    return ((((entry) & 0xFF00000000000000) >> 32)
        | (((entry) & 0x000000FF00000000) >> 16)
        | (((entry) & 0x00000000FFFF0000) >> 16));
}

__u32 gdt_get_limit(__u64 entry) {
    return (__u32) ((((entry) & 0x000F000000000000) >> 32) | ((entry) & 0x000000000000FFFF));
}

unsigned char gdt_get_g(__u64 entry) {
    return (unsigned char)((entry & 0x0080000000000000) >> 55);
}

unsigned char gdt_get_db(__u64 entry) {
    return (unsigned char)((entry & 0x0040000000000000) >> 54);
}

unsigned char gdt_get_l(__u64 entry) {
    return (unsigned char)((entry & 0x0020000000000000) >> 53);
}

unsigned char gdt_get_avl(__u64 entry) {
    return (unsigned char)((entry & 0x0010000000000000) >> 52);
}

unsigned char gdt_get_p(__u64 entry) {
    return (unsigned char)((entry & 0x0000800000000000) >> 47);
}

unsigned char gdt_get_dpl(__u64 entry) {
    return (unsigned char)((entry & 0x0000600000000000) >> 45);
}

unsigned char gdt_get_s(__u64 entry) {
    return (unsigned char)((entry & 0x0000100000000000) >> 44);
}

unsigned char gdt_get_type(__u64 entry) {
    return (unsigned char)((entry & 0x00000F0000000000) >> 40);
}

void kvm_segment_from_gdt(__u64 entry, int table_index, struct kvm_segment *segment) {
    segment->base = gdt_get_base(entry);
    segment->limit = gdt_get_limit(entry);
    segment->selector = table_index * 8;
    segment->type = gdt_get_type(entry);
    segment->present = gdt_get_p(entry);
    segment->dpl = gdt_get_dpl(entry);
    segment->db = gdt_get_db(entry);
    segment->s = gdt_get_s(entry);
    segment->l = gdt_get_l(entry);
    segment->g = gdt_get_g(entry);
    segment->avl = gdt_get_avl(entry);
    segment->padding = 0;
    segment->unusable = gdt_get_p(entry) == 0 ? 1 : 0;
}





int main(int argc, char **argv)
{
    int kvm, vmfd, vcpufd, ret;

    if(argc != 2)
        err(1, "you should give the program one argument : the binary executable code file name");

    printf("reading binary executable code file '%s'\n", argv[1]);
    int codeSize = 0;
    char *code = loadBinary(argv[1], &codeSize);
    if( ! codeSize ){
        err(1, "empty code");
    }

    // jumping to 64bit long mode directly :
    /*
    1) Build paging structures (PML4, PDPT, PD and PTs)
    2) Enable PAE in CR4
    3) Set CR3 so it points to the PML4
    4) Enable long mode in the EFER MSR
    5) Enable paging and protected mode at the same time (activate long mode)
    6) Load a GDT
    7) Do a "far jump" to some 64 bit code
    */

    
    kvm = open("/dev/kvm", O_RDWR | O_CLOEXEC);
    if (kvm == -1)
        err(1, "/dev/kvm (you can run : sudo setfacl -m u:${USER}:rw /dev/kvm)");
    printf("kvm opened, fd=%d\n", kvm);

    /* Make sure we have the stable version of the API */
    ret = ioctl(kvm, KVM_GET_API_VERSION, NULL);
    if (ret == -1)
        err(1, "KVM_GET_API_VERSION");
    if (ret != KVM_API_VERSION)
        errx(1, "KVM_GET_API_VERSION %d, expected %d because this program was compiled against this version", ret, KVM_API_VERSION);
    printf("kvm version ok: %d\n", ret);

    vmfd = ioctl(kvm, KVM_CREATE_VM, (unsigned long)0);
    if (vmfd == -1)
        err(1, "KVM_CREATE_VM");

    uint8_t* mem = createMemoryRegion(vmfd, 0, CODE_GUEST_ADDRESS, codeSize);
    memcpy(mem, code, codeSize);

    uint8_t* mmuTable = createMemoryRegion(vmfd, 1, MMU_TABLES_ADDRESS, 0x4000);
    memset(mmuTable, 0, 0x4000);
    buildMmuTables(mmuTable, MMU_TABLES_ADDRESS);

    const int GDT_SIZE = 8 * 4;
    uint64_t* gdt = createMemoryRegion(vmfd, 2, GDT_ADDRESS, GDT_SIZE);
    gdt[0] = GDT_ENTRY(0x0000, 0, 0x0000);  // NULL
    gdt[1] = GDT_ENTRY(0xa09b, 0, 0xfffff); // CODE segment
    gdt[2] = GDT_ENTRY(0xc093, 0, 0xfffff); // DATA segment
    gdt[3] = GDT_ENTRY(0x808b, 0, 0xfffff); // TaskState segment (might be useless for us...)

    vcpufd = ioctl(vmfd, KVM_CREATE_VCPU, (unsigned long)0);
    if (vcpufd == -1)
        err(1, "KVM_CREATE_VCPU");

    struct kvm_run *run = getKvmCpuRunData(kvm, vcpufd);

    // disable irqs...
    // lidt [IDT]                        ; Load a zero length IDT so that any NMI causes a triple fault.
    // cr4 to 10100000b (Set the PAE and PGE bit.)
    // cr3 to MMU_TABLES_ADDRESS
    //    mov ecx, 0xC0000080               ; Read from the EFER MSR. 
    //    rdmsr
    //    or eax, 0x00000100                ; Set the LME bit.
    //    wrmsr
    // cr0 |= 0x80000001 activate long mode by enabling paging and protection simultaneously.
    // lgdt [GDT.Pointer]                ; Load GDT.Pointer defined below.

    /*
    FROM FIRECRACKER

    const EFER_LMA: u64 = 0x400;
    const EFER_LME: u64 = 0x100;

    const X86_CR0_PE: u64 = 0x1;
    const X86_CR0_PG: u64 = 0x8000_0000;
    const X86_CR4_PAE: u64 = 0x20;
    */

    #define EFER_LMA 0x400
    #define EFER_LME 0x100
    #define X86_CR0_PE 0x1
    #define X86_CR0_PG 0x80000000
    #define X86_CR4_PAE 0x20

    struct kvm_sregs sregs;
    ret = ioctl(vcpufd, KVM_GET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_GET_SREGS");
    
    // load a zero length IDT so that any NMI causes a triple fault.
    sregs.idt.base = 0;
    sregs.idt.limit = 0;
    // populates the GDT
    sregs.gdt.base = GDT_ADDRESS;
    sregs.gdt.limit = GDT_SIZE - 1;
    kvm_segment_from_gdt(gdt[1], 1, &sregs.cs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.ds);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.es);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.fs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.gs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.ss);
    kvm_segment_from_gdt(gdt[3], 3, &sregs.tr);    
    // 64-bit protected mode with pagination and PAE
    sregs.efer |= EFER_LMA | EFER_LME;
    sregs.cr0 |= X86_CR0_PG | X86_CR0_PE;
    sregs.cr3 = MMU_TABLES_ADDRESS;
    sregs.cr4 |= X86_CR4_PAE;

    ret = ioctl(vcpufd, KVM_SET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_SET_SREGS");

    struct kvm_regs regs = {
        .rip = CODE_GUEST_ADDRESS,
    };
    ret = ioctl(vcpufd, KVM_SET_REGS, &regs);
    if (ret == -1)
        err(1, "KVM_SET_REGS");

    dumpRegisters(vcpufd);

    printf("\nstart kvm run loop\n");

    while (1) {
        printf("\nKVM_RUN => ");
        ret = ioctl(vcpufd, KVM_RUN, NULL);
        if (ret == -1)
            err(1, "KVM_RUN");

        switch (run->exit_reason) {
            case KVM_EXIT_MMIO:
                printf("KVM_EXIT_MMIO : the vcpu is %s %d byte(s) at address %016llx\n", run->mmio.is_write ? "writing" : "reading", run->mmio.len, run->mmio.phys_addr);
                if( run ->mmio.is_write) {
                    printf("the written data is : %016llx\n", *(unsigned long long*)(&run->mmio.data[0]));
                }
                else {
                    printf("we simlulate having the value 0x78 in memory\n");
                    run->mmio.data[0] = 0x78;
                }
                break;

            case KVM_EXIT_HLT:
                printf("KVM_EXIT_HLT, the vcpu has exited, finished\n");
                dumpRegisters(vcpufd);
                return 0;

            case KVM_EXIT_IO:
                if (run->io.direction == KVM_EXIT_IO_OUT && run->io.size == 1 && run->io.port == 0x3f8 && run->io.count == 1)
                    putchar(*(((char *)run) + run->io.data_offset));
                else
                    errx(1, "unhandled KVM_EXIT_IO");
                break;

            case KVM_EXIT_FAIL_ENTRY:
                errx(1, "KVM_EXIT_FAIL_ENTRY: hardware_entry_failure_reason = 0x%llx",
                    (unsigned long long)run->fail_entry.hardware_entry_failure_reason);

            case KVM_EXIT_INTERNAL_ERROR:
                errx(1, "KVM_EXIT_INTERNAL_ERROR: suberror = 0x%x", run->internal.suberror);

            default:
                errx(1, "exit_reason = 0x%x", run->exit_reason);
        }

        dumpRegisters(vcpufd);
    }
}