package apijwt

    import (
        "github.com/ltearno/my-own-cluster/enginewasm"

        
        
    )
    

func BindJwtFunctionsWASM(wctx enginewasm.WasmProcessContext, cookie interface{}) {
	wctx.BindAPIFunction("jwt", "verify_jwt", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        jwt := cs.GetParamString(0, 1)


        

        res, err := VerifyJwt(wctx.Fctx, cookie, jwt)
        if err != nil {
            return uint32(0xffff), err
        }
        
        
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil
    })
    }