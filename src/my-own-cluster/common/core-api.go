package common

import (
	"encoding/base64"
	"fmt"
)

/**

Implementation of core functions that the my-own-cluster runtime exposes
to functions.

*/

func Test(ctx *FunctionExecutionContext) (int, error) {
	fmt.Printf("guest called 'test()'\n")
	return 0, nil
}

func GetInputBufferID(ctx *FunctionExecutionContext) (int, error) {
	return ctx.InputExchangeBufferID, nil
}

func GetOutputBufferID(ctx *FunctionExecutionContext) (int, error) {
	return ctx.OutputExchangeBufferID, nil
}

func FreeBuffer(ctx *FunctionExecutionContext, bufferID int) error {
	ctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
	return nil
}

func PlugFunction(ctx *FunctionExecutionContext, method string, path string, name string, startFunction string) error {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction)
	return nil
}

func PlugFile(ctx *FunctionExecutionContext, method string, path string, name string) error {
	ctx.Orchestrator.PlugFile(method, path, name)
	return nil
}

func RegisterBlobWithName(ctx *FunctionExecutionContext, name string, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlobWithName(name, contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func RegisterBlob(ctx *FunctionExecutionContext, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlob(contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func Base64Decode(ctx *FunctionExecutionContext, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
