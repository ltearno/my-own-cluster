package apigpu

import (
	"github.com/ltearno/my-own-cluster/enginewasm"
)

func BindOpenGLFunctionsWASM(wctx enginewasm.WasmProcessContext, cookie interface{}) {
	wctx.BindAPIFunction("gpu", "compute_shader", "i(ii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
		specification := cs.GetParamString(0, 1)

		res, err := ComputeShader(wctx.Fctx, cookie, specification)
		if err != nil {
			return uint32(0xffff), err
		}

		return uint32(res), err
	})

	wctx.BindAPIFunction("gpu", "create_image_from_rgba_float_pixels", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
		width := cs.GetParamInt(0)
		height := cs.GetParamInt(1)
		pixelsExchangeBufferId := cs.GetParamInt(2)
		pngExchangeBufferId := cs.GetParamInt(3)

		res, err := CreateImageFromRgbaFloatPixels(wctx.Fctx, cookie, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
		if err != nil {
			return uint32(0xffff), err
		}

		return uint32(res), err
	})

	wctx.BindAPIFunction("gpu", "create_image_from_r_float_pixels", "i(iiii)", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
		width := cs.GetParamInt(0)
		height := cs.GetParamInt(1)
		pixelsExchangeBufferId := cs.GetParamInt(2)
		pngExchangeBufferId := cs.GetParamInt(3)

		res, err := CreateImageFromRFloatPixels(wctx.Fctx, cookie, width, height, pixelsExchangeBufferId, pngExchangeBufferId)
		if err != nil {
			return uint32(0xffff), err
		}

		return uint32(res), err
	})
}
