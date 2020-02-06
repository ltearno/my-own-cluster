package wasm

import (
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

	nextBufferID := 0
	buffers := make(map[int][]byte)

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
		bufferID := nextBufferID
		nextBufferID++
		buffers[bufferID] = buffer

		return uint32(bufferID), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_buffer_size", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		buffer, ok := buffers[int(bufferID)]
		if !ok {
			fmt.Printf("GET BUFFER SIZE FOR UNKNOWN BUFFER\n")
			return 0, nil
		}

		return uint32(len(buffer)), nil
	})

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "get_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		sourceBuffer, ok := buffers[int(bufferID)]
		if !ok {
			fmt.Printf("GET BUFFER ON UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

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

		delete(buffers, int(bufferID))

		return uint32(0), nil
	})
}
