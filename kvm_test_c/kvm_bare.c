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

// sudo setfacl -m u:${USER}:rw /dev/kvm

#define CODE_GUEST_ADDRESS 0x8000

// Our expected segment selectors.
#define __BOOT_CS 2
#define __BOOT_DS 3
#define __BOOT_TR1 4
#define __BOOT_TR2 5
const int BootCsSelector = __BOOT_CS * sizeof(__u64);
const int BootDsSelector = __BOOT_DS * sizeof(__u64);
const int BootTrSelector = __BOOT_TR1 * sizeof(__u64);

#define GDT_ENTRY(flags, base, limit)               \
    ((((base)  & _AC(0xff000000,ULL)) << (56-24)) | \
     (((flags) & _AC(0x0000f0ff,ULL)) << 40) |      \
     (((limit) & _AC(0x000f0000,ULL)) << (48-16)) | \
     (((base)  & _AC(0x00ffffff,ULL)) << 16) |      \
     (((limit) & _AC(0x0000ffff,ULL))))

// Our boot GDT table.
static inline void build_64bit_gdt(void* page) {
    __u64* entry = (__u64*)page;
    entry[__BOOT_CS] = GDT_ENTRY(0xa09a, 0, 0xfffff);
    entry[__BOOT_DS] = GDT_ENTRY(0xc092, 0, 0xfffff);
    entry[__BOOT_TR1] = GDT_ENTRY(0x808b, 0, 0xfffff);
    entry[__BOOT_TR2] = GDT_ENTRY(0x0000, 0, 0);
}


// Our page table builders.
static inline void build_pml4(
    void* page,
    __u64 pgd_addr,
    int size) {

    // We will only build a single entry.
    // Thus, we will only be able to address 512GB from within
    // the linux startup routines. I think we'll be okay.
    __u64* entry = (__u64*)page;

    // Encodes: present, writable
    *entry = (pgd_addr & 0x000ffffffffff000) | 0x3;
}
static inline void build_pgd(
    void* page,
    __u64 pde_addr,
    int size) {

    // Wait! Here we only build a single entry.
    // Uh-oh, that means our startup routine is now further
    // limited -- to only 1GB! We'll probably still be okay.
    __u64* entry = (__u64*)page;

    // Encodes: present, writable
    *entry = (pde_addr & 0x000ffffffffff000) | 0x3;
}
static inline void build_pde(
    void* page,
    int size) {

    // Here we build an entry for every possible
    // index. Each entry represents 2MB addressable.
    __u64* entry = (__u64*)page;
    int i = 0, max = size / sizeof(*entry);
    for (i = 0; i < max; i += 1) {
        // Encodes: PSE (2mb), present, writable
        entry[i] = ((((__u64)i) << 21) & 0x000fffffffe00000) | 0x83;
    }
}



/*void ffprintfBinary(int v) {
    char buf[65];
    itoa(v, buf, 2);
    printf("%s\n", buf);
}*/

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

    /* Allocate one aligned page of guest memory to hold the code. */
    int codeMMapSize = codeSize;
    if(codeMMapSize % 0x1000)
        codeMMapSize = codeMMapSize - (codeMMapSize % 0x1000) + 0x1000;
    printf("code mmap size: %d\n", codeMMapSize);
    uint8_t *mem = mmap(NULL, codeMMapSize, PROT_READ | PROT_WRITE, MAP_SHARED | MAP_ANONYMOUS, -1, 0);
    if (!mem)
        err(1, "allocating guest memory");
    memcpy(mem, code, codeSize);

    struct kvm_userspace_memory_region region = {
        .slot = 0,
        .guest_phys_addr = CODE_GUEST_ADDRESS,
        .memory_size = codeMMapSize,
        .userspace_addr = (uint64_t)mem,
    };
    ret = ioctl(vmfd, KVM_SET_USER_MEMORY_REGION, &region);
    if (ret == -1)
        err(1, "KVM_SET_USER_MEMORY_REGION");

    vcpufd = ioctl(vmfd, KVM_CREATE_VCPU, (unsigned long)0);
    if (vcpufd == -1)
        err(1, "KVM_CREATE_VCPU");

    struct kvm_run *run = getKvmCpuRunData(kvm, vcpufd);

    /* Initialize CS to point at 0, via a read-modify-write of sregs. */
    struct kvm_sregs sregs;
    ret = ioctl(vcpufd, KVM_GET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_GET_SREGS");
    sregs.cs.base = 0;
    sregs.cs.selector = 0;
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
                ioctl(vcpufd, KVM_GET_REGS, &regs);
                printf("KVM_EXIT_HLT, the vcpu has exited, finished\n");
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