package wasm

/*
#cgo CFLAGS: -Iinclude

#include "wasi_core.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"unsafe"

	common "my-own-cluster/common"

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
	Orchestrator *common.Orchestrator

	Name string
	Mode string

	WasmBytes     []byte
	StartFunction string

	InputBuffer []byte

	HasFinishedRunning     bool
	Result                 int
	OutputExchangeBufferID int

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

func CreateWasmContext(o *common.Orchestrator, mode string, functionName string, startFunction string, wasmBytes []byte, input []byte, outputExchangeBufferID int) *WasmProcessContext {
	wctx := &WasmProcessContext{
		Orchestrator:           o,
		Name:                   functionName,
		Mode:                   mode,
		WasmBytes:              wasmBytes,
		InputBuffer:            input,
		OutputExchangeBufferID: outputExchangeBufferID,
		StartFunction:          startFunction,
		APIPlugins:             []APIPlugin{},
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
	l := common.Min(len(buffer), len(vf.ReadBuffer)-vf.ReadPos)
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
	l := common.Min(len(buffer), len(vf.Wctx.InputBuffer)-vf.ReadPos)
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
	exchangeBuffer := vf.Wctx.Orchestrator.GetExchangeBuffer(vf.Wctx.OutputExchangeBufferID)
	written := exchangeBuffer.Write(buffer)
	vf.WritePos += written

	return written
}

func (vf *OutputAccessState) Close() int {
	exchangeBuffer := vf.Wctx.Orchestrator.GetExchangeBuffer(vf.Wctx.OutputExchangeBufferID)
	exchangeBuffer.Close()
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

	l := common.Min(len(buffer), len(*vf.Response)-vf.ReadPos)
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

func PorcelainPrepareWasm(o *common.Orchestrator, mode string, functionName string, startFunction string, wasmBytes []byte, inputBytes []byte, outputExchangeBufferID int, trace bool) (*WasmProcessContext, error) {
	wctx := CreateWasmContext(o, mode, functionName, startFunction, wasmBytes, inputBytes, outputExchangeBufferID)
	if wctx == nil {
		return nil, errors.New("cannot create wasm context")
	}

	wctx.Trace = trace

	wctx.AddAPIPlugin(NewMyOwnClusterAPIPlugin())
	wctx.AddAPIPlugin(NewTinyGoAPIPlugin())
	wctx.AddAPIPlugin(NewAutoLinkAPIPlugin())

	return wctx, nil
}

func PorcelainAddWASIPlugin(wctx *WasmProcessContext, wasiFileName string, arguments []string) {
	wctx.AddAPIPlugin(NewWASIHostPlugin(wasiFileName, arguments, map[int]VirtualFile{
		0: CreateStdInVirtualFile(wctx, wctx.InputBuffer),
		1: wctx.Orchestrator.GetExchangeBuffer(wctx.OutputExchangeBufferID), // CreateStdOutVirtualFile(wctx)
		2: CreateStdErrVirtualFile(wctx),
	}))
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

	// TODO move that in the MyOwnCluster API plugin
	// auto import and dynamically link functions together
	// TODO watch for updates on https://webassembly.org/docs/dynamic-linking/
	for m := range wctx.GetImportedModules() {
		if moduleFunctionTechID, ok := wctx.Orchestrator.GetFunctionTechIDFromName(m); ok {
			fmt.Printf("emulating %s imported module with function %s techID:%s...\n", m, m, moduleFunctionTechID)

			for i := 0; i < wctx.Module.NumFunctions(); i++ {
				f, err := wctx.Module.GetFunction(uint(i))
				if err != nil {
					continue
				}

				iModule := f.GetImportModule()
				iField := f.GetImportField()
				iSignature := f.GetSignature()
				if iModule != nil && *iModule == m && iField != nil {
					fmt.Printf("- imports func %s '%s' from module %s\n", *iField, iSignature, *iModule)
					wasmBytes, ok := wctx.Orchestrator.GetFunctionBytesByFunctionName(m)
					if !ok {
						fmt.Printf("error: can't find sub function bytes (%s)\n", m)
						continue
					}

					// TODO check imported signature is the same as exported signature...

					wctx.BindAPIFunction(m, *iField, iSignature, func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
						parameters := make([]int, f.GetNumArgs())
						for i := 0; i < int(f.GetNumArgs()); i++ {
							parameters[i] = int(cs.GetParamUINT32(i))
						}

						outputExchangeBufferID := wctx.Orchestrator.CreateExchangeBuffer()

						subWctx, err := PorcelainPrepareWasm(
							wctx.Orchestrator,
							"direct",
							m,
							*iField,
							wasmBytes,
							[]byte{},
							outputExchangeBufferID,
							wctx.Trace,
						)
						if err != nil {
							return 0xffff, err
						}

						subWctx.Run(parameters)

						return uint32(subWctx.Result), nil
					})
				}
			}

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

	wctx.HasFinishedRunning = true
	wctx.Result = result

	return wctx.Result, err
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
