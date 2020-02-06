package wasm

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/ltearno/go-wasm3"
)

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

	// params : buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "register_buffer", "i(ii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		buffer := cs.GetParamByteBuffer(0, 1)
		bufferID := wctx.Orchestrator.CreateOutputPort()
		outputPort := wctx.Orchestrator.GetOutputPort(bufferID)
		if outputPort == nil {
			return 0, errors.New("cannot get just created output port")
		}

		outputPort.Write(buffer)

		return uint32(bufferID), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_buffer_size", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		outputPort := wctx.Orchestrator.GetOutputPort(int(bufferID))
		if outputPort == nil {
			fmt.Printf("GET BUFFER SIZE FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		buffer := outputPort.GetBuffer()

		return uint32(len(buffer)), nil
	})

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "get_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		outputPort := wctx.Orchestrator.GetOutputPort(int(bufferID))
		if outputPort == nil {
			fmt.Printf("GET BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		sourceBuffer := outputPort.GetBuffer()

		if len(buffer) != len(sourceBuffer) {
			fmt.Printf("GET BUFFER WITH WRONG SIZE given %d, required %d\n", len(buffer), len(sourceBuffer))
			return 0, nil
		}

		copy(buffer, sourceBuffer)

		return uint32(len(sourceBuffer)), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "free_buffer", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)

		wctx.Orchestrator.ReleaseOutputPort(int(bufferID))

		return uint32(0), nil
	})
}
