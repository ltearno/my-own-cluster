package wasm

import (
	"fmt"
	"my-own-cluster/tools"
	"unsafe"
)

// CallSite represents the call information (memory start and stack pointer)
type CallSite struct {
	sp  unsafe.Pointer
	mem unsafe.Pointer
}

// GetParamUINT32 retruevs uint
func (cs *CallSite) GetParamUINT32(index int) uint32 {
	return getParameter(cs.sp, index)
}

// GetParamInt returns int
func (cs *CallSite) GetParamInt(index int) int {
	return int(getParameter(cs.sp, index))
}

// GetParamPointer retrive pointer
func (cs *CallSite) GetParamPointer(index int) unsafe.Pointer {
	return m3ApiOffsetToPtr(cs.mem, getParameter(cs.sp, index))
}

// GetParamByteBuffer GetParamByteBuffer
func (cs *CallSite) GetParamByteBuffer(addrParamIndex int, lengthParamIndex int) []byte {
	addr := cs.GetParamPointer(addrParamIndex)
	size := cs.GetParamUINT32(lengthParamIndex)

	if size == 0 || addr == nil {
		return nil
	}

	bytes := (*[1 << 30]byte)(addr)[:size:size]

	return bytes
}

// GetParamString GetParamString
func (cs *CallSite) GetParamString(addrParamIndex int, lengthParamIndex int) string {
	bytes := cs.GetParamByteBuffer(addrParamIndex, lengthParamIndex)

	return string(bytes)
}

// GetParamUINT32Ptr GetParamUINT32Ptr
func (cs *CallSite) GetParamUINT32Ptr(index int) *uint32 {
	return (*uint32)(cs.GetParamPointer(index))
}

// Print ptins
func (cs *CallSite) Print() {
	lineLength := 8
	count := 16

	chars := ""
	for i := 0; i < count; i++ {
		if i%lineLength == 0 {
			fmt.Printf("%08x: ", unsafe.Pointer(uintptr(cs.sp)+uintptr(i)))
			chars = ""
		}

		b := *(*byte)(unsafe.Pointer(uintptr(cs.sp) + uintptr(i)))
		fmt.Printf("%02x ", b)
		if tools.IsLetter(string(b)) {
			chars = chars + string(b)
		} else {
			chars = chars + "."
		}

		if (i+1)%lineLength == 0 {
			fmt.Printf(" %s\n", chars)
		}
	}

	fmt.Println()
}

func getParameter(sp unsafe.Pointer, index int) uint32 {
	return *(*uint32)(unsafe.Pointer(uintptr(sp) + uintptr(8)*uintptr(index)))
}

func m3ApiOffsetToPtr(mem unsafe.Pointer, offset uint32) unsafe.Pointer {
	return unsafe.Pointer(uintptr(mem) + uintptr(offset))
}

func m3ApiPtrToOffset(mem unsafe.Pointer, ptr unsafe.Pointer) uint32 {
	return uint32(uintptr(ptr) - uintptr(mem))
}
