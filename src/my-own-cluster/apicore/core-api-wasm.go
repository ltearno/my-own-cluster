package apicore

    import (
        "my-own-cluster/enginewasm"

        "bytes"
        "encoding/binary"
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
    
	wctx.BindAPIFunction("my-own-cluster", "get_buffer_headers", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        

        res, err := GetBufferHeaders(wctx.Fctx, bufferId)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                var b bytes.Buffer
                bs := make([]byte, 4)

                // pair count
                binary.LittleEndian.PutUint32(bs, uint32(len(res)))
                b.Write(bs)

                for k, v := range res {
                    // write key
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(k))))
                    b.Write(bs)
                    b.Write([]byte(k))

                    // write value
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(v))))
                    b.Write(bs)
                    b.Write([]byte(v))
                }

                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(b.Bytes())

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
    
	wctx.BindAPIFunction("my-own-cluster", "base64_encode", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        input := cs.GetParamByteBuffer(0, 1)


        

        res, err := Base64Encode(wctx.Fctx, input)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "register_blob_with_name", "i(iiiiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        name := cs.GetParamString(0, 1)
contentType := cs.GetParamString(2, 3)
content := cs.GetParamByteBuffer(4, 5)


        

        res, err := RegisterBlobWithName(wctx.Fctx, name, contentType, content)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "register_blob", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        contentType := cs.GetParamString(0, 1)
content := cs.GetParamByteBuffer(2, 3)


        

        res, err := RegisterBlob(wctx.Fctx, contentType, content)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_blob_tech_id_from_name", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        name := cs.GetParamString(0, 1)


        

        res, err := GetBlobTechIdFromName(wctx.Fctx, name)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_blob_bytes_as_string", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        name := cs.GetParamString(0, 1)


        

        res, err := GetBlobBytesAsString(wctx.Fctx, name)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "plug_function", "i(iiiiiiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        method := cs.GetParamString(0, 1)
path := cs.GetParamString(2, 3)
name := cs.GetParamString(4, 5)
startFunction := cs.GetParamString(6, 7)


        

        res, err := PlugFunction(wctx.Fctx, method, path, name, startFunction)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "plug_file", "i(iiiiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        method := cs.GetParamString(0, 1)
path := cs.GetParamString(2, 3)
name := cs.GetParamString(4, 5)


        

        res, err := PlugFile(wctx.Fctx, method, path, name)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_status", "i()", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        

        

        res, err := GetStatus(wctx.Fctx)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
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
    
	wctx.BindAPIFunction("my-own-cluster", "persistence_get", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        key := cs.GetParamByteBuffer(0, 1)


        

        res, err := PersistenceGet(wctx.Fctx, key)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                    resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                    resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                    resultBuffer.Write(res)
                    return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "persistence_get_subset", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        prefix := cs.GetParamString(0, 1)


        

        res, err := PersistenceGetSubset(wctx.Fctx, prefix)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                var b bytes.Buffer
                bs := make([]byte, 4)

                // pair count
                binary.LittleEndian.PutUint32(bs, uint32(len(res)))
                b.Write(bs)

                for k, v := range res {
                    // write key
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(k))))
                    b.Write(bs)
                    b.Write([]byte(k))

                    // write value
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(v))))
                    b.Write(bs)
                    b.Write([]byte(v))
                }

                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(b.Bytes())

                return uint32(resultBufferID), nil
    })
    
	wctx.BindAPIFunction("my-own-cluster", "print_debug", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        text := cs.GetParamString(0, 1)


        

        res, err := PrintDebug(wctx.Fctx, text)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "get_time", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        dest := cs.GetParamByteBuffer(0, 1)


        

        res, err := GetTime(wctx.Fctx, dest)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    
	wctx.BindAPIFunction("my-own-cluster", "free_buffer", "i(i)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        bufferId := cs.GetParamInt(0)


        

        res, err := FreeBuffer(wctx.Fctx, bufferId)
        if err != nil {
            return uint32(0xffff), err
        }
        
        return uint32(res), err
    })
    }