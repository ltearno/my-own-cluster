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
	//As it happens, syscall is depreciated. We use "golang.org/x/sys/unix" in its stead
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
	version, err := C.ioctl(devkvm, C.KVM_GET_API_VERSION, 0)
	if err != nil || version != 12 {
		return -1, err
	}

	//Create vm
	vmFd, err := C.ioctl(devkvm, C.KVM_CREATE_VCPU, 0)
	if err != nil {
		return -1, err
	}
	defer unix.Close(vmFd)

	//Return it all
}

//CreateNewCell makes a new vCPU and memory band in a VM and packs it all tightly
func CreateNewCell(vmFd int) (*Cell, error) {
	return nil, nil
}

func main() {

	vm, err := CreateVM()
	fmt.Println(vm)
	fmt.Println(err)
}
