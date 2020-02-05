package wasm

import (
	common "my-own-cluster/common"
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

// GetParamPointer retrive pointer
func (cs *CallSite) GetParamPointer(index int) unsafe.Pointer {
	return m3ApiOffsetToPtr(cs.mem, getParameter(cs.sp, index))
}

// GetParamByteBuffer GetParamByteBuffer
func (cs *CallSite) GetParamByteBuffer(addrParamIndex int, lengthParamIndex int) []byte {
	addr := cs.GetParamPointer(addrParamIndex)
	size := cs.GetParamUINT32(lengthParamIndex)

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
	common.PrintStack(cs.sp, 16)
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
