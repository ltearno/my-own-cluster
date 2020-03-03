package duktape

    import (
        "my-own-cluster/common"
        "my-own-cluster/openglapi"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )

func BindOpenGLFunctionsJs(fctx *common.FunctionExecutionContext, ctx *duktape.Context) {
        ctx.PushGoFunction(func(c *duktape.Context) int {
            specification := c.SafeToString(-1)

            res, err := openglapi.ComputeShader(fctx, specification)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.PutPropString(-2, "computeShader")
        
        ctx.PushGoFunction(func(c *duktape.Context) int {
            width := int(c.GetNumber(-4))
height := int(c.GetNumber(-3))
pixelsExchangeBufferId := int(c.GetNumber(-2))
pngExchangeBufferId := int(c.GetNumber(-1))

            res, err := openglapi.CreateImageFromRgbafloatPixels(fctx, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.PutPropString(-2, "createImageFromRgbafloatPixels")
        }