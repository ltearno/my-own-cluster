package apijwt

    import (
        "my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )

func BindJwtFunctionsJs(ctx enginejs.JSProcessContext) {
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            jwt := c.SafeToString(-1)

            res, err := VerifyJwt(ctx.Fctx, jwt)
            if err != nil {
                return 0
            }
            
            c.PushString(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "verifyJwt")
        }