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
            bufferId := int(c.GetNumber(-1))

            res, err := GetBufferHeaders(ctx.Fctx, bufferId)
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
        ctx.Context.PutPropString(-2, "getBufferHeaders")
        
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
            input := c.SafeToBytes(-1)

            res, err := Base64Encode(ctx.Fctx, input)
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

            res, err := RegisterBlobWithName(ctx.Fctx, name, contentType, content)
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

            res, err := RegisterBlob(ctx.Fctx, contentType, content)
            if err != nil {
                return 0
            }
            
            c.PushString(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "registerBlob")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            name := c.SafeToString(-1)

            res, err := GetBlobTechIdFromName(ctx.Fctx, name)
            if err != nil {
                return 0
            }
            
            c.PushString(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getBlobTechIdFromName")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            name := c.SafeToString(-1)

            res, err := GetBlobBytesAsString(ctx.Fctx, name)
            if err != nil {
                return 0
            }
            
            c.PushString(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "getBlobBytesAsString")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            method := c.SafeToString(-4)
path := c.SafeToString(-3)
name := c.SafeToString(-2)
startFunction := c.SafeToString(-1)

            res, err := PlugFunction(ctx.Fctx, method, path, name, startFunction)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "plugFunction")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            method := c.SafeToString(-3)
path := c.SafeToString(-2)
name := c.SafeToString(-1)

            res, err := PlugFile(ctx.Fctx, method, path, name)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "plugFile")
        
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            
            res, err := GetStatus(ctx.Fctx)
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
            prefix := c.SafeToString(-1)

            res, err := PersistenceGetSubset(ctx.Fctx, prefix)
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

            res, err := CallFunction(ctx.Fctx, name, startFunction, arguments, mode, inputExchangeBufferId, outputExchangeBufferId, posixFileName, posixArguments)
            if err != nil {
                return 0
            }
            
            c.PushInt(res)
    
            return 1
        })
        ctx.Context.PutPropString(-2, "callFunction")
        }