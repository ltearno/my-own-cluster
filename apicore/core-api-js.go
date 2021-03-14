package apicore

import (
	"github.com/ltearno/my-own-cluster/enginejs"

	"gopkg.in/ltearno/go-duktape.v3"
)

func BindMyOwnClusterFunctionsJs(ctx enginejs.JSProcessContext, cookie interface{}) {
	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := GetInputBufferID(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getInputBufferId")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := GetOutputBufferID(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getOutputBufferId")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := CreateExchangeBuffer(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "createExchangeBuffer")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-2))
		content := c.SafeToBytes(-1)

		res, err := WriteExchangeBuffer(ctx.Fctx, cookie, bufferId, content)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "writeExchangeBuffer")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-3))
		name := c.SafeToString(-2)
		value := c.SafeToString(-1)

		res, err := WriteExchangeBufferHeader(ctx.Fctx, cookie, bufferId, name, value)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "writeExchangeBufferHeader")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-2))
		statusCode := int(c.GetNumber(-1))

		res, err := WriteExchangeBufferStatusCode(ctx.Fctx, cookie, bufferId, statusCode)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "writeExchangeBufferStatusCode")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-1))

		res, err := ReadExchangeBuffer(ctx.Fctx, cookie, bufferId)
		if err != nil {
			return 0
		}

		if res == nil {
			return 0
		}
		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.Context.PutPropString(-2, "readExchangeBuffer")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-1))

		res, err := ReadExchangeBufferHeaders(ctx.Fctx, cookie, bufferId)
		if err != nil {
			return 0
		}

		c.PushObject()
		for k, v := range res {
			c.PushString(v)
			c.PutPropString(-2, k)
		}

		return 1
	})
	ctx.Context.PutPropString(-2, "readExchangeBufferHeaders")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		encoded := c.SafeToString(-1)

		res, err := Base64Decode(ctx.Fctx, cookie, encoded)
		if err != nil {
			return 0
		}

		if res == nil {
			return 0
		}
		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.Context.PutPropString(-2, "base64Decode")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		input := c.SafeToBytes(-1)

		res, err := Base64Encode(ctx.Fctx, cookie, input)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "base64Encode")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-3)
		contentType := c.SafeToString(-2)
		content := c.SafeToBytes(-1)

		res, err := RegisterBlobWithName(ctx.Fctx, cookie, name, contentType, content)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "registerBlobWithName")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		contentType := c.SafeToString(-2)
		content := c.SafeToBytes(-1)

		res, err := RegisterBlob(ctx.Fctx, cookie, contentType, content)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "registerBlob")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)

		res, err := GetBlobTechIdFromName(ctx.Fctx, cookie, name)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getBlobTechIdFromName")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-1)

		res, err := GetBlobBytesAsString(ctx.Fctx, cookie, name)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getBlobBytesAsString")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		method := c.SafeToString(-6)
		path := c.SafeToString(-5)
		name := c.SafeToString(-4)
		startFunction := c.SafeToString(-3)
		data := c.SafeToString(-2)
		tagsJson := c.SafeToString(-1)

		res, err := PlugFunction(ctx.Fctx, cookie, method, path, name, startFunction, data, tagsJson)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "plugFunction")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		method := c.SafeToString(-4)
		path := c.SafeToString(-3)
		name := c.SafeToString(-2)
		tagsJson := c.SafeToString(-1)

		res, err := PlugFile(ctx.Fctx, cookie, method, path, name, tagsJson)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "plugFile")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		method := c.SafeToString(-2)
		path := c.SafeToString(-1)

		res, err := UnplugPath(ctx.Fctx, cookie, method, path)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "unplugPath")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := GetStatus(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getStatus")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		key := c.SafeToBytes(-2)
		value := c.SafeToBytes(-1)

		res, err := PersistenceSet(ctx.Fctx, cookie, key, value)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "persistenceSet")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		url := c.SafeToString(-1)

		res, err := GetUrl(ctx.Fctx, cookie, url)
		if err != nil {
			return 0
		}

		if res == nil {
			return 0
		}
		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getUrl")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		key := c.SafeToBytes(-1)

		res, err := PersistenceGet(ctx.Fctx, cookie, key)
		if err != nil {
			return 0
		}

		if res == nil {
			return 0
		}
		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.Context.PutPropString(-2, "persistenceGet")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		prefix := c.SafeToString(-1)

		res, err := PersistenceGetSubset(ctx.Fctx, cookie, prefix)
		if err != nil {
			return 0
		}

		c.PushObject()
		for k, v := range res {
			c.PushString(v)
			c.PutPropString(-2, k)
		}

		return 1
	})
	ctx.Context.PutPropString(-2, "persistenceGetSubset")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		text := c.SafeToString(-1)

		res, err := PrintDebug(ctx.Fctx, cookie, text)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "printDebug")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		dest := c.SafeToBytes(-1)

		res, err := GetTime(ctx.Fctx, cookie, dest)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "getTime")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-1))

		res, err := FreeBuffer(ctx.Fctx, cookie, bufferId)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "freeBuffer")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-8)
		startFunction := c.SafeToString(-7)
		arguments := []int{}
		// #define DUK_ENUM_ARRAY_INDICES_ONLY       (1U << 5)    // only enumerate array indices
		c.Enum(-6, (1 << 5))
		for c.Next(-1, true) {
			arguments = append(arguments, c.GetInt(-1))
			c.Pop()
			c.Pop()
		}
		c.Pop()
		mode := c.SafeToString(-5)
		inputExchangeBufferId := int(c.GetNumber(-4))
		outputExchangeBufferId := int(c.GetNumber(-3))
		posixFileName := c.SafeToString(-2)
		posixArguments := []string{}
		// #define DUK_ENUM_ARRAY_INDICES_ONLY       (1U << 5)    // only enumerate array indices
		c.Enum(-1, (1 << 5))
		for c.Next(-1, true) {
			posixArguments = append(posixArguments, c.SafeToString(-1))
			c.Pop()
			c.Pop()
		}
		c.Pop()

		res, err := CallFunction(ctx.Fctx, cookie, name, startFunction, arguments, mode, inputExchangeBufferId, outputExchangeBufferId, posixFileName, posixArguments)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "callFunction")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := ExportDatabase(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		if res == nil {
			return 0
		}
		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.Context.PutPropString(-2, "exportDatabase")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		proxySpecJson := c.SafeToString(-1)

		res, err := BetaWebProxy(ctx.Fctx, cookie, proxySpecJson)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "betaWebProxy")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {

		res, err := IsTrace(ctx.Fctx, cookie)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "isTrace")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		name := c.SafeToString(-3)
		startFunction := c.SafeToString(-2)
		data := c.SafeToString(-1)

		res, err := PlugFilter(ctx.Fctx, cookie, name, startFunction, data)
		if err != nil {
			return 0
		}

		c.PushString(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "plugFilter")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		id := c.SafeToString(-1)

		res, err := UnplugFilter(ctx.Fctx, cookie, id)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "unplugFilter")
}
