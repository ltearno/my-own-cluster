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

/**
 * GUEST MEMORY LAYOUT
 */

// as for now, only 0x000000 to 0x1ff000 memory is mapped, you cannot go beyond yet !
// this has to do with the buildMmuTablesXXX(..) functions

#define MMU_TABLES_ADDRESS  0x01000
#define GDT_ADDRESS         0x05000
#define STACK_ADDRESS       0x06000
#define CODE_GUEST_ADDRESS  0x10000

#define STACK_SIZE          (CODE_GUEST_ADDRESS - STACK_ADDRESS)

/**
 * Few useful definitions
 */

#define EFER_LMA            0x400
#define EFER_LME            0x100
#define X86_CR0_PE          0x1
#define X86_CR0_PG          0x80000000
#define X86_CR4_PAE         0x20
#define RFLAGS_IF_BIT       (1 << 9)

// from Linux kernel source code
#define GDT_ENTRY(flags, base, limit)               \
    ((((base)  & _AC(0xff000000,ULL)) << (56-24)) | \
     (((flags) & _AC(0x0000f0ff,ULL)) << 40) |      \
     (((limit) & _AC(0x000f0000,ULL)) << (48-16)) | \
     (((base)  & _AC(0x00ffffff,ULL)) << 16) |      \
     (((limit) & _AC(0x0000ffff,ULL))))

// print a value in binary format
void printfBinary(int v) {
    unsigned int mask=1<<((sizeof(int)<<3)-1);
    while(mask) {
        printf("%d", (v&mask ? 1 : 0));
        mask >>= 1;
    }
    printf("\n");
}

// show a vcpu registers content
void dumpRegisters(int vcpufd) {
    struct kvm_regs regs;
    struct kvm_sregs sregs;
    int ret;
    // vcpufd represents the state of the vcpu file descriptor. The id is actually passed when calling the KVM_NEW_VCPU ioctl
    printf("registers for vcpu %d:\n", vcpufd);

    ret = ioctl(vcpufd, KVM_GET_REGS, &regs);
    if (ret == -1)
        err(1, "KVM_GET_REGS");

    ret = ioctl(vcpufd, KVM_GET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_GET_SREGS");

    printf(" rax:%016llx, rbx:%016llx, rcx:%016llx, rdx:%016llx, rip:%016llx\n", regs.rax, regs.rbx, regs.rcx, regs.rdx, regs.rip);
    printf(" rsp:%016llx\n", regs.rsp);
    printf(" cr0:%016llx, cr2:%016llx, cr3:%016llx, cr4:%016llx, cr8:%016llx\n", sregs.cr0, sregs.cr2, sregs.cr3, sregs.cr4, sregs.cr8);
    printf(" es: %016llx\n", sregs.es.base);
    printf(" cr0:   ");
    printfBinary(sregs.cr0);
    printf(" cr4:   ");
    printfBinary(sregs.cr4);
    printf(" rflags:");
    printfBinary(regs.rflags);
}

// loads a binary file, allocate and return the file content
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
    hexDump(buffer, stat.st_size);
    printf( "\n");

    return buffer;
}

// dump a buffer as bytes in hexa
void hexDump(void* buffer, int len){
    if(len>16*5)
        len = 16*5+1;
    for(int i=0;i<len; i++ ){
        printf(" %02x", ((unsigned char*)buffer)[i]);
        if(i%16==15)
            printf("\n");
    }
    if(len==16*5+1)
        printf(" ...\n");
}

// get kvm_run struct associated with a vcpu run state
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

/**
 * Create a memory mapped region and add it in the specified slot of a kvm vcpu at a specified physical address
 * 
 * Returns the mapped memory
 */
void* createMemoryRegion(int vmfd, int slot, __u64 guestPhysicalAddress, int size) {
    // align mmapped size to 4kB page size
    int mmapSize = size;
    if(mmapSize % 0x1000)
        mmapSize = mmapSize - (mmapSize % 0x1000) + 0x1000;
    printf("mmap size: 0x%x -> 0x%x\n", size, mmapSize);

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

/**
 * Builds a very simple identity memory mapped paging tables using Huge pages on PDE level (level 2)
 * 
 * this means that only the first 1GB is availables,
 * only addresses from 0x0000000000000000 to 0x000000003FFFFFFF (9+9+12 = 30 bits) are mapped
 * mmuTablePhysicalAddress should be 0x1000 aligned
 * if mmuTablePhysicalAddress is given inside the mapped region (0x1ff000), the mmu tables will then be manipulable directly
 * if the mmuTablePhysicalAddress is above the mapped region, mmu tables will be innaccessible from the running program.
 * After a call to 'buildMmuTablesXXX' you should set cr3 to mmuTablePhysicalAddress
 * 
 * good page : https://wiki.osdev.org/Paging, https://wiki.osdev.org/Page_Tables#Long_mode_.2864-bit.29_page_map
 */
void buildMmuTablesHugePages(uint8_t* mmuTable, __u64 mmuTablePhysicalAddress) {
    *(__u64*)&mmuTable[0x0000] = ((mmuTablePhysicalAddress + 0x1000) & 0x000ffffffffff000) | 0x3;
    *(__u64*)&mmuTable[0x1000] = ((mmuTablePhysicalAddress + 0x2000) & 0x000ffffffffff000) | 0x3;
    
    // level 2 table (pde) begins at 0x3000 and takes 0x1000 bytes
    // doc https://software.intel.com/sites/default/files/managed/7c/f1/253668-sdm-vol-3a.pdf table 4.17
    __u64* l2table = (__u64*)&mmuTable[0x2000];
    for(int i=0; i<0x1000/sizeof(__u64); i++)
        l2table[i] = ((i << 21) & 0x000ffffffffff000) | 0x83; // PRESENT, WRITABLE, PS (to set huge page)
}

/**
 * Builds a very simple identity memory mapped paging tables.
 * 
 * it does not use huge pages, and only the level 1 pages have multiple entries
 * this means that only the first 1MB is accessibles,
 * only addresses from 0x0000000000000000 to 0x00000000001FF000 are mapped
 * mmuTablePhysicalAddress should be 0x1000 aligned
 * if mmuTablePhysicalAddress is given inside the mapped region (0x1ff000), the mmu tables will then be manipulable directly
 * if the mmuTablePhysicalAddress is above the mapped region, mmu tables will be innaccessible from the running program.
 * After a call to 'buildMmuTablesXXX' you should set cr3 to mmuTablePhysicalAddress
 * 
 * good page : https://wiki.osdev.org/Paging, https://wiki.osdev.org/Page_Tables#Long_mode_.2864-bit.29_page_map
 */
void buildMmuTablesNormalPages(uint8_t* mmuTable, __u64 mmuTablePhysicalAddress) {
    *(__u64*)&mmuTable[0x0000] = ((mmuTablePhysicalAddress + 0x1000) & 0x000ffffffffff000) | 0x3;
    *(__u64*)&mmuTable[0x1000] = ((mmuTablePhysicalAddress + 0x2000) & 0x000ffffffffff000) | 0x3;
    *(__u64*)&mmuTable[0x2000] = ((mmuTablePhysicalAddress + 0x3000) & 0x000ffffffffff000) | 0x3;

    // level 1 table begins at 0x3000 and takes 0x1000 bytes
    __u64* l1table = (__u64*)&mmuTable[0x3000];
    for(int i=0; i<0x1000/sizeof(__u64); i++)
        l1table[i] = ((i << 12) & 0x000ffffffffff000) | 0x3;
}

/**
 * GDT functions, from firecracker (and most probably rust-vmm)
 * 
 * they get flag values from gdt a entry
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

// fill a kvm_segment struct from GDT entries in x86 format
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

/**
 * Open KVM and ensure the API version is ok
 */
int openKvm() {
    int kvm = open("/dev/kvm", O_RDWR | O_CLOEXEC);
    if (kvm == -1)
        err(1, "/dev/kvm (you can run : sudo setfacl -m u:${USER}:rw /dev/kvm)");
    printf("kvm opened, fd=%d\n", kvm);

    /* Make sure we have the stable version of the API */
    int ret = ioctl(kvm, KVM_GET_API_VERSION, NULL);
    if (ret == -1)
        err(1, "KVM_GET_API_VERSION");
    if (ret != KVM_API_VERSION)
        errx(1, "KVM_GET_API_VERSION %d, expected %d because this program was compiled against this version", ret, KVM_API_VERSION);
    printf("kvm version ok: %d\n", ret);

    return kvm;
}

/**
 * Here we go,
 * 
 * We load a binary 64 bit code, spin up a kvm vcpu, set it to 64 bit mode with paging,
 * pae, empty idt, gdt and so on.
 * We then run it and implement some exited MMIO exchange.
 * 
 * Use of io eventfd is coming soon. We have to see if we can take the code from virtio-mmio 
 * in the Linux kernel source code
 * 
 * This program will fail and exit as soon as a slight error is detected.
 * This is on purpose because this program is mainly a playground for learning KVM.
 */
int main(int argc, char **argv)
{
    if(argc < 2)
        err(1, "you should give the program one or two arguments : the binary executable code file name and the start address in hexa (0 by default)");

    printf("reading binary executable code file '%s'\n", argv[1]);
    int codeSize = 0;
    char *code = loadBinary(argv[1], &codeSize);
    if( ! codeSize ){
        err(1, "empty code");
    }

    unsigned long long startAddress = 0;
    if(argc >= 3)
        sscanf(argv[2], "%llx", &startAddress);
    printf("start address: 0x%llx\n", startAddress);
    
    int kvm = openKvm();

    // create a VM
    int vmfd = ioctl(kvm, KVM_CREATE_VM, (unsigned long)0);
    if (vmfd == -1)
        err(1, "KVM_CREATE_VM");

    // create and init the code memory region
    uint8_t* mem = createMemoryRegion(vmfd, 0, CODE_GUEST_ADDRESS, codeSize);
    memcpy(mem, code, codeSize);
    free(code);

    // create and init the MMU paging tables memory region
    uint8_t* mmuTable = createMemoryRegion(vmfd, 1, MMU_TABLES_ADDRESS, 0x4000);
    memset(mmuTable, 0, 0x4000);
    buildMmuTablesHugePages(mmuTable, MMU_TABLES_ADDRESS);

    // create and init the GDT memory region
    // https://wiki.osdev.org/Global_Descriptor_Table
    const int GDT_SIZE = 8 * 4;
    uint64_t* gdt = createMemoryRegion(vmfd, 2, GDT_ADDRESS, GDT_SIZE);
    gdt[0] = GDT_ENTRY(0x0000, 0, 0x0000);  // NULL
    gdt[1] = GDT_ENTRY(0xa09b, 0, 0xfffff); // CODE segment
    gdt[2] = GDT_ENTRY(0xc093, 0, 0xfffff); // DATA segment
    gdt[3] = GDT_ENTRY(0x808b, 0, 0xfffff); // TaskState segment (might be useless for us...)

    // create a stack space
    uint8_t* stack = createMemoryRegion(vmfd, 3, STACK_ADDRESS, STACK_SIZE);
    memset(stack, 0xfe, STACK_SIZE);

    // create a VCPU in the VM
    int vcpufd = ioctl(vmfd, KVM_CREATE_VCPU, (unsigned long)0);
    if (vcpufd == -1)
        err(1, "KVM_CREATE_VCPU");

    struct kvm_run *run = getKvmCpuRunData(kvm, vcpufd);

    // TODO : disable irqs (why ?)
    //See below at KVM_GET_REGS

    struct kvm_sregs sregs;
    int ret = ioctl(vcpufd, KVM_GET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_GET_SREGS");
    
    // load a zero length IDT so that any NMI causes a triple fault.
    sregs.idt.base = 0;
    sregs.idt.limit = 0;
    // register the GDT
    sregs.gdt.base = GDT_ADDRESS;
    sregs.gdt.limit = GDT_SIZE - 1;
    // prepare common segment registers
    kvm_segment_from_gdt(gdt[1], 1, &sregs.cs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.ds);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.es);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.fs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.gs);
    kvm_segment_from_gdt(gdt[2], 2, &sregs.ss);
    kvm_segment_from_gdt(gdt[3], 3, &sregs.tr);    
    // 64-bit protected mode with pagination and PAE,
    // you should know why you put those flags !
    // go here for details on each bit : https://wiki.osdev.org/Global_Descriptor_Table
    sregs.efer |= EFER_LMA | EFER_LME;
    sregs.cr0 |= X86_CR0_PG | X86_CR0_PE;
    sregs.cr3 = MMU_TABLES_ADDRESS;
    sregs.cr4 |= X86_CR4_PAE;

    ret = ioctl(vcpufd, KVM_SET_SREGS, &sregs);
    if (ret == -1)
        err(1, "KVM_SET_SREGS");

    // position rip to the start of our program and provide a stack pointer
    struct kvm_regs regs = {
        .rip = CODE_GUEST_ADDRESS + startAddress,
        .rsp = STACK_ADDRESS + STACK_SIZE,
        .rbp = STACK_ADDRESS + STACK_SIZE, // seen in arch/src/x86_64/regs.rs of firecracker/chromium os
        .rflags = RFLAGS_IF_BIT,
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
                    run->mmio.data[0] = 0x12;
                    printf("we simlulate having a value in memory (%x)\n", run->mmio.data[0]);
                }
                break;

            case KVM_EXIT_HLT:
                printf("KVM_EXIT_HLT, the vcpu has exited, finished\n");

                dumpRegisters(vcpufd);

                printf("stack content (last 64 bytes):\n");
                hexDump(stack + STACK_SIZE - 64, 64);
                printf( "\n");

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
                printf("KVM_EXIT_INTERNAL_ERROR: suberror = 0x%x\n", run->internal.suberror);

                dumpRegisters(vcpufd);

                printf("stack content (last 64 bytes):\n");
                hexDump(stack + STACK_SIZE - 64, 64);
                printf( "\n");

            default:
                errx(1, "exit_reason = 0x%x", run->exit_reason);
        }

        dumpRegisters(vcpufd);
    }
}
