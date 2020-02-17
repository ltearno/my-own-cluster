package coreapi

import (
	"encoding/base64"
	"fmt"
	"my-own-cluster/common"
)

/**

Implementation of core functions that the my-own-cluster runtime exposes
to functions.

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

func RegisterFunction(ctx *common.FunctionExecutionContext, name string, codeType string, codeBytes []byte) (string, error) {
	techID := ctx.Orchestrator.RegisterFunction(name, codeType, codeBytes)
	return techID, nil
}

func PlugFunction(ctx *common.FunctionExecutionContext, method string, path string, name string, startFunction string) (string, error) {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction)
	return "ok", nil
}

func PlugFile(ctx *common.FunctionExecutionContext, method string, path string, contentType string, fileBytes []byte) (string, error) {
	techID := ctx.Orchestrator.PlugFile(method, path, contentType, fileBytes)
	return techID, nil
}

func Base64Decode(ctx *common.FunctionExecutionContext, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
