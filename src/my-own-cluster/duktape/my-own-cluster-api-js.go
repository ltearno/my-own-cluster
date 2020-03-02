package duktape

import (
	"my-own-cluster/common"
	"my-own-cluster/coreapi"

	"gopkg.in/ltearno/go-duktape.v3"
)

func BindMyOwnClusterFunctionsJs(fctx *common.FunctionExecutionContext, ctx *duktape.Context) {
	ctx.PushGoFunction(func(c *duktape.Context) int {

		res, err := coreapi.GetInputBufferID(fctx)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.PutPropString(-2, "getInputBufferId")

	ctx.PushGoFunction(func(c *duktape.Context) int {

		res, err := coreapi.GetOutputBufferID(fctx)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.PutPropString(-2, "getOutputBufferId")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-2))
		content := SafeToBytes(c, -1)

		res, err := coreapi.WriteExchangeBuffer(fctx, bufferId, content)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.PutPropString(-2, "writeExchangeBuffer")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		bufferId := int(c.GetNumber(-1))

		res, err := coreapi.ReadExchangeBuffer(fctx, bufferId)
		if err != nil {
			return 0
		}

		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.PutPropString(-2, "readExchangeBuffer")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		encoded := c.SafeToString(-1)

		res, err := coreapi.Base64Decode(fctx, encoded)
		if err != nil {
			return 0
		}

		dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
		copy(dest, res)

		return 1
	})
	ctx.PutPropString(-2, "base64Decode")

	ctx.PushGoFunction(func(c *duktape.Context) int {
		key := SafeToBytes(c, -2)
		value := SafeToBytes(c, -1)

		res, err := coreapi.PersistenceSet(fctx, key, value)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.PutPropString(-2, "persistenceSet")
}
