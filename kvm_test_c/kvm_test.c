#include <err.h>
#include <errno.h>
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
#include <sys/stat.h>

/**
 * You can find the KVM documentation here :
 * 
 * https://www.kernel.org/doc/Documentation/virtual/kvm/api.txt
 * 
 * a snapshot of this file is available here : kvm-api.txt
 * 
 **/

int main(int argc, char* argv[]) {

	const uint8_t code[] = {
		0xb8, 0x2a, 0x00, 0x00, 0x00,
		0xf4,
	};	

	//func to read file size, thanks stackoverflow
	off_t fsize(const char *filename) {
		struct stat st; 
		if (stat(filename, &st) == 0) {
			return st.st_size;
		}
	    return -1; 
	}

	//Create the fd for the driver
	int devkvm_fd = open("/dev/kvm", O_RDWR | O_CLOEXEC);
	//Check that it is not negative
	if(devkvm_fd < 0) {
		return 1;
	}

	int ret = ioctl(devkvm_fd, KVM_CHECK_EXTENSION, KVM_CAP_USER_MEMORY);
    	if(ret == -1) {
		close(devkvm_fd);
		return 1;
	}
	//Get system kvm version for the lulz
	int kvm_api_version = ioctl(devkvm_fd, KVM_GET_API_VERSION, NULL);
	printf("version %d\n", kvm_api_version);
	
	// create vm, see doc section 4.2 KVM_CREATE_VM
	int vm_fd = ioctl(devkvm_fd, KVM_CREATE_VM, 0);
	if(vm_fd == -1) {
		printf("error create vm");
		return 1;
	}

	// create vcpu, see doc section 4.7 KVM_CREATE_VCPU
	int vcpu_fd = ioctl(vm_fd, KVM_CREATE_VCPU, 0);
	if(vm_fd == -1) {
		printf("error create vcpu");
		return 1;
	}

	//get xsave to try and identify wtf it is
	/*struct kvm_xsave xsave;
	int xsave_r = ioctl(vcpu_fd, KVM_GET_XSAVE, xsave);
	for(int i = 0; i < 1024; i++) {
		printf("%d\n", xsave.region[i]);
	}*/
	//xsave appears to be a mix of fixed and nondeterministic 

	//memory allocation
	//from file
	/*int mem_fd;
	mem_fd = open("hw", O_RDWR);
	//File size
	size_t mem_size = 0x8000; 
	//size_t mem_size = (size_t)fsize("hw");
	size_t prg_size = (size_t)fsize("hw");
	printf("mem size=%u\n", mem_size);
	printf("prg size=%u\n", prg_size);
	__u64 *memory = mmap(NULL, mem_size, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_PRIVATE, mem_fd, 0);
	if(memory == MAP_FAILED) {
		printf("error mem map");
		close(vcpu_fd);
		close(devkvm_fd);
		return 1;
	}*/

	//test: from scratch
	//works
	size_t mem_size = 4096;
	__u64 *memory = mmap(NULL, mem_size, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_SHARED | MAP_ANONYMOUS, -1, 0);
	memcpy(memory, code, sizeof(code));

	//Tell kvm about this memory space
	struct kvm_userspace_memory_region region = {
		.slot = 0,
		.guest_phys_addr = mem_size,
		.memory_size = mem_size,
		.userspace_addr = memory,
	};
	int memset_rc = ioctl(vm_fd, KVM_SET_USER_MEMORY_REGION, &region);
	if(memset_rc == -1) {
		printf("error set mem region");
	}

	

	//ssize_t s = write(STDOUT_FILENO, memory, 1024);
	
	// Now, we have to set BOTH RIP and CS to the beginning of the previous segment (i.e. they have to take the value of "memory" or 0, not sure)

	//Set code segment (CS)
	//Code segment orchestrates jumps, set to 0 to start at beginning of guest memory
	struct kvm_sregs vcpu_sregs;
	int get_sregs_sc = ioctl(vcpu_fd, KVM_GET_SREGS, &vcpu_sregs);
	if(get_sregs_sc == -1) {
		printf("can't get sregs");
	}
	vcpu_sregs.cs.base = 0;
	vcpu_sregs.cs.selector = 0;
	int set_sregs_sc = ioctl(vcpu_fd, KVM_SET_SREGS, &vcpu_sregs);
	if(set_sregs_sc == -1) {
		printf("can't set sregs");
	}

	//set instruction pointer (RIP) to "memory" variable, where guest memory begins
	struct kvm_regs vcpu_regs;
	//pass pointer to created object. if you make a pointer to struct, enjoy your segfault (out of experience)
	/*int get_regs_sc = ioctl(vcpu_fd, KVM_GET_REGS, &vcpu_regs);
	if(get_regs_sc == -1) {
		printf("can't get regs");
	}*/
	vcpu_regs.rip = memory;
	vcpu_regs.rax = 3;
	vcpu_regs.rbx = 0;
	vcpu_regs.rflags = 0x2;
	int set_regs_sc = ioctl(vcpu_fd, KVM_SET_REGS, &vcpu_regs);
	if(set_regs_sc == -1) {
		printf("can't set regs");
	}


	// get KVM_RUN implicit parameter block for interrupt handling (this is documented in doc section 4.10 KVM_RUN)
	// CAUTION! Contrary to all that might seem logical, this is a SYSTEM (/dev/kvm/ fd) ioctl and not a VCPU one
	int mmap_size = ioctl(devkvm_fd, KVM_GET_VCPU_MMAP_SIZE, NULL);
	printf("KVM_VCPU_MMAP_SIZE = %d\n", mmap_size);
	//Size problem: -1
	struct kvm_run *kvm_run_parameters = mmap(NULL, mmap_size, PROT_READ | PROT_WRITE, MAP_SHARED, vcpu_fd, 0);
	if (kvm_run_parameters<0) {
		printf("%u\n", errno);
	}
	printf("ready to run\n");
	//See doc section 4.10 KVM_RUN
	//Apparently, code has to run in a loop
	int run_ret;
	int run_flag = 1;
	while (run_flag) {
		run_ret = ioctl(vcpu_fd, KVM_RUN, NULL);
		if(run_ret == -1) {
			err(1, "KVM_RUN");
			break;
		}
		switch (kvm_run_parameters->exit_reason) {
			case KVM_EXIT_HLT:
				printf("Exit: hlt\n");
				puts("KVM_EXIT_HLT");
				run_flag = 0;
				break;
			default:
				//Exit reason 9 is "KVM_EXIT_FAIL_ENTRY"
				printf("oh shit exit reason = %x\n", kvm_run_parameters->exit_reason);
				run_flag = 0;
				break;
		}
	}
	//Read memory
	//ssize_t s = write(STDOUT_FILENO, memory, mem_size);
	printf("get regs\n");
	int get_regs_sc = ioctl(vcpu_fd, KVM_GET_REGS, &vcpu_regs);	
	
	printf("Reading register RAX:\n");
       	printf("%lld\n", vcpu_regs.rax);
	
	//more info on the KVM_EXIT_FAIL_ENTRY can be found in this struct
	//apparently 80000021 is invalid guest state
	printf("Failure entry: %x\n", kvm_run_parameters->fail_entry.hardware_entry_failure_reason);

	//Dispose of borrowed memory
	//TODO: P sure we forgot shit
	munmap(kvm_run_parameters, mmap_size);
	munmap(memory, mem_size);

	//Close all open resources
	close(vcpu_fd);
	close(vm_fd);
	close(devkvm_fd);

	return 0;

}
