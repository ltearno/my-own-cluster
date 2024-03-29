package enginewasm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"unsafe"

	"github.com/ltearno/my-own-cluster/common"
	"github.com/ltearno/my-own-cluster/tools"

	wasm3 "github.com/ltearno/go-wasm3"
)

// WasmProcessContext represents a running WASM context
type WasmProcessContext struct {
	Fctx *common.FunctionExecutionContext

	Runtime *wasm3.Runtime
	Module  *wasm3.Module

	APIPlugins []WASMAPIPlugin

	//GuestAllocatorFn func(int) int
}

type WASMAPIPlugin interface {
	Bind(wctx *WasmProcessContext)
}

type VirtualFile interface {
	Read(buffer []byte) int
	Write(buffer []byte) (int, error)
	Close() int
}

type WasmWasm3Engine struct {
}

func NewWasmWasm3Engine() *WasmWasm3Engine {
	return &WasmWasm3Engine{}
}

func CreateWasmContext(fctx *common.FunctionExecutionContext) *WasmProcessContext {
	wctx := &WasmProcessContext{
		Fctx:       fctx,
		APIPlugins: []WASMAPIPlugin{},
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
	l := tools.Min(len(buffer), len(vf.ReadBuffer)-vf.ReadPos)
	copy(buffer, (vf.ReadBuffer)[vf.ReadPos:][:l])
	vf.ReadPos = vf.ReadPos + l

	return l
}

func (vf *StdAccess) Write(buffer []byte) (int, error) {
	fmt.Printf("%s: %s\n", vf.Name, string(buffer))
	return len(buffer), nil
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
	inputBuffer := vf.Wctx.Fctx.Orchestrator.GetExchangeBuffer(vf.Wctx.Fctx.InputExchangeBufferID)

	l := tools.Min(len(buffer), len(inputBuffer.GetBuffer())-vf.ReadPos)
	copy(buffer, (inputBuffer.GetBuffer())[vf.ReadPos:][:l])
	vf.ReadPos = vf.ReadPos + l

	return l
}

func (vf *InputAccessState) Write(buffer []byte) (int, error) {
	fmt.Printf("CALLED WRITE ON INPUT STREAM !\n")
	return 0, nil
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

func (vf *OutputAccessState) Write(buffer []byte) (int, error) {
	exchangeBuffer := vf.Wctx.Fctx.Orchestrator.GetExchangeBuffer(vf.Wctx.Fctx.OutputExchangeBufferID)
	written, _ := exchangeBuffer.Write(buffer)
	vf.WritePos += written

	return written, nil
}

func (vf *OutputAccessState) Close() int {
	exchangeBuffer := vf.Wctx.Fctx.Orchestrator.GetExchangeBuffer(vf.Wctx.Fctx.OutputExchangeBufferID)
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

func (vf *WebAccessState) Write(buffer []byte) (int, error) {
	fmt.Printf("CALLED WRITE ON WEB STREAM !\n")
	return 0, nil
}

func (vf *WebAccessState) Read(buffer []byte) int {
	vf.readAll()

	l := tools.Min(len(buffer), len(*vf.Response)-vf.ReadPos)
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

func (wctx *WasmProcessContext) AddAPIPlugin(plugin WASMAPIPlugin) {
	wctx.APIPlugins = append(wctx.APIPlugins, plugin)
}

func (e *WasmWasm3Engine) PrepareContext(fctx *common.FunctionExecutionContext) (common.ExecutionEngineContext, error) {
	wctx := CreateWasmContext(fctx)
	if wctx == nil {
		return nil, errors.New("cannot create wasm context")
	}

	return wctx, nil
}

type WrappedExchangeBufferInVirtualFile struct {
	buffer  []byte
	b       common.ExchangeBuffer
	readPos int
}

func WrapExchangeBufferInVirtualFile(b common.ExchangeBuffer) VirtualFile {
	return &WrappedExchangeBufferInVirtualFile{
		buffer: nil,
		b:      b,
	}
}

func (b WrappedExchangeBufferInVirtualFile) Read(buffer []byte) int {
	if b.buffer == nil {
		b.buffer = b.b.GetBuffer()
	}

	// refactored and not tested yet (2020/07/20)
	notRead := len(b.buffer) - b.readPos
	if notRead <= 0 {
		return 0
	}

	toRead := tools.Min(notRead, len(buffer))
	if toRead > 0 {
		copy(buffer[0:toRead], b.buffer[b.readPos:b.readPos+toRead])
		b.readPos += toRead
	}

	return toRead
}

func (b WrappedExchangeBufferInVirtualFile) Write(buffer []byte) (int, error) {
	res, err := b.b.Write(buffer)
	return res, err
}

func (b WrappedExchangeBufferInVirtualFile) Close() int {
	return b.Close()
}

// Run runs the process
func (wctx *WasmProcessContext) Run() error {
	wctx.Runtime = wasm3.NewRuntime(&wasm3.Config{
		Environment: wasm3.NewEnvironment(),
		StackSize:   64 * 1024, // original 64ko
		EnableWASI:  false,
	})

	if wctx.Fctx.Trace {
		wctx.Runtime.PrintRuntimeInfo()
	}

	{
		module, err := wctx.Runtime.ParseModule(wctx.Fctx.CodeBytes)
		if err != nil {
			return errors.New("cannot parse module")
		}

		wctx.Module = module
	}

	_, err := wctx.Runtime.LoadModule(wctx.Module)
	if err != nil {
		return errors.New("cannot load module")
	}

	if wctx.Fctx.Trace {
		wctx.PrintImportedModules()
	}

	for _, plugin := range wctx.APIPlugins {
		plugin.Bind(wctx)
	}

	// TODO move that in the MyOwnCluster API plugin
	// auto import and dynamically link functions together
	// TODO watch for updates on https://webassembly.org/docs/dynamic-linking/
	for m := range wctx.GetImportedModules() {
		apiProvider := wctx.Fctx.Orchestrator.GetAPIProvider(m)
		if apiProvider != nil {
			if wctx.Fctx.Trace {
				fmt.Printf("emulating '%s' imported module with api provider '%s'\n", m, m)
			}
			apiProvider.BindToExecutionEngineContext(wctx)
			continue
		}

		moduleFunctionTechID, err := wctx.Fctx.Orchestrator.GetBlobTechIDFromName(m)
		if err == nil {
			if wctx.Fctx.Trace {
				fmt.Printf("emulating '%s' imported module with function '%s' techID:%s\n", m, m, moduleFunctionTechID)
			}

			for i := 0; i < wctx.Module.NumFunctions(); i++ {
				f, err := wctx.Module.GetFunction(uint(i))
				if err != nil {
					continue
				}

				iModule := f.GetImportModule()
				iField := f.GetImportField()
				iSignature := f.GetSignature()
				if iModule != nil && *iModule == m && iField != nil {
					if wctx.Fctx.Trace {
						fmt.Printf("- imports func %s '%s' from module %s\n", *iField, iSignature, *iModule)
					}
					wasmBytes, err := wctx.Fctx.Orchestrator.GetBlobBytesByTechID(moduleFunctionTechID)
					if err != nil {
						fmt.Printf("error: can't find sub function bytes (%s)\n", m)
						continue
					}

					// TODO check imported signature is the same as exported signature...

					wctx.BindAPIFunction(m, *iField, iSignature, func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
						parameters := make([]int, f.GetNumArgs())
						for i := 0; i < int(f.GetNumArgs()); i++ {
							parameters[i] = int(cs.GetParamUINT32(i))
						}

						outputExchangeBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
						inputExchangeBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()

						subFctx := &common.FunctionExecutionContext{
							Name:                   m,
							StartFunction:          *iField,
							HasFinishedRunning:     false,
							InputExchangeBufferID:  inputExchangeBufferID,
							OutputExchangeBufferID: outputExchangeBufferID,
							Orchestrator:           wctx.Fctx.Orchestrator,
							Result:                 0,
							Mode:                   "direct",
							CodeBytes:              wasmBytes,
							Arguments:              parameters,
						}

						e := NewWasmWasm3Engine()
						subWctx, err := e.PrepareContext(subFctx)
						if err != nil {
							return 0xffff, err
						}

						subWctx.Run()

						return uint32(subFctx.Result), nil
					})
				}
			}
			continue
		}

		if wctx.Fctx.Mode == "posix" {
			if m == "wasi_unstable" || m == "wasi_snapshot_preview1" {
				if wctx.Fctx.Trace {
					fmt.Printf("emulating '%s' imported module with WASI runtime layer\n", m)
				}

				inputExchangeBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(wctx.Fctx.InputExchangeBufferID)

				// TODO : provide the WASI interface by a rust program doing it with the core api compiled to wasm and pushed as a wasm module (but we need api descriptions for that !)
				wasiHostPlugin := NewWASIHostPlugin(wctx.Fctx.POSIXFileName, wctx.Fctx.POSIXArguments, map[int]VirtualFile{
					0: CreateStdInVirtualFile(wctx, inputExchangeBuffer.GetBuffer()),
					1: WrapExchangeBufferInVirtualFile(wctx.Fctx.Orchestrator.GetExchangeBuffer(wctx.Fctx.OutputExchangeBufferID)),
					2: CreateStdErrVirtualFile(wctx),
				})
				wasiHostPlugin.Bind(wctx)

				continue
			}
		}

		if m == "env" {
			if wctx.Fctx.Trace {
				fmt.Printf("emulating '%s' imported module with TinyGo runtime layer\n", m)
			}

			BindTinyGoRuntimeAPI(wctx)
			continue
		}

		return fmt.Errorf("cannot emulate imported module '%s'", m)
	}

	// WITHOUT CALLING THIS, THE fn.Call() call fails ! need to investigate
	wctx.Runtime.FindFunction(wctx.Fctx.StartFunction)

	fn, err := wctx.Module.GetFunctionByName(wctx.Fctx.StartFunction)
	if err != nil {
		return fmt.Errorf("not found '%s' function (using module.GetFunctionByName)", wctx.Fctx.StartFunction)
	}

	/*guestAllocatorFn, err := wctx.Runtime.FindFunction("moc_allocator")
	if guestAllocatorFn != nil && err == nil {
		guestAllocatorFnWrapper := func(size int) int {
			ptr, err := guestAllocatorFn(size)
			if err != nil {
				return 0
			}
			return ptr
		}
		wctx.GuestAllocatorFn = guestAllocatorFnWrapper

		fmt.Printf("found GuestAllocatorFunction in wasm module (%v)\n", wctx.GuestAllocatorFn)
	}*/

	wctx.Fctx.Result = 0
	result, err := fn.Call2(wctx.Fctx.Arguments)
	if wctx.Fctx.Mode != "posix" {
		wctx.Fctx.Result = result
	}

	wctx.Fctx.HasFinishedRunning = true
	wctx.Fctx.Result = result

	return nil /*err*/ // the error "[trap] program called exit" should not be seen as an error, how to do that ?
}

// WasmCallHandler is the type of functions called back by wasm3 runtime
type WasmCallHandler func(wctx *WasmProcessContext, cs *CallSite) (uint32, error)

// BindAPIFunction binds a module+function name in wasm3 to a go routine
func (wctx *WasmProcessContext) BindAPIFunction(moduleName string, functionName string, signature string, handler WasmCallHandler) {
	if wctx.Fctx.Trace {
		fmt.Printf("binding function '%s'::'%s' signature:%s\n", moduleName, functionName, signature)
	}

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
	if wctx.Fctx.Trace {
		fmt.Printf("binding not_test_implemented stub function '%s'::'%s' signature:%s\n", module, name, signature)
	}

	wctx.Runtime.AttachFunction(module, name, signature, func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		fmt.Printf("called not yet implemented function '%s'... ABORTING WASM PROGRAM EXECUTION\n", name)
		return -2
	})
}
