package apigpu

import (
	"github.com/ltearno/my-own-cluster/enginejs"

	"gopkg.in/ltearno/go-duktape.v3"
)

func BindOpenGLFunctionsJs(ctx enginejs.JSProcessContext, cookie interface{}) {
	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		specification := c.SafeToString(-1)

		res, err := ComputeShader(ctx.Fctx, cookie, specification)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "computeShader")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		width := int(c.GetNumber(-4))
		height := int(c.GetNumber(-3))
		pixelsExchangeBufferId := int(c.GetNumber(-2))
		pngExchangeBufferId := int(c.GetNumber(-1))

		res, err := CreateImageFromRgbaFloatPixels(ctx.Fctx, cookie, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "createImageFromRgbaFloatPixels")

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		width := int(c.GetNumber(-4))
		height := int(c.GetNumber(-3))
		pixelsExchangeBufferId := int(c.GetNumber(-2))
		pngExchangeBufferId := int(c.GetNumber(-1))

		res, err := CreateImageFromRFloatPixels(ctx.Fctx, cookie, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
		if err != nil {
			return 0
		}

		c.PushInt(res)

		return 1
	})
	ctx.Context.PutPropString(-2, "createImageFromRFloatPixels")
}
