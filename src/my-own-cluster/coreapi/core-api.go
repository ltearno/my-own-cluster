package coreapi

import (
	"encoding/base64"
	"fmt"
	"my-own-cluster/common"
)

/**
 * Implementation of core functions that execution engines can use to exposes functionality to their runtimes
 */

func Test(ctx *common.FunctionExecutionContext) (int, error) {
	fmt.Printf("guest called 'test()'\n")
	return 0, nil
}

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
