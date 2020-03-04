package apigpu

    import (
        "my-own-cluster/enginewasm"
    )
    

func BindOpenGLFunctionsWASM(wctx enginewasm.WasmProcessContext) {
    // wasm params : specification 
	wctx.BindAPIFunction("my-own-cluster", "compute_shader", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        specification := cs.GetParamString(0, 1)


        

        res, err := ComputeShader(wctx.Fctx, specification)
        
        return uint32(res), err
    })
    
    // wasm params : width height pixels_exchange_buffer_id png_exchange_buffer_id 
	wctx.BindAPIFunction("my-own-cluster", "create_image_from_rgba_float_pixels", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        width := cs.GetParamInt(0)
height := cs.GetParamInt(1)
pixelsExchangeBufferId := cs.GetParamInt(2)
pngExchangeBufferId := cs.GetParamInt(3)


        

        res, err := CreateImageFromRgbaFloatPixels(wctx.Fctx, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
        
        return uint32(res), err
    })
    
    // wasm params : width height pixels_exchange_buffer_id png_exchange_buffer_id 
	wctx.BindAPIFunction("my-own-cluster", "create_image_from_r_float_pixels", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        width := cs.GetParamInt(0)
height := cs.GetParamInt(1)
pixelsExchangeBufferId := cs.GetParamInt(2)
pngExchangeBufferId := cs.GetParamInt(3)


        

        res, err := CreateImageFromRFloatPixels(wctx.Fctx, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
        
        return uint32(res), err
    })
    }