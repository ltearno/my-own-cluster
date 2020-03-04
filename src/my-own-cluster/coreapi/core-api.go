package coreapi

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"net/http"
	"time"

	"gopkg.in/ltearno/go-duktape.v3"
)

type CoreAPIProvider struct{}

func NewCoreAPIProvider() (common.APIProvider, error) {
	return &CoreAPIProvider{}, nil
}

func (p *CoreAPIProvider) BindToExecutionEngineContext(ctx common.ExecutionEngineContextBounding) {
	wctx, ok := ctx.(*enginewasm.WasmProcessContext)
	if ok {
		wctx := *wctx
		BindMyOwnClusterFunctionsWASM(wctx)

		// params : key addr, key length
		wctx.BindAPIFunction("my-own-cluster", "persistence_get", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			key := cs.GetParamByteBuffer(0, 1)

			value, present := wctx.Fctx.Orchestrator.PersistenceGet(key)
			if !present {
				return uint32(0xffff), nil
			}

			bufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
			buffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(bufferID)
			buffer.Write(value)

			return uint32(bufferID), nil
		})

		// params : prefix addr, prefix length
		wctx.BindAPIFunction("my-own-cluster", "persistence_get_subset", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			prefix := cs.GetParamByteBuffer(0, 1)

			subset, err := wctx.Fctx.Orchestrator.PersistenceGetSubset(prefix)
			if err != nil {
				return uint32(0xffff), nil
			}

			var b bytes.Buffer
			bs := make([]byte, 4)

			binary.LittleEndian.PutUint32(bs, uint32(len(subset)))
			b.Write(bs)

			for i := 0; i < len(subset); i++ {
				binary.LittleEndian.PutUint32(bs, uint32(len(subset[i])))
				b.Write(bs)

				b.Write(subset[i])
			}

			bufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
			buffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(bufferID)
			buffer.Write(b.Bytes())

			return uint32(bufferID), nil
		})

		// pub fn get_buffer_headers(buffer_id: u32) -> u32;
		wctx.BindAPIFunction("my-own-cluster", "get_buffer_headers", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			exchangeBufferID := cs.GetParamUINT32(0)

			exchangeBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(int(exchangeBufferID))
			if exchangeBuffer == nil {
				return uint32(0xffff), nil
			}

			var b bytes.Buffer
			bs := make([]byte, 4)

			binary.LittleEndian.PutUint32(bs, uint32(2*exchangeBuffer.GetHeadersCount()))
			b.Write(bs)

			exchangeBuffer.GetHeaders(func(name string, value string) {
				binary.LittleEndian.PutUint32(bs, uint32(len([]byte(name))))
				b.Write(bs)
				b.Write([]byte(name))

				binary.LittleEndian.PutUint32(bs, uint32(len([]byte(value))))
				b.Write(bs)
				b.Write([]byte(value))
			})

			bufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
			buffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(bufferID)
			buffer.Write(b.Bytes())

			return uint32(bufferID), nil
		})

		// params : buffer id, buffer addr, buffer length
		wctx.BindAPIFunction("my-own-cluster", "print_debug", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			buffer := cs.GetParamByteBuffer(0, 1)

			fmt.Printf("\n[my-own-cluster api, ctx %s, print_debug]: %s\n", wctx.Fctx.Name, string(buffer))

			return uint32(len(buffer)), nil
		})

		// params : time addr (should be wide enough for the 64 bit timestamp)
		wctx.BindAPIFunction("my-own-cluster", "get_time", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			timestampPtr := cs.GetParamPointer(0)

			*(*int64)(timestampPtr) = time.Now().UnixNano()

			return uint32(0), nil
		})

		// params : buffer addr, buffer length
		wctx.BindAPIFunction("my-own-cluster", "register_buffer", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			buffer := cs.GetParamByteBuffer(0, 1)
			bufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
			exchangeBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(bufferID)
			if exchangeBuffer == nil {
				return 0, errors.New("cannot get just created exchange buffer")
			}

			exchangeBuffer.Write(buffer)

			return uint32(bufferID), nil
		})

		// params : buffer id
		wctx.BindAPIFunction("my-own-cluster", "free_buffer", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			bufferID := cs.GetParamUINT32(0)
			wctx.Fctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
			return uint32(0), nil
		})

		// params :
		// params : url addr, url length
		// returns: content buffer id
		wctx.BindAPIFunction("my-own-cluster", "get_url", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
			url := cs.GetParamString(0, 1)

			resp, err := http.Get(url)
			if err != nil {
				return uint32(0xffff), nil
			}
			defer resp.Body.Close()

			bytes, _ := ioutil.ReadAll(resp.Body)

			contentBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
			contentBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(contentBufferID)

			contentBuffer.Write(bytes)

			return uint32(contentBufferID), nil
		})
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		// those wrappers are written by hand...
		BindMocFunctionsMano(*jsctx)

		// soon will be completely replaced by those ones, generated from 'my-own-cluster.api.json'
		BindMyOwnClusterFunctionsJs(*jsctx)
	}
}

/**
 * Implementation of core functions that execution engines can use to exposes functionality to their runtimes
 */

func GetInputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.InputExchangeBufferID, nil
}

func GetOutputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.OutputExchangeBufferID, nil
}

func FreeBuffer(ctx *common.FunctionExecutionContext, bufferID int) error {
	ctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
	return nil
}

func PlugFunction(ctx *common.FunctionExecutionContext, method string, path string, name string, startFunction string) error {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction)
	return nil
}

func PlugFile(ctx *common.FunctionExecutionContext, method string, path string, name string) error {
	ctx.Orchestrator.PlugFile(method, path, name)
	return nil
}

func RegisterBlobWithName(ctx *common.FunctionExecutionContext, name string, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlobWithName(name, contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func RegisterBlob(ctx *common.FunctionExecutionContext, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlob(contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func Base64Decode(ctx *common.FunctionExecutionContext, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func Base64Encode(ctx *common.FunctionExecutionContext, b []byte) string {
	return base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(b)
}

func WriteExchangeBuffer(ctx *common.FunctionExecutionContext, bufferID int, content []byte) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.Write(content)

	return len(content), nil
}

func WriteExchangeBufferHeader(ctx *common.FunctionExecutionContext, bufferID int, name string, value string) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(int(bufferID))
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.SetHeader(string(name), string(value))

	return 0, nil
}

func ReadExchangeBuffer(ctx *common.FunctionExecutionContext, bufferID int) ([]byte, error) {
	buffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if buffer == nil {
		fmt.Printf("buffer %d not found for reading\n", bufferID)
		return nil, nil
	}

	bufferBytes := buffer.GetBuffer()

	return bufferBytes, nil
}

func PersistenceSet(ctx *common.FunctionExecutionContext, key []byte, value []byte) (int, error) {
	ok := ctx.Orchestrator.PersistenceSet(key, value)
	if !ok {
		return 0, errors.New("cannot persist key !")
	}

	return 0, nil
}

func CreateExchangeBuffer(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.Orchestrator.CreateExchangeBuffer(), nil
}

func ReadExchangeBufferHeaders(ctx *common.FunctionExecutionContext, bufferID int) ([]byte, error) {
	buffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	headers := make(map[string]string)
	buffer.GetHeaders(func(name string, value string) { headers[name] = value })
	b, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func BindMocFunctionsMano(ctx enginejs.JSProcessContext) {
	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferID := int(c.GetNumber(-3))
		name := c.SafeToString(-2)
		value := c.SafeToString(-1)

		buffer := ctx.Fctx.Orchestrator.GetExchangeBuffer(bufferID)
		if buffer == nil {
			fmt.Printf("buffer %d not found for writing header %s\n", bufferID, name)
			return 0
		}

		buffer.SetHeader(name, value)

		return 0
	})
	ctx.Context.PutPropString(-2, "writeExchangeBufferHeader")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		decoded := c.SafeToBytes(-1)
		encoded := Base64Encode(ctx.Fctx, decoded)

		c.PushString(encoded)

		return 1
	})
	ctx.Context.PutPropString(-2, "base64Encode")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		contentBytesPtr, contentBytesLength := c.GetBuffer(-1)
		contentBytes := (*[1 << 30]byte)(contentBytesPtr)[:contentBytesLength:contentBytesLength]
		contentType := c.SafeToString(-2)
		name := c.SafeToString(-3)

		techID, err := RegisterBlobWithName(ctx.Fctx, name, contentType, contentBytes)
		if err != nil {
			fmt.Printf("[ERROR] registerBlobWithName failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.Context.PutPropString(-2, "registerBlobWithName")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		contentBytesPtr, contentBytesLength := c.GetBuffer(-1)
		contentBytes := (*[1 << 30]byte)(contentBytesPtr)[:contentBytesLength:contentBytesLength]
		contentType := c.SafeToString(-2)

		techID, err := RegisterBlob(ctx.Fctx, contentType, contentBytes)
		if err != nil {
			fmt.Printf("[ERROR] registerBlob failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.Context.PutPropString(-2, "registerBlob")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)

		techID, err := ctx.Fctx.Orchestrator.GetBlobTechIDFromName(name)
		if err != nil {
			fmt.Printf("[ERROR] getBlobTechIDFromName failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.Context.PutPropString(-2, "getBlobTechIDFromName")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		techID := c.SafeToString(-1)

		contentBytes, err := ctx.Fctx.Orchestrator.GetBlobBytesByTechID(techID)
		if err != nil {
			fmt.Printf("[ERROR] getBlobTechIDFromName failed\n")
			return 0
		}

		c.PushString(string(contentBytes))
		return 1
	})
	ctx.Context.PutPropString(-2, "getBlobBytesAsString")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		startFunction := c.SafeToString(-1)
		name := c.SafeToString(-2)
		path := c.SafeToString(-3)
		method := c.SafeToString(-4)

		err := PlugFunction(ctx.Fctx, method, path, name, startFunction)
		if err != nil {
			fmt.Printf("[ERROR] plugFunction failed\n")
			return 0
		}

		return 0
	})
	ctx.Context.PutPropString(-2, "plugFunction")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)
		path := c.SafeToString(-2)
		method := c.SafeToString(-3)

		err := PlugFile(ctx.Fctx, method, path, name)
		if err != nil {
			fmt.Printf("[ERROR] plugFile failed\n")
			return 0
		}

		return 0
	})
	ctx.Context.PutPropString(-2, "plugFile")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		c.PushInt(ctx.Fctx.Orchestrator.CreateExchangeBuffer())

		return 1
	})
	ctx.Context.PutPropString(-2, "createExchangeBuffer")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-8)
		startFunction := c.SafeToString(-7)
		arguments := []int{}
		c.Enum(-6, (1 << 5))
		for c.Next(-1, true) {
			arguments = append(arguments, c.GetInt(-1))
			c.Pop()
			c.Pop()
		}
		c.Pop()
		mode := c.SafeToString(-5)
		inputExchangeBufferID := c.GetInt(-4)
		outputExchangeBufferID := c.GetInt(-3)
		posixFileName := c.SafeToString(-2)
		// #define DUK_ENUM_ARRAY_INDICES_ONLY       (1U << 5)    /* only enumerate array indices */
		posixArguments := []string{}
		c.Enum(-1, (1 << 5))
		for c.Next(-1, true) {
			posixArguments = append(posixArguments, c.SafeToString(-1))
			c.Pop()
			c.Pop()
		}
		c.Pop()

		newCtx := ctx.Fctx.Orchestrator.NewFunctionExecutionContext(
			name,
			startFunction,
			arguments,
			ctx.Fctx.Trace,
			mode,
			&posixFileName,
			&posixArguments,
			inputExchangeBufferID,
			outputExchangeBufferID,
		)

		err := newCtx.Run()
		if err != nil {
			fmt.Printf("[ERROR] callFunction failed (%v)\n", err)
			return 0
		}

		outputExchangeBuffer := ctx.Fctx.Orchestrator.GetExchangeBuffer(outputExchangeBufferID)
		outputBytes := outputExchangeBuffer.GetBuffer()

		// push a json with status result and output
		c.PushObject()
		c.PushBoolean(true)
		c.PutPropString(-2, "status")
		c.PushInt(newCtx.Result)
		c.PutPropString(-2, "result")
		dest := (*[1 << 30]byte)(c.PushBuffer(len(outputBytes), false))[:len(outputBytes):len(outputBytes)]
		copy(dest, outputBytes)
		c.PutPropString(-2, "output")

		// release exchange buffers
		ctx.Fctx.Orchestrator.ReleaseExchangeBuffer(inputExchangeBufferID)
		ctx.Fctx.Orchestrator.ReleaseExchangeBuffer(outputExchangeBufferID)

		return 1
	})
	ctx.Context.PutPropString(-2, "callFunction")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		c.PushString(ctx.Fctx.Orchestrator.GetStatus())
		return 1
	})
	ctx.Context.PutPropString(-2, "getStatus")
}
