package openglapi

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"my-own-cluster/common"
	"my-own-cluster/opengl"
	"unsafe"
)

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
		if !bindingSpec.IsOUT {
			continue
		}

		fmt.Printf("copying out for binding %d on target %s\n", binding, bindingSpec.Target)

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

func CreateImageFromRgbafloatPixels(ctx *common.FunctionExecutionContext, width int, height int, pixelsExchangeBufferID int, pngExchangeBufferID int) (int, error) {
	pixelsBuffer := ctx.Orchestrator.GetExchangeBuffer(pixelsExchangeBufferID)
	pixels := pixelsBuffer.GetBuffer()

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	l := 4 * width * height
	fb := (*[1 << 26]float32)((*[1 << 26]float32)(unsafe.Pointer(&pixels[0])))[:l:l]

	// Set color for each pixel.
	pixelIndex := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{uint8(10.0 * fb[pixelIndex]), 0, 0, 255})
			pixelIndex += 4
		}
	}

	// Encode as PNG.
	//f, _ := os.Create("image.png")
	//var f bytes.Buffer
	//png.Encode(&f, img)

	png.Encode(ctx.Orchestrator.GetExchangeBuffer(pngExchangeBufferID), img)

	return 0, nil
}
