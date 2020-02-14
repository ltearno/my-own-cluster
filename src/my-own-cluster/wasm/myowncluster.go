package wasm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
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

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "print_debug", "i(ii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		buffer := cs.GetParamByteBuffer(0, 1)

		fmt.Printf("\n[my-own-cluster api, ctx %s, print_debug]: %s\n", wctx.Name, string(buffer))

		return uint32(len(buffer)), nil
	})

	// params : time addr (should be wide enough for the 64 bit timestamp)
	wctx.BindAPIFunction("my-own-cluster", "get_time", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		timestampPtr := cs.GetParamPointer(0)

		*(*int64)(timestampPtr) = time.Now().UnixNano()

		return uint32(0), nil
	})

	// params : buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "register_buffer", "i(ii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		buffer := cs.GetParamByteBuffer(0, 1)
		bufferID := wctx.Orchestrator.CreateExchangeBuffer()
		exchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(bufferID)
		if exchangeBuffer == nil {
			return 0, errors.New("cannot get just created exchange buffer")
		}

		exchangeBuffer.Write(buffer)

		return uint32(bufferID), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_buffer_size", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		exchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(int(bufferID))
		if exchangeBuffer == nil {
			fmt.Printf("GET EXCHANGE BUFFER SIZE FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		buffer := exchangeBuffer.GetBuffer()

		return uint32(len(buffer)), nil
	})

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "get_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		exchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(int(bufferID))
		if exchangeBuffer == nil {
			fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		sourceBuffer := exchangeBuffer.GetBuffer()

		if len(buffer) != len(sourceBuffer) {
			fmt.Printf("GET BUFFER WITH WRONG SIZE given %d, required %d\n", len(buffer), len(sourceBuffer))
			return 0, nil
		}

		copy(buffer, sourceBuffer)

		return uint32(len(sourceBuffer)), nil
	})

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "write_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		exchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(int(bufferID))
		if exchangeBuffer == nil {
			fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		exchangeBuffer.Write(buffer)

		return uint32(len(buffer)), nil
	})

	// params : buffer id, buffer addr, buffer length
	wctx.BindAPIFunction("my-own-cluster", "write_buffer_header", "i(iiiii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)
		name := cs.GetParamByteBuffer(1, 2)
		value := cs.GetParamByteBuffer(3, 4)

		exchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(int(bufferID))
		if exchangeBuffer == nil {
			fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
			return 0, nil
		}

		exchangeBuffer.SetHeader(string(name), string(value))

		return uint32(1), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "free_buffer", "i(i)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		bufferID := cs.GetParamUINT32(0)

		wctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))

		return uint32(0), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_input_buffer_id", "i()", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		return uint32(wctx.InputExchangeBufferID), nil
	})

	// params : buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_output_buffer_id", "i()", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		return uint32(wctx.OutputExchangeBufferID), nil
	})

	// params : url addr, url length
	// returns: content buffer id
	wctx.BindAPIFunction("my-own-cluster", "get_url", "i(ii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		url := cs.GetParamString(0, 1)

		resp, err := http.Get(url)
		if err != nil {
			return uint32(0xffff), nil
		}
		defer resp.Body.Close()

		bytes, _ := ioutil.ReadAll(resp.Body)

		contentBufferID := wctx.Orchestrator.CreateExchangeBuffer()
		contentBuffer := wctx.Orchestrator.GetExchangeBuffer(contentBufferID)

		contentBuffer.Write(bytes)

		return uint32(contentBufferID), nil
	})
}
