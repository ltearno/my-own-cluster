package apigpu

import (
	"errors"
	"fmt"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"unsafe"
)

/*
#cgo LDFLAGS: -lGL -lEGL -lgbm
#include "engine-opengl.h"
*/
import "C"

type GLSLOpenGLDeviceContext struct {
	ctx *C.DeviceInformation
}

type GLSLOpenGLContext struct {
	ctx  *C.ContextInformation
	fctx *common.FunctionExecutionContext
}

type GPUAPIProvider struct {
}

func NewGPUAPIProvider() (common.APIProvider, error) {
	return &GPUAPIProvider{}, nil
}

func (p *GPUAPIProvider) BindToExecutionEngineContext(ctx common.ExecutionEngineContextBounding) {
	wctx, ok := ctx.(*enginewasm.WasmProcessContext)
	if ok {
		BindOpenGLFunctionsWASM(*wctx, p)
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		BindOpenGLFunctionsJs(*jsctx, p)
	}
}

var openGLDeviceContext *GLSLOpenGLDeviceContext = nil

func InitOpenGLContext(fctx *common.FunctionExecutionContext) (*GLSLOpenGLContext, error) {
	if openGLDeviceContext == nil {
		openGLDeviceContext = &GLSLOpenGLDeviceContext{
			ctx: &C.DeviceInformation{},
		}
		res := C.initOpenGLDevice(openGLDeviceContext.ctx)
		if res != 0 {
			return nil, fmt.Errorf("cannot instantiate OpenGL device (%d)", res)
		}
	}

	ctx := &GLSLOpenGLContext{
		ctx:  &C.ContextInformation{},
		fctx: fctx,
	}
	C.initOpenGLContext(openGLDeviceContext.ctx, ctx.ctx)

	return ctx, nil
}

func (c *GLSLOpenGLContext) SetContext() {
	C.setContext(openGLDeviceContext.ctx, c.ctx)
}

func (c *GLSLOpenGLContext) ClearContext() {
	C.clearContext(openGLDeviceContext.ctx, c.ctx)
}

func (c *GLSLOpenGLContext) CompileAndBindShader(shaderCodeBytes []byte) error {
	res := C.compileAndBindShader(C.CString(string(shaderCodeBytes)))
	if res < 0 {
		return errors.New("cannot compile and bind shader")
	}

	if c.fctx.Trace {
		fmt.Printf("compiled and linked shader\n")
	}

	return nil
}

func (c *GLSLOpenGLContext) BindStorageBuffer(binding int, bufferContent []byte) (int, error) {
	bufferIndex := C.bindStorageBuffer(C.int(binding), C.CBytes(bufferContent), C.int(len(bufferContent)))
	if c.fctx.Trace {
		fmt.Printf("bound storage buffer to index %d\n", bufferIndex)
	}
	return int(bufferIndex), nil
}

func (c *GLSLOpenGLContext) BindTexture2DRGBAFloat(binding int, width int, height int) (int, error) {
	textureIndex := C.bindTexture2DRGBAFloat(C.int(binding), C.int(width), C.int(height))
	if c.fctx.Trace {
		fmt.Printf("bound texture to index %d\n", textureIndex)
	}
	return int(textureIndex), nil
}

func (c *GLSLOpenGLContext) BindTexture2DRFloat(binding int, width int, height int) (int, error) {
	textureIndex := C.bindTexture2DRFloat(C.int(binding), C.int(width), C.int(height))
	if c.fctx.Trace {
		fmt.Printf("bound texture to index %d\n", textureIndex)
	}
	return int(textureIndex), nil
}

func (c *GLSLOpenGLContext) DispatchCompute(x int, y int, z int) {
	if c.fctx.Trace {
		fmt.Printf("dispatching compute %d %d %d\n", x, y, z)
	}
	C.glDispatchCompute(C.uint(x), C.uint(y), C.uint(z))
	C.glMemoryBarrier(C.GL_ALL_BARRIER_BITS)
	if c.fctx.Trace {
		fmt.Println("dispatched ok")
	}
}

func (c *GLSLOpenGLContext) GetStorageBuffer(bufferIndex int, buffer []byte) error {
	if c.fctx.Trace {
		fmt.Printf("getting buffer %d content\n", bufferIndex)
	}
	C.getStorageBuffer(C.int(bufferIndex), unsafe.Pointer(&buffer[0]), C.int(len(buffer)))
	return nil
}

func (c *GLSLOpenGLContext) GetTexture2DRGBAFloatBuffer(textureIndex int, buffer []byte) error {
	if c.fctx.Trace {
		fmt.Printf("getting texture %d content\n", textureIndex)
	}
	C.getTexture2DRGBAFloatBuffer(C.int(textureIndex), unsafe.Pointer(&buffer[0]), C.int(len(buffer)))
	return nil
}

func (c *GLSLOpenGLContext) GetTexture2DRFloatBuffer(textureIndex int, buffer []byte) error {
	if c.fctx.Trace {
		fmt.Printf("getting texture %d content\n", textureIndex)
	}
	C.getTexture2DRFloatBuffer(C.int(textureIndex), unsafe.Pointer(&buffer[0]), C.int(len(buffer)))
	return nil
}

func (c *GLSLOpenGLContext) DeleteBuffer(bufferIndex int) error {
	C.deleteBuffer(C.int(bufferIndex))
	return nil
}

func (c *GLSLOpenGLContext) DeleteTexture(textureIndex int) error {
	C.deleteTexture(C.int(textureIndex))
	return nil
}
