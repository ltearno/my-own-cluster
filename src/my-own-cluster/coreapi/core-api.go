package coreapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"my-own-cluster/common"
	"my-own-cluster/opengl"
	"os"
	"unsafe"
)

/**
 * Implementation of core functions that execution engines can use to exposes functionality to their runtimes
 */

func GetInputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.InputExchangeBufferID, nil
}

func GetOutputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.OutputExchangeBufferID, nil
}

func FreeBuffer(ctx *common.FunctionExecutionContext, bufferID int) error {
	ctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
	return nil
}

func PlugFunction(ctx *common.FunctionExecutionContext, method string, path string, name string, startFunction string) error {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction)
	return nil
}

func PlugFile(ctx *common.FunctionExecutionContext, method string, path string, name string) error {
	ctx.Orchestrator.PlugFile(method, path, name)
	return nil
}

func RegisterBlobWithName(ctx *common.FunctionExecutionContext, name string, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlobWithName(name, contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func RegisterBlob(ctx *common.FunctionExecutionContext, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlob(contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func Base64Decode(ctx *common.FunctionExecutionContext, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func Base64Encode(ctx *common.FunctionExecutionContext, b []byte) string {
	return base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(b)
}

func WriteExchangeBuffer(ctx *common.FunctionExecutionContext, bufferID int, content []byte) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.Write(content)

	return len(content), nil
}

func ReadExchangeBuffer(ctx *common.FunctionExecutionContext, bufferID int) ([]byte, error) {
	buffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if buffer == nil {
		fmt.Printf("buffer %d not found for reading\n", bufferID)
		return nil, nil
	}

	bufferBytes := buffer.GetBuffer()

	return bufferBytes, nil
}

func PersistenceSet(ctx *common.FunctionExecutionContext, key []byte, value []byte) (int, error) {
	ok := ctx.Orchestrator.PersistenceSet(key, value)
	if !ok {
		return 0, errors.New("cannot persist key !")
	}

	return 0, nil
}

func CreateExchangeBuffer(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.Orchestrator.CreateExchangeBuffer(), nil
}

type ComputeShaderBindingSpecification struct {
	Target           string `json:"target"` // STORAGE, TEXTURE_2D_RGBA_FLOAT
	ExchangeBufferID int    `json:"exchange_buffer_id"`
	IsIN             bool   `json:"is_in"`
	IsOUT            bool   `json:"is_out"`
	Width            int    `json:"width,omitempty"`
	Height           int    `json:"height,omitempty"`
}

type ComputeShaderSpecification struct {
	ShaderName   string                              `json:"shader_name"`
	Bindings     []ComputeShaderBindingSpecification `json:"bindings"`
	DispatchSize [3]int                              `json:"dispatch_size"`
}

func ComputeShader(ctx *common.FunctionExecutionContext, specificationJSON string) (int, error) {
	// TODO later : maybe the present node does not have a GPU and we need to launch this elsewhere...
	// TODO later : maybe this should be catched by the opengl exec engine...

	fmt.Printf("COMPUTE SHADER BEGINS\n")

	spec := &ComputeShaderSpecification{}
	err := json.Unmarshal([]byte(specificationJSON), spec)
	if err != nil {
		fmt.Printf("cannot unmarshall spec\n")
		return -1, nil
	}

	openglCtx, err := opengl.InitOpenGLContext()
	if err != nil {
		fmt.Printf("cannot init opengl\n")
		return -2, err
	}

	openglCtx.SetContext()

	// get shader code
	shaderCodeBlobTechID, err := ctx.Orchestrator.GetBlobTechIDFromReference(spec.ShaderName)
	if err != nil {
		fmt.Printf("cannot get shader code blob techID\n")
		return -3, err
	}
	shaderCodeBytes, err := ctx.Orchestrator.GetBlobBytesByTechID(shaderCodeBlobTechID)
	if err != nil {
		fmt.Printf("cannot get shader code bytes\n")
		return -4, err
	}

	err = openglCtx.CompileAndBindShader(shaderCodeBytes)
	if err != nil {
		fmt.Printf("cannot compile shader\n")
		return -5, err
	}

	storageIndices := make(map[int]int)
	textureIndices := make(map[int]int)

	// parse context buffers and create bounded buffers and textures
	for binding, bindingSpec := range spec.Bindings {
		switch bindingSpec.Target {
		case "STORAGE":
			buffer := ctx.Orchestrator.GetExchangeBuffer(bindingSpec.ExchangeBufferID).GetBuffer()
			bufferIndex, err := openglCtx.BindStorageBuffer(binding, buffer)
			if err != nil {
				fmt.Printf("cannot bind buffer\n")
				return -18, err
			}
			storageIndices[binding] = bufferIndex
			break

		case "TEXTURE_2D_RGBA_FLOAT":
			textureIndex, err := openglCtx.BindTexture2DRGBAFloat(binding, bindingSpec.Width, bindingSpec.Height)
			if err != nil {
				fmt.Printf("cannot bind texture\n")
				return -18, err
			}
			textureIndices[binding] = textureIndex
			break
		}
	}

	// dispatch compute
	openglCtx.DispatchCompute(spec.DispatchSize[0], spec.DispatchSize[1], spec.DispatchSize[2])

	// copy out buffers to the specified exchange buffers
	for binding, bindingSpec := range spec.Bindings {
		fmt.Printf("copying out for binding %d on target %s\n", binding, bindingSpec.Target)

		if !bindingSpec.IsOUT {
			continue
		}

		switch bindingSpec.Target {
		case "STORAGE":
			buffer := ctx.Orchestrator.GetExchangeBuffer(bindingSpec.ExchangeBufferID).GetBuffer()
			bufferIndex := storageIndices[binding]

			err := openglCtx.GetStorageBuffer(bufferIndex, buffer)
			if err != nil {
				fmt.Printf("cannot read output buffer\n")
				return -20, err
			}
			break

		case "TEXTURE_2D_RGBA_FLOAT":
			buffer := ctx.Orchestrator.GetExchangeBuffer(bindingSpec.ExchangeBufferID).GetBuffer()
			textureIndex := textureIndices[binding]

			err := openglCtx.GetTextureBuffer(textureIndex, buffer)
			if err != nil {
				fmt.Printf("cannot read output texture\n")
				return -20, err
			}

			width := 1024
			height := width
			upLeft := image.Point{0, 0}
			lowRight := image.Point{width, height}

			img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

			l := 4 * width * height
			fb := (*[1 << 26]float32)((*[1 << 26]float32)(unsafe.Pointer(&buffer[0])))[:l:l]

			// Set color for each pixel.
			pixelIndex := 0
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					img.Set(x, y, color.RGBA{uint8(10.0 * fb[pixelIndex]), 0, 0, 255})
					pixelIndex += 4
				}
			}

			// Encode as PNG.
			f, _ := os.Create("image.png")
			png.Encode(f, img)
			break
		}
	}

	// done !
	for _, bufferIndex := range storageIndices {
		fmt.Printf("delete buffer %d\n", bufferIndex)
		openglCtx.DeleteBuffer(bufferIndex)
	}
	for _, textureIndex := range textureIndices {
		fmt.Printf("delete texture %d\n", textureIndex)
		openglCtx.DeleteTexture(textureIndex)
	}

	openglCtx.ClearContext()

	return 0, nil
}
