package wasm

    import (
        "my-own-cluster/coreapi"
    )

func BindMyOwnClusterFunctionsWASM(wctx *WasmProcessContext) {
    // wasm params :  
	wctx.BindAPIFunction("my-own-cluster", "get_input_buffer_id", "i()", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        

        

        res, err := coreapi.GetInputBufferID(wctx.Fctx)
        
        return uint32(res), err
    })
    
    // wasm params :  
	wctx.BindAPIFunction("my-own-cluster", "get_output_buffer_id", "i()", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        

        

        res, err := coreapi.GetOutputBufferID(wctx.Fctx)
        
        return uint32(res), err
    })
    
    // wasm params :  
	wctx.BindAPIFunction("my-own-cluster", "create_exchange_buffer", "i()", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        

        

        res, err := coreapi.CreateExchangeBuffer(wctx.Fctx)
        
        return uint32(res), err
    })
    
    // wasm params : buffer_id content 
	wctx.BindAPIFunction("my-own-cluster", "write_exchange_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)
content := cs.GetParamByteBuffer(1, 2)


        

        res, err := coreapi.WriteExchangeBuffer(wctx.Fctx, bufferId, content)
        
        return uint32(res), err
    })
    
    // wasm params : buffer_id name value 
	wctx.BindAPIFunction("my-own-cluster", "write_exchange_buffer_header", "i(iiiii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)
name := cs.GetParamString(1, 2)
value := cs.GetParamString(3, 4)


        

        res, err := coreapi.WriteExchangeBufferHeader(wctx.Fctx, bufferId, name, value)
        
        return uint32(res), err
    })
    
    // wasm params : buffer_id result_buffer_addr result_buffer_length
	wctx.BindAPIFunction("my-own-cluster", "read_exchange_buffer", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        resultBuffer := cs.GetParamByteBuffer(1, 2)

        res, err := coreapi.ReadExchangeBuffer(wctx.Fctx, bufferId)
        
        if resultBuffer != nil && len(resultBuffer)>=len(res) {
                copy(resultBuffer, res)
        }
        return uint32(len(res)), err
    })
    
    // wasm params : buffer_id result_buffer_addr result_buffer_length
	wctx.BindAPIFunction("my-own-cluster", "read_exchange_buffer_headers", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        resultBuffer := cs.GetParamByteBuffer(1, 2)

        res, err := coreapi.ReadExchangeBufferHeaders(wctx.Fctx, bufferId)
        
        if resultBuffer != nil && len(resultBuffer)>=len(res) {
                copy(resultBuffer, res)
        }
        return uint32(len(res)), err
    })
    
    // wasm params : encoded result_buffer_addr result_buffer_length
	wctx.BindAPIFunction("my-own-cluster", "base64_decode", "i(iiii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        encoded := cs.GetParamString(0, 1)


        resultBuffer := cs.GetParamByteBuffer(2, 3)

        res, err := coreapi.Base64Decode(wctx.Fctx, encoded)
        
        if resultBuffer != nil && len(resultBuffer)>=len(res) {
                copy(resultBuffer, res)
        }
        return uint32(len(res)), err
    })
    
    // wasm params : key value 
	wctx.BindAPIFunction("my-own-cluster", "persistence_set", "i(iiii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        key := cs.GetParamByteBuffer(0, 1)
value := cs.GetParamByteBuffer(2, 3)


        

        res, err := coreapi.PersistenceSet(wctx.Fctx, key, value)
        
        return uint32(res), err
    })
    }