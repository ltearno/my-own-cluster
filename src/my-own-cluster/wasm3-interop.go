package main

/*
#cgo CFLAGS: -Iinclude

#include "wasi_core.h"
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"unsafe"

	wasm3 "github.com/ltearno/go-wasm3"
)

var preopen = []string{
	"<stdin>",
	"<stdout>",
	"<stderr>",
	"./",
	"../",
}

const (
	WASI_ESUCCESS                  = C.__WASI_ESUCCESS
	WASI_EBADF                     = C.__WASI_EBADF
	WASI_PREOPENTYPE_DIR           = C.__WASI_PREOPENTYPE_DIR
	WASI_FILETYPE_CHARACTER_DEVICE = C.__WASI_FILETYPE_CHARACTER_DEVICE
	WASI_FILETYPE_DIRECTORY        = C.__WASI_FILETYPE_DIRECTORY
	WASI_FILETYPE_REGULAR_FILE     = C.__WASI_FILETYPE_REGULAR_FILE
)

func setWasiStat(fd int, fdStatAddr unsafe.Pointer) {
	fdStat := (*C.__wasi_fdstat_t)(fdStatAddr)
	if fd >= 0 && fd <= 2 {
		fdStat.fs_filetype = C.__WASI_FILETYPE_CHARACTER_DEVICE
	} else if fd == 3 || fd == 4 {
		fdStat.fs_filetype = C.__WASI_FILETYPE_DIRECTORY
	} else {
		fdStat.fs_filetype = C.__WASI_FILETYPE_REGULAR_FILE
	}
	fdStat.fs_flags = 0
	fdStat.fs_rights_base = C.ulong(0xfffffffffff)
	fdStat.fs_rights_inheriting = C.ulong(0xfffffffffff)
}

// WasmProcessContext represents a running WASM context
type WasmProcessContext struct {
	Orchestrator *Orchestrator

	Name string
	Mode string

	WasmBytes     []byte
	StartFunction string

	InputBuffer []byte

	Result       int
	OutputPortID int

	Runtime *wasm3.Runtime
	Module  *wasm3.Module

	APIPlugins []APIPlugin

	Trace bool
}

type APIPlugin interface {
	Bind(wctx *WasmProcessContext)
}

type VirtualFile interface {
	Read(buffer []byte) int
	Write(buffer []byte) int
	Close() int
}

func CreateWasmContext(o *Orchestrator, mode string, functionName string, startFunction string, input []byte, outputPortID int) *WasmProcessContext {
	techID, ok := o.GetFunctionTechIDFromName(functionName)
	if !ok {
		return nil
	}

	wasmBytes, ok := o.GetFunctionBytes(techID)
	if !ok {
		return nil
	}

	wctx := &WasmProcessContext{
		Orchestrator:  o,
		Name:          functionName,
		Mode:          mode,
		WasmBytes:     wasmBytes,
		InputBuffer:   input,
		OutputPortID:  outputPortID,
		StartFunction: startFunction,
		APIPlugins:    []APIPlugin{},
	}

	return wctx
}

type StdAccess struct {
	Name       string
	ReadBuffer []byte
	ReadPos    int
}

func CreateStdInVirtualFile(wctx *WasmProcessContext, buffer []byte) VirtualFile {
	return &StdAccess{
		Name:       "stdin",
		ReadBuffer: buffer,
	}
}

func CreateStdOutVirtualFile(wctx *WasmProcessContext) VirtualFile {
	return &StdAccess{
		Name: "stdout",
	}
}

func CreateStdErrVirtualFile(wctx *WasmProcessContext) VirtualFile {
	return &StdAccess{
		Name: "stderr",
	}
}

func (vf *StdAccess) Read(buffer []byte) int {
	l := min(len(buffer), len(vf.ReadBuffer)-vf.ReadPos)
	copy(buffer, (vf.ReadBuffer)[vf.ReadPos:][:l])
	vf.ReadPos = vf.ReadPos + l

	return l
}

func (vf *StdAccess) Write(buffer []byte) int {
	fmt.Printf("%s: %s\n", vf.Name, string(buffer))
	return len(buffer)
}

func (vf *StdAccess) Close() int {
	return 0
}

type InputAccessState struct {
	Wctx    *WasmProcessContext
	ReadPos int
}

func CreateInputVirtualFile(wctx *WasmProcessContext) VirtualFile {
	return &InputAccessState{
		ReadPos: 0,
		Wctx:    wctx,
	}
}

func (vf *InputAccessState) Read(buffer []byte) int {
	l := min(len(buffer), len(vf.Wctx.InputBuffer)-vf.ReadPos)
	copy(buffer, (vf.Wctx.InputBuffer)[vf.ReadPos:][:l])
	vf.ReadPos = vf.ReadPos + l

	return l
}

func (vf *InputAccessState) Write(buffer []byte) int {
	fmt.Printf("CALLED WRITE ON INPUT STREAM !\n")
	return 0
}

func (vf *InputAccessState) Close() int {
	return 0
}

type OutputAccessState struct {
	Wctx     *WasmProcessContext
	WritePos int
}

func CreateOutputVirtualFile(wctx *WasmProcessContext) VirtualFile {
	return &OutputAccessState{
		WritePos: 0,
		Wctx:     wctx,
	}
}

func (vf *OutputAccessState) Read(buffer []byte) int {
	fmt.Printf("CALLED READ ON OUTPUT STREAM !\n")
	return 0
}

func (vf *OutputAccessState) Write(buffer []byte) int {
	port := vf.Wctx.Orchestrator.GetOutputPort(vf.Wctx.OutputPortID)
	written := port.Write(buffer)
	vf.WritePos += written

	return written
}

func (vf *OutputAccessState) Close() int {
	port := vf.Wctx.Orchestrator.GetOutputPort(vf.Wctx.OutputPortID)
	port.Close()
	return 0
}

type WebAccessState struct {
	Path     string
	Response *[]byte
	ReadPos  int
}

func CreateWebAccessVirtualFile(path string) VirtualFile {
	return &WebAccessState{
		Path:     path,
		Response: nil,
		ReadPos:  0,
	}
}

func (vf *WebAccessState) Write(buffer []byte) int {
	fmt.Printf("CALLED WRITE ON WEB STREAM !\n")
	return 0
}

func (vf *WebAccessState) Read(buffer []byte) int {
	vf.readAll()

	l := min(len(buffer), len(*vf.Response)-vf.ReadPos)
	copy(buffer, (*vf.Response)[vf.ReadPos:][:l])
	vf.ReadPos = vf.ReadPos + l

	return l
}

func (vf *WebAccessState) Close() int {
	return 0
}

func (vf *WebAccessState) readAll() {
	if vf.Response != nil {
		return
	}

	resp, err := http.Get(vf.Path)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	vf.Response = &bytes
	vf.ReadPos = 0
}

func (wctx *WasmProcessContext) GetImportedModules() map[string]bool {
	importedModules := make(map[string]bool)
	for i := 0; i < wctx.Module.NumFunctions(); i++ {
		f, err := wctx.Module.GetFunction(uint(i))
		if err != nil {
			continue
		}

		iModule := f.GetImportModule()
		iField := f.GetImportField()
		if iModule != nil && iField != nil {
			importedModules[*iModule] = true
		}
	}

	return importedModules
}

func (wctx *WasmProcessContext) PrintImportedModules() {
	fmt.Printf("Imported functions\n")
	for i := 0; i < wctx.Module.NumFunctions(); i++ {
		f, err := wctx.Module.GetFunction(uint(i))
		if err != nil {
			continue
		}

		iModule := f.GetImportModule()
		if iModule != nil {
			iField := f.GetImportField()
			fmt.Printf(" - '%s'::'%s'\n", *iModule, *iField)
		}
	}
}

func (wctx *WasmProcessContext) AddAPIPlugin(plugin APIPlugin) {
	wctx.APIPlugins = append(wctx.APIPlugins, plugin)
}

// Run runs the process
func (wctx *WasmProcessContext) Run(arguments []int) (int, error) {
	wctx.Runtime = wasm3.NewRuntime(&wasm3.Config{
		Environment: wasm3.NewEnvironment(),
		StackSize:   64 * 1024,
		EnableWASI:  false,
	})

	{
		module, err := wctx.Runtime.ParseModule(wctx.WasmBytes)
		if err != nil {
			return -1, err
		}

		wctx.Module = module
	}

	_, err := wctx.Runtime.LoadModule(wctx.Module)
	if err != nil {
		return -2, err
	}

	if wctx.Trace {
		wctx.PrintImportedModules()
	}

	for _, plugin := range wctx.APIPlugins {
		plugin.Bind(wctx)
	}

	for m := range wctx.GetImportedModules() {
		if moduleFunctionTechID, ok := wctx.Orchestrator.GetFunctionTechIDFromName(m); ok {
			fmt.Printf("emulating %s imported module with function %s techID:%s...\n", m, m, moduleFunctionTechID)
			wctx.BindAPIFunction(m, "rustMultiply", "i(ii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
				a := cs.GetParamUINT32(0)
				b := cs.GetParamUINT32(1)

				fmt.Printf("  >- emulation called -<\n")

				portID := wctx.Orchestrator.CreateOutputPort()
				subWctx := CreateWasmContext(wctx.Orchestrator, "direct", m, "rustMultiply", []byte{}, portID)
				if subWctx == nil {
					return 42, nil
				}

				subWctx.AddAPIPlugin(NewMyOwnClusterAPIPlugin())

				subWctx.Run([]int{int(a), int(b)})

				fmt.Printf("SUB WCTX returned %d\n", subWctx.Result)

				return uint32(subWctx.Result), nil
			})
		}
	}

	// WITHOUT CALLING THIS, THE fn.Call() call fails ! need to investigate
	wctx.Runtime.FindFunction(wctx.StartFunction)

	fn, err := wctx.Module.GetFunctionByName(wctx.StartFunction)
	if err != nil {
		log.Printf("not found '%s' function (using module.GetFunctionByName)", wctx.StartFunction)
		return -3, err
	}

	fmt.Printf("calling function_name:\"%s\" start_function:\"%s\" mode:%s input_size:%d ...", wctx.Name, wctx.StartFunction, wctx.Mode, len(wctx.InputBuffer))

	wctx.Result = 0
	result, err := fn.Call2(arguments)
	if wctx.Mode != "posix" {
		wctx.Result = result
	}

	fmt.Printf(" result:%d, err:%+v\n", wctx.Result, err)

	wctx.Result = result

	// TODO temporary and hacky, should only be done if the port has not been redirected, a concept to clarify...
	wctx.Orchestrator.GetOutputPort(wctx.OutputPortID).Close()

	return wctx.Result, err
}

type TinyGoAPIPlugin struct{}

func NewTinyGoAPIPlugin() APIPlugin {
	return &TinyGoAPIPlugin{}
}

func (p *TinyGoAPIPlugin) Bind(wctx *WasmProcessContext) {
	importedModules := wctx.GetImportedModules()
	if _, ok := importedModules["env"]; !ok {
		return
	}

	//if wctx.Trace {
	fmt.Println("binding TinyGo 0.11.0 API...")
	//}

	wctx.Runtime.AttachFunction("env", "io_get_stdout", "i()", func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		if wctx.Trace {
			fmt.Printf("TinyGo talks to us !!!!\n")
		}

		*(*uint32)(sp) = 1

		return 0
	})

	wctx.BindAPIFunction("env", "resource_write", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		fd := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		if wctx.Trace {
			fmt.Printf("TinyGo on fd %d says '%s'\n", fd, string(buffer))
		}

		return uint32(len(buffer)), nil
	})

	wctx.BindNotYetImplementedFunction("env", "runtime.ticks", "i()")
	wctx.BindNotYetImplementedFunction("env", "runtime.sleepTicks", "i(i)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueLength", "i(*ii)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueCall", "i(i*iiiiiii)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueIndex", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueGet", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueNew", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueSet", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueSetIndex", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.stringVal", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valuePrepareString", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueLoadString", "i()")
}

type MyOwnClusterAPIPlugin struct{}

func NewMyOwnClusterAPIPlugin() APIPlugin {
	return &MyOwnClusterAPIPlugin{}
}

func (p *MyOwnClusterAPIPlugin) Bind(wctx *WasmProcessContext) {
	importedModules := wctx.GetImportedModules()
	if _, ok := importedModules["my-own-cluster"]; !ok {
		return
	}

	if wctx.Trace {
		fmt.Println("binding MyOwnCluster API...")
	}

	wctx.Runtime.AttachFunction("my-own-cluster", "test", "i()", func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		*(*uint32)(sp) = 0
		return 0
	})
}

// CallSite represents the call information (memory start and stack pointer)
type CallSite struct {
	sp  unsafe.Pointer
	mem unsafe.Pointer
}

// WasmCallHandler is the type of functions called back by wasm3 runtime
type WasmCallHandler func(wctx *WasmProcessContext, cs *CallSite) (uint32, error)

// BindAPIFunction binds a module+function name in wasm3 to a go routine
func (wctx *WasmProcessContext) BindAPIFunction(moduleName string, functionName string, signature string, handler WasmCallHandler) {
	wctx.Runtime.AttachFunction(moduleName, functionName, signature, func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		callSite := &CallSite{
			sp:  sp,
			mem: mem,
		}

		result, err := handler(wctx, callSite)
		if err != nil {
			fmt.Println("BOUND API ERROR !", err)
			return -1
		}

		*(*uint32)(sp) = result
		return 0
	})
}

// BindNotYetImplementedFunction exits the whole process when not yet implemented function is called
func (wctx *WasmProcessContext) BindNotYetImplementedFunction(module string, name string, signature string) {
	wctx.Runtime.AttachFunction(module, name, signature, func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		fmt.Printf("called wasi function '%s', but it is not yet implemented... ABORTING PROGRAM EXECUTION\n", name)
		return -2
	})
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
	printStack(cs.sp, 16)
}
