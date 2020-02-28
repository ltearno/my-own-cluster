package duktape

import (
	"encoding/base64"
	"fmt"
	"my-own-cluster/common"

	"gopkg.in/olebedev/go-duktape.v3"
)

type JSProcessContext struct {
	Fctx *common.FunctionExecutionContext

	Context *duktape.Context
}

type JavascriptDuktapeEngine struct {
}

func NewJavascriptDuktapeEngine() *JavascriptDuktapeEngine {
	return &JavascriptDuktapeEngine{}
}

func (e *JavascriptDuktapeEngine) PrepareContext(fctx *common.FunctionExecutionContext) (common.ExecutionEngineContext, error) {
	ctx := duktape.New()

	ctx.PushGlobalObject()
	ctx.PushObject()

	ctx.PushGoFunction(func(c *duktape.Context) int {
		res, err := common.GetInputBufferID(fctx)
		if err != nil {
			c.PushInt(-1)
		} else {
			c.PushInt(res)
		}

		return 1
	})
	ctx.PutPropString(-2, "getInputBufferId")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		res, err := common.GetOutputBufferID(fctx)
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
		decoded, err := common.Base64Decode(fctx, encoded)
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

		techID, err := common.RegisterBlobWithName(fctx, name, contentType, contentBytes)
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

		techID, err := common.RegisterBlob(fctx, contentType, contentBytes)
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

		err := common.PlugFunction(fctx, method, path, name, startFunction)
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

		err := common.PlugFile(fctx, method, path, name)
		if err != nil {
			fmt.Printf("[ERROR] plugFile failed\n")
			return 0
		}

		return 0
	})
	ctx.PutPropString(-2, "plugFile")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-7)
		startFunction := c.SafeToString(-6)
		arguments := []int{}
		c.Enum(-5, (1 << 5))
		for c.Next(-1, true) {
			arguments = append(arguments, c.GetInt(-1))
			c.Pop()
			c.Pop()
		}
		c.Pop()
		mode := c.SafeToString(-4)
		input := []byte(c.SafeToString(-3))
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

		inputExchangeBufferID := fctx.Orchestrator.CreateExchangeBuffer()
		inputExchangeBuffer := fctx.Orchestrator.GetExchangeBuffer(inputExchangeBufferID)
		inputExchangeBuffer.Write(input)
		outputExchangeBufferID := fctx.Orchestrator.CreateExchangeBuffer()

		newCtx := fctx.Orchestrator.NewFunctionExecutionContext(
			name,
			startFunction,
			arguments,
			fctx.Trace,
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

		outputExchangeBuffer := fctx.Orchestrator.GetExchangeBuffer(outputExchangeBufferID)

		// push a json with status result and output
		c.PushObject()
		c.PushBoolean(true)
		c.PutPropString(-2, "status")
		c.PushInt(newCtx.Result)
		c.PutPropString(-2, "result")
		c.PushString(base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(outputExchangeBuffer.GetBuffer()))
		c.PutPropString(-2, "output")

		// release exchange buffers
		fctx.Orchestrator.ReleaseExchangeBuffer(inputExchangeBufferID)
		fctx.Orchestrator.ReleaseExchangeBuffer(outputExchangeBufferID)

		return 1
	})
	ctx.PutPropString(-2, "callFunction")

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
