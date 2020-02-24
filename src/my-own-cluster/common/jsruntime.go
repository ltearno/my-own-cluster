package common

import (
	"fmt"

	"gopkg.in/olebedev/go-duktape.v3"
)

type JSProcessContext struct {
	Fctx *FunctionExecutionContext

	Context *duktape.Context
}

func PorcelainPrepareJs(fctx *FunctionExecutionContext) (*JSProcessContext, error) {
	ctx := duktape.New()

	ctx.PushGlobalObject()
	ctx.PushObject()

	ctx.PushGoFunction(func(c *duktape.Context) int {
		res, err := GetInputBufferID(fctx)
		if err != nil {
			c.PushInt(-1)
		} else {
			c.PushInt(res)
		}

		return 1
	})
	ctx.PutPropString(-2, "getInputBufferId")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		res, err := GetOutputBufferID(fctx)
		if err != nil {
			c.PushInt(-1)
		} else {
			c.PushInt(res)
		}

		return 1
	})
	ctx.PutPropString(-2, "getOutputBufferId")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		bufferID := int(c.GetNumber(-1))

		buffer := fctx.Orchestrator.GetExchangeBuffer(bufferID)
		if buffer == nil {
			fmt.Printf("buffer %d not found for reading\n", bufferID)
			return 0
		}

		c.PushString(string(buffer.GetBuffer()))

		return 1
	})
	ctx.PutPropString(-2, "readExchangeBufferAsString")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		bufferID := int(c.GetNumber(-2))

		var contentBytes []byte
		switch c.GetType(-1) {
		case duktape.TypeString:
			contentBytes = []byte(c.SafeToString(-1))
			break
		default:
			fmt.Printf("cannot guess content type when writing on buffer %d\n", bufferID)
			return 0
		}

		buffer := fctx.Orchestrator.GetExchangeBuffer(bufferID)
		if buffer == nil {
			fmt.Printf("buffer %d not found for writing\n", bufferID)
			return 0
		}

		buffer.Write(contentBytes)

		c.PushInt(len(contentBytes))
		return 1
	})
	ctx.PutPropString(-2, "writeExchangeBufferFromString")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		bufferID := int(c.GetNumber(-3))
		name := c.SafeToString(-2)
		value := c.SafeToString(-1)

		buffer := fctx.Orchestrator.GetExchangeBuffer(bufferID)
		if buffer == nil {
			fmt.Printf("buffer %d not found for writing header %s\n", bufferID, name)
			return 0
		}

		buffer.SetHeader(name, value)

		return 0
	})
	ctx.PutPropString(-2, "writeExchangeBufferHeader")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		encoded := c.SafeToString(-1)
		decoded, err := Base64Decode(fctx, encoded)
		if err != nil {
			fmt.Printf("cannot decode base64\n")
			return 0
		}

		dest := (*[1 << 30]byte)(c.PushBuffer(len(decoded), false))[:len(decoded):len(decoded)]

		copy(dest, decoded)

		return 1
	})
	ctx.PutPropString(-2, "base64Decode")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		contentBytesPtr, contentBytesLength := c.GetBuffer(-1)
		contentBytes := (*[1 << 30]byte)(contentBytesPtr)[:contentBytesLength:contentBytesLength]
		contentType := c.SafeToString(-2)
		name := c.SafeToString(-3)

		techID, err := RegisterBlobWithName(fctx, name, contentType, contentBytes)
		if err != nil {
			fmt.Printf("[ERROR] registerBlobWithName failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.PutPropString(-2, "registerBlobWithName")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		contentBytesPtr, contentBytesLength := c.GetBuffer(-1)
		contentBytes := (*[1 << 30]byte)(contentBytesPtr)[:contentBytesLength:contentBytesLength]
		contentType := c.SafeToString(-2)

		techID, err := RegisterBlob(fctx, contentType, contentBytes)
		if err != nil {
			fmt.Printf("[ERROR] registerBlob failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.PutPropString(-2, "registerBlob")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)

		techID, err := fctx.Orchestrator.GetBlobTechIDFromName(name)
		if err != nil {
			fmt.Printf("[ERROR] getBlobTechIDFromName failed\n")
			return 0
		}

		c.PushString(techID)
		return 1
	})
	ctx.PutPropString(-2, "getBlobTechIDFromName")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		techID := c.SafeToString(-1)

		contentBytes, err := fctx.Orchestrator.GetBlobBytesByTechID(techID)
		if err != nil {
			fmt.Printf("[ERROR] getBlobTechIDFromName failed\n")
			return 0
		}

		c.PushString(string(contentBytes))
		return 1
	})
	ctx.PutPropString(-2, "getBlobBytesAsString")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		startFunction := c.SafeToString(-1)
		name := c.SafeToString(-2)
		path := c.SafeToString(-3)
		method := c.SafeToString(-4)

		err := PlugFunction(fctx, method, path, name, startFunction)
		if err != nil {
			fmt.Printf("[ERROR] plugFunction failed\n")
			return 0
		}

		return 0
	})
	ctx.PutPropString(-2, "plugFunction")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)
		path := c.SafeToString(-2)
		method := c.SafeToString(-3)

		err := PlugFile(fctx, method, path, name)
		if err != nil {
			fmt.Printf("[ERROR] plugFile failed\n")
			return 0
		}

		return 0
	})
	ctx.PutPropString(-2, "plugFile")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		c.PushString(fctx.Orchestrator.GetStatus())
		return 1
	})
	ctx.PutPropString(-2, "getStatus")

	ctx.PutPropString(-2, "moc")
	ctx.Pop()

	return &JSProcessContext{
		Fctx:    fctx,
		Context: ctx,
	}, nil
}

func (jsctx *JSProcessContext) Run() error {
	jsctx.Context.PushString(string(jsctx.Fctx.CodeBytes))
	jsctx.Context.Eval()
	jsctx.Context.Pop()
	jsctx.Context.PushGlobalObject()
	ok := jsctx.Context.GetPropString(-1, jsctx.Fctx.StartFunction)
	if !ok {
		return fmt.Errorf("cannot find start function %s", jsctx.Fctx.StartFunction)
	}

	jsctx.Context.Call(0)

	jsctx.Fctx.Result = jsctx.Context.GetInt(-1)

	jsctx.Context.DestroyHeap()

	return nil
}
