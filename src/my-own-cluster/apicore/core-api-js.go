package apicore

    import (
        "my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )

func BindMyOwnClusterFunctionsJs(ctx enginejs.JSProcessContext) {
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            
            res, err := GetInputBufferID(ctx.Fctx)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getInputBufferId")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            
            res, err := GetOutputBufferID(ctx.Fctx)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getOutputBufferId")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            
            res, err := CreateExchangeBuffer(ctx.Fctx)
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

            res, err := WriteExchangeBuffer(ctx.Fctx, bufferId, content)
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

            res, err := WriteExchangeBufferHeader(ctx.Fctx, bufferId, name, value)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "writeExchangeBufferHeader")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            bufferId := int(c.GetNumber(-1))

            res, err := GetExchangeBufferSize(ctx.Fctx, bufferId)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getExchangeBufferSize")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            bufferId := int(c.GetNumber(-1))

            res, err := ReadExchangeBuffer(ctx.Fctx, bufferId)
            if err != nil {
                return 0
            }
            
            dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                    copy(dest, res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "readExchangeBuffer")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            bufferId := int(c.GetNumber(-1))

            res, err := ReadExchangeBufferHeaders(ctx.Fctx, bufferId)
            if err != nil {
                return 0
            }
            
            dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "readExchangeBufferHeaders")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            encoded := c.SafeToString(-1)

            res, err := Base64Decode(ctx.Fctx, encoded)
            if err != nil {
                return 0
            }
            
            dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "base64Decode")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            key := c.SafeToBytes(-2)
value := c.SafeToBytes(-1)

            res, err := PersistenceSet(ctx.Fctx, key, value)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "persistenceSet")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            url := c.SafeToString(-1)

            res, err := GetUrl(ctx.Fctx, url)
            if err != nil {
                return 0
            }
            
            dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getUrl")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            key := c.SafeToBytes(-1)

            res, err := PersistenceGet(ctx.Fctx, key)
            if err != nil {
                return 0
            }
            
            dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "persistenceGet")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            text := c.SafeToString(-1)

            res, err := PrintDebug(ctx.Fctx, text)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "printDebug")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            dest := c.SafeToBytes(-1)

            res, err := GetTime(ctx.Fctx, dest)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getTime")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            bufferId := int(c.GetNumber(-1))

            res, err := FreeBuffer(ctx.Fctx, bufferId)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "freeBuffer")
        }