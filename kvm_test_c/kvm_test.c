#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/ioctl.h>
#include <linux/kvm.h>
#include <sys/mman.h>
#include <sys/types.h>

int main(int argc, char* argv[]) {

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
	
	//create vm
	//May want to remplace ARM IPA size with 0 to ensure backwards comp
	int vm_fd = ioctl(devkvm_fd, KVM_CREATE_VM, 0);
	if(vm_fd == -1) {
		printf("error");
		return 1;
	}

	//create vcpu
	int vcpu_fd = ioctl(vm_fd, KVM_CREATE_VCPU, 0);
	if(vm_fd == -1) {
		printf("error");
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
	int mem_fd;
	mem_fd = open("hw", O_RDWR);
	__u64 *memory = mmap(NULL, 4096, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_PRIVATE, mem_fd, 0);
	if(memory == MAP_FAILED) {
		printf("error mem map");
		close(vcpu_fd);
		close(devkvm_fd);
		return 1;
	}

	//test: from scratch
	//works
	//__u64 *memory = mmap(NULL, 4096, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_SHARED | MAP_ANONYMOUS, -1, 0);

	//Tell kvm about this memory space
	struct kvm_userspace_memory_region region = {
		.slot = 0,
		.guest_phys_addr = 4096,
		.memory_size = 4096,
		.userspace_addr = memory,
	};
	int memset_rc = ioctl(vm_fd, KVM_SET_USER_MEMORY_REGION, &region);
	if(memset_rc == -1) {
		printf("error set mem region");
	}

	//Get KVM_RUN implicit parameter block for the lulz
	int mmap_size = ioctl(vcpu_fd, KVM_GET_VCPU_MMAP_SIZE, NULL);
	struct kvm_run *kvm_run_parameters = mmap(NULL, mmap_size, PROT_READ, MAP_PRIVATE, vcpu_fd, 0);

	//ssize_t s = write(STDOUT_FILENO, memory, 1024);
	
	// Now, we have to set BOTH RIP and CS to the beginning of the previous segment (i.e. they have to take the value of "memory" or 0, not sure)

	//Set code segment (CS)
	//TODO: understand wtf: why set to 0, what is segment, etc
	struct kvm_sregs vcpu_sregs;
	int get_sregs_sc = ioctl(vcpu_fd, KVM_GET_SREGS, &vcpu_sregs);
	if(get_sregs_sc == -1) {
		printf("can't get sregs");
	}
	sregs.cs.base = 0;
	sregs.cs.selector = 0;
	int set_sregs_cs = ioctl(vcpu_fd, KVM_SET_SREGS, vcpu_sregs);
	if(set_sregs_sc == -1) {
		printf("can't set sregs");
	}

	//set instruction pointer (RIP) to "memory" variable
	//TODO: understand why set to "memory" and no 0
	struct kvm_regs vcpu_regs;
	//pass pointer to created object. if you make a pointer to struct, enjoy your segfault (out of experience)
	int get_regs_sc = ioctl(vcpu_fd, KVM_GET_REGS, &vcpu_regs);
	if(get_regs_sc == -1) {
		printf("can't get regs");
	}
	vcpu_regs.rip = memory;
	int set_regs_sc = ioctl(vcpu_fd, KVM_SET_REGS, vcpu_regs);
	if(set_regs_sc == -1) {
		printf("can't set regs");
	}



	//Pick a god and pray
	//TODO: if no errors, real code.
	int time_to_run = ioctl(vcpu_fd, KVM_RUN, NULL);
	printf("%d\n", time_to_run);


	//Dispose of borrowed memory
	//TODO: P sure we forgot shit
	munmap(kvm_run_parameters, mmap_size);
	munmap(memory, 1024);

	//Close all open resources
	close(vcpu_fd);
	close(vm_fd);
	close(devkvm_fd);

	return 0;
}
