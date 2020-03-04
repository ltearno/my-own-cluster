package duktape

import (
	"fmt"
	"my-own-cluster/common"
	"my-own-cluster/coreapi"

	"gopkg.in/ltearno/go-duktape.v3"
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

	// those wrappers are written by hand...
	BindMocFunctionsMano(fctx, ctx)

	// soon will be completely replaced by those ones, generated from 'my-own-cluster.api.json'
	BindMyOwnClusterFunctionsJs(fctx, ctx)

	// TODO : this should be factorized in the concept of pluggable "runtime providers"
	BindOpenGLFunctionsJs(fctx, ctx)

	ctx.PutPropString(-2, "moc")
	ctx.Pop()

	return &JSProcessContext{
		Fctx:    fctx,
		Context: ctx,
	}, nil
}

func SafeToBytes(c *duktape.Context, index int) []byte {
	var input []byte = nil
	switch c.GetType(index) {
	case duktape.TypeString:
		input = []byte(c.SafeToString(index))
		break
	case duktape.TypeBuffer:
		inputPtr, inputLength := c.GetBuffer(index)
		input = (*[1 << 30]byte)(inputPtr)[:inputLength:inputLength]
	case duktape.TypeObject:
		if c.IsBufferData(index) {
			inputPtr, inputLength := c.GetBufferData(index)
			input = (*[1 << 30]byte)(inputPtr)[:inputLength:inputLength]
		} else {
			fmt.Printf("cannot handle TypeObject content type of input param when SafeToBytes\n")
			return nil
		}
	case duktape.TypePointer:
		fmt.Printf("cannot handle TypePointer content type of input param when SafeToBytes\n")
		return nil
	default:
		fmt.Printf("cannot guess content type of input param when SafeToBytes\n")
		return nil
	}
	return input
}

func (jsctx *JSProcessContext) Run() error {
	jsctx.Context.PushString(string(jsctx.Fctx.CodeBytes))
	err := jsctx.Context.Peval()
	if err != nil {
		return fmt.Errorf("cannot eval script, probably syntax error")
	}

	jsctx.Context.Pop()
	jsctx.Context.PushGlobalObject()
	ok := jsctx.Context.GetPropString(-1, jsctx.Fctx.StartFunction)
	if !ok {
		return fmt.Errorf("cannot find start function %s", jsctx.Fctx.StartFunction)
	}

	res := jsctx.Context.Pcall(0)
	if res != 0 {
		return fmt.Errorf("error while executing js function '%s'", jsctx.Fctx.StartFunction)
	}

	jsctx.Fctx.Result = jsctx.Context.GetInt(-1)

	jsctx.Context.DestroyHeap()

	return nil
}

func BindMocFunctionsMano(fctx *common.FunctionExecutionContext, ctx *duktape.Context) {
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
		decoded := SafeToBytes(c, -1)
		encoded := coreapi.Base64Encode(fctx, decoded)

		c.PushString(encoded)

		return 1
	})
	ctx.PutPropString(-2, "base64Encode")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		contentBytesPtr, contentBytesLength := c.GetBuffer(-1)
		contentBytes := (*[1 << 30]byte)(contentBytesPtr)[:contentBytesLength:contentBytesLength]
		contentType := c.SafeToString(-2)
		name := c.SafeToString(-3)

		techID, err := coreapi.RegisterBlobWithName(fctx, name, contentType, contentBytes)
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

		techID, err := coreapi.RegisterBlob(fctx, contentType, contentBytes)
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

		err := coreapi.PlugFunction(fctx, method, path, name, startFunction)
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

		err := coreapi.PlugFile(fctx, method, path, name)
		if err != nil {
			fmt.Printf("[ERROR] plugFile failed\n")
			return 0
		}

		return 0
	})
	ctx.PutPropString(-2, "plugFile")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		c.PushInt(fctx.Orchestrator.CreateExchangeBuffer())

		return 1
	})
	ctx.PutPropString(-2, "createExchangeBuffer")

	ctx.PushGoFunction(func(c *duktape.Context) int {
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
}
