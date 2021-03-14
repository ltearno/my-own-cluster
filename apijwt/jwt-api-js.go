package apijwt

    import (
        "github.com/ltearno/my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )

func BindJwtFunctionsJs(ctx enginejs.JSProcessContext, cookie interface{}) {
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            jwt := c.SafeToString(-1)

            res, err := VerifyJwt(ctx.Fctx, cookie, jwt)
            if err != nil {
                return 0
            }
            
            c.PushString(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "verifyJwt")
        }