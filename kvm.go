package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

/*
#include <linux/kvm.h>
#include <sys/ioctl.h>
#include <errno.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/stat.h>
//For system call constants
*/
import "C"

//Cell is defined as a (VM)/CPU/Memory/(Jailer) ensemble.
//For now, nothing prevents us from using the same VM for all cells, which excludes it from this ensemble (but still, keep the fd here)
//The Jailer is a big TODO here.
type Cell struct {
	vmFd     int
	cpuFd    int
	memoryFd int
}

//CreateVM serves to make a new vmFd, which can then be used to spawn cells
func CreateVM() (int, error) {
	//Open /dev/kvm/
	//As it happens, syscall is depreciated. We use "golang.org/x/sys/unix" instead
	devkvm, err := unix.Open("/dev/kvm", C.O_RDWR|C.O_CLOEXEC, 0)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	//In case.
	defer unix.Close(devkvm)

	//For all the following, we are going to have to use the IOCTL
	//C IOCTL doesn't appear to work
	//Try IoctlGetInt and IoctlSetInt from lib unix
	//Note that with those, no need to pass an unsafe pointer. (see https://github.com/golang/sys/blob/master/unix/ioctl.go)

	//Check version
	version, err := unix.IoctlGetInt(devkvm, C.KVM_GET_API_VERSION)
	if err != nil {
		return -1, err
	}

	fmt.Printf("KVM version: %d\n", version)

	//Create vm
	vmFd, err := unix.IoctlGetInt(devkvm, C.KVM_CREATE_VM)
	if err != nil {
		return -1, err
	}
	defer unix.Close(vmFd)

	return vmFd, nil
}

/*
//CreateNewCell makes a new vCPU and memory band in a VM and packs it all tightly
func CreateNewCell(vmFd int, memFile string, memSize int) (*Cell, error) {
	//Handle memory
	memoryFd, memory, err := allocateMemory(vmFd, memFile, memSize)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//Create vcpu
	cpuFd, err := unix.IoctlGetInt(vmFd, C.KVM_CREATE_VCPU, 0)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//Set CS/IP
	//Set CS to 0
	var vcpuSregs C.struct_kvm_sregs
	if rc := unix.IoctlGetInt(cpuFd, C.KVM_GET_SREGS, &vcpuSregs); rc == -1 {
		err = errors.New("Can't get Sregs")
		return nil, err
	}
	vcpuSregs.cs.base = 0
	vcpuSregs.cs.selector = 0
	if err := unix.IoctlSetInt(cpuFd, C.KVM_SET_SREGS, &vcpuSregs); err != nill {
		err = errors.New("Can't set Sregs")
		return nil, err
	}
	//Set IP to memory
	var vcpuRegs C.struct_kvm_regs
	if rc := unix.IoctlGetInt(cpuFd, C.KVM_GET_REGS, &vcpuRegs); rc == -1 {
		err = errors.New("Can't get Regs")
		return nil, err
	}
	vcpuRegs.rip = memory
	if err := unix.IoctlSetInt(cpuFd, C.KVM_SET_REGS, &vcpuRegs); err != nil {
		err = errors.New("Can't set Regs")
		return nil, err
	}
	//Package and return the new cell
	newCell := &Cell{
		vmFd:     vmFd,
		cpuFd:    cpuFd,
		memoryFd: memoryFd,
	}
	return newCell, nil
}


//Apparently you require a file in go for memory allocation
func allocateMemory(vmFd int, memFile string, memSize int) (int, C.__u64, error) {
	//Open the file
	memoryFd, err := unix.Open(memFile, C.O_RDWR, 0)
	//Get file size, mmap current size + memsize
	fileInfo, err := memoryFd.Stat()
	if err != nil {
		fmt.Println(err)
		return 0, -1, err
	}
	//Data is a byte array. Not sure how useful mmap-ing is here
	fullsize := memSize + int(fileInfo.Size())
	//Go's mmap does not offer a pointer to the reserved memory. I'm using C's mmap to feed kvm the right pointer.
	memory, err := C.mmap(C.NULL, fullsize, C.PROT_READ|C.PROT_WRITE|C.PROT_EXEC, C.MAP_SHARED|C.MAP_ANONYMOUS, memoryFd, 0)
	if err != nil {
		fmt.Println(err)
		return 0, -1, err
	}
	//Feed kvm the details
	region := &C.struct_kvm_userspace_memory_region{
		slot:            C.__u32(0),
		guest_phys_addr: C.__u64(memory),
		memory_size:     C.__u64(memory),
		userspace_addr:  C.__u64(memory),
	}
	//This is supposed to set an int. Passed a pointer to struct, no idea how it will work. There's no such Ioctl func for structs.
	err = unix.IoctlSetInt(vmFd, C.KVM_SET_USER_MEMORY_REGION, region)
	if err != nil {
		fmt.Println(err)
		return 0, -1, err
	}
	return memoryFd, memory, nil
}

*/

func TestKVM() {
	vm, err := CreateVM()
	fmt.Println(vm)
	fmt.Println(err)
}
