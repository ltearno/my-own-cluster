package apicore

    import (
        "my-own-cluster/enginewasm"
    )
    

func BindMyOwnClusterFunctionsWASM(wctx enginewasm.WasmProcessContext) {
	wctx.BindAPIFunction("my-own-cluster", "get_input_buffer_id", "i()", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        

        

        res, err := GetInputBufferID(wctx.Fctx)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_output_buffer_id", "i()", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        

        

        res, err := GetOutputBufferID(wctx.Fctx)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "create_exchange_buffer", "i()", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        

        

        res, err := CreateExchangeBuffer(wctx.Fctx)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "write_exchange_buffer", "i(iii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)
content := cs.GetParamByteBuffer(1, 2)


        

        res, err := WriteExchangeBuffer(wctx.Fctx, bufferId, content)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "write_exchange_buffer_header", "i(iiiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)
name := cs.GetParamString(1, 2)
value := cs.GetParamString(3, 4)


        

        res, err := WriteExchangeBufferHeader(wctx.Fctx, bufferId, name, value)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_exchange_buffer_size", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        

        res, err := GetExchangeBufferSize(wctx.Fctx, bufferId)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "read_exchange_buffer", "i(iii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        resultBuffer := cs.GetParamByteBuffer(1, 2)

        res, err := ReadExchangeBuffer(wctx.Fctx, bufferId)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                if resultBuffer != nil && len(resultBuffer)>=len(res) {
                    copy(resultBuffer, res)
                    return uint32(len(res)), nil
                } else {
                    return uint32(len(res)), nil
                }
    })
    
	wctx.BindAPIFunction("my-own-cluster", "read_exchange_buffer_headers", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        

        res, err := ReadExchangeBufferHeaders(wctx.Fctx, bufferId)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(res)
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "base64_decode", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        encoded := cs.GetParamString(0, 1)


        

        res, err := Base64Decode(wctx.Fctx, encoded)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(res)
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "persistence_set", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        key := cs.GetParamByteBuffer(0, 1)
value := cs.GetParamByteBuffer(2, 3)


        

        res, err := PersistenceSet(wctx.Fctx, key, value)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_url", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        url := cs.GetParamString(0, 1)


        

        res, err := GetUrl(wctx.Fctx, url)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(res)
                return uint32(resultBufferID), nil
    })
    }