package apicore

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"net/http"
	"time"
	"unsafe"
)

type CoreAPIProvider struct{}

func NewCoreAPIProvider() (common.APIProvider, error) {
	return &CoreAPIProvider{}, nil
}

func (p *CoreAPIProvider) BindToExecutionEngineContext(ctx common.ExecutionEngineContextBounding) {
	wctx, ok := ctx.(*enginewasm.WasmProcessContext)
	if ok {
		wctx := *wctx
		BindMyOwnClusterFunctionsWASM(wctx)
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		BindMyOwnClusterFunctionsJs(*jsctx)
	}
}

/**
 * Implementation of core functions that execution engines can use to exposes functionality to their runtimes
 */

func GetInputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.InputExchangeBufferID, nil
}

func GetOutputBufferID(ctx *common.FunctionExecutionContext) (int, error) {
	return ctx.OutputExchangeBufferID, nil
}

func FreeBuffer(ctx *common.FunctionExecutionContext, bufferID int) (int, error) {
	ctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
	return 0, nil
}

func PlugFunction(ctx *common.FunctionExecutionContext, method string, path string, name string, startFunction string) (int, error) {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction)
	return 0, nil
}

func PlugFile(ctx *common.FunctionExecutionContext, method string, path string, name string) (int, error) {
	ctx.Orchestrator.PlugFile(method, path, name)
	return 0, nil
}

func UnplugPath(ctx *common.FunctionExecutionContext, method string, path string) (int, error) {
	ctx.Orchestrator.UnplugPath(method, path)
	return 0, nil
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

func Base64Encode(ctx *common.FunctionExecutionContext, b []byte) (string, error) {
	return base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(b), nil
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

func WriteExchangeBufferHeader(ctx *common.FunctionExecutionContext, bufferID int, name string, value string) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(int(bufferID))
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.SetHeader(string(name), string(value))

	return 0, nil
}

func GetExchangeBufferSize(ctx *common.FunctionExecutionContext, bufferID int) (int, error) {
	buffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if buffer == nil {
		fmt.Printf("buffer %d not found for getting size\n", bufferID)
		return -1, nil
	}

	bufferBytes := buffer.GetBuffer()

	return len(bufferBytes), nil
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

func ReadExchangeBufferHeaders(ctx *common.FunctionExecutionContext, bufferID int) (map[string]string, error) {
	res := make(map[string]string)

	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(int(bufferID))
	if exchangeBuffer == nil {
		return nil, nil
	}

	exchangeBuffer.GetHeaders(func(name string, value string) {
		res[name] = value
	})
	return res, nil
}

func GetUrl(ctx *common.FunctionExecutionContext, url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	return bytes, nil
}

func PersistenceGet(ctx *common.FunctionExecutionContext, key []byte) ([]byte, error) {
	value, present := ctx.Orchestrator.PersistenceGet(key)
	if !present {
		return nil, nil
	}

	return value, nil
}

func PrintDebug(ctx *common.FunctionExecutionContext, text string) (int, error) {
	fmt.Printf("[print_debug %s::%s] %s\n", ctx.Name, ctx.StartFunction, text)

	return 0, nil
}

func GetTime(ctx *common.FunctionExecutionContext, destBuffer []byte) (int, error) {
	*(*int64)(unsafe.Pointer(&destBuffer[0])) = time.Now().UnixNano()

	return 0, nil
}

func GetBlobTechIdFromName(ctx *common.FunctionExecutionContext, name string) (string, error) {
	techID, err := ctx.Orchestrator.GetBlobTechIDFromName(name)

	return techID, err
}

func GetBlobBytesAsString(ctx *common.FunctionExecutionContext, techID string) (string, error) {
	contentBytes, err := ctx.Orchestrator.GetBlobBytesByTechID(techID)

	return string(contentBytes), err
}

func GetStatus(ctx *common.FunctionExecutionContext) (string, error) {
	return ctx.Orchestrator.GetStatus(), nil
}

func CallFunction(ctx *common.FunctionExecutionContext, name string, startFunction string, arguments []int, mode string, inputExchangeBufferID int, outputExchangeBufferID int, posixFileName string, posixArguments []string) (int, error) {
	newCtx := ctx.Orchestrator.NewFunctionExecutionContext(
		name,
		startFunction,
		arguments,
		ctx.Trace,
		mode,
		&posixFileName,
		&posixArguments,
		inputExchangeBufferID,
		outputExchangeBufferID,
	)

	err := newCtx.Run()
	if err != nil {
		fmt.Printf("[ERROR] callFunction failed (%v)\n", err)
		return -1, err
	}

	return newCtx.Result, nil
}

func PersistenceGetSubset(ctx *common.FunctionExecutionContext, prefix string) (map[string]string, error) {
	subset, err := ctx.Orchestrator.PersistenceGetSubset([]byte(prefix))
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	i := 0
	for i < len(subset) {
		k := subset[i]
		i++
		v := subset[i]
		i++
		res[string(k)] = string(v)
	}
	return res, nil
}

func ExportDatabase(ctx *common.FunctionExecutionContext) ([]byte, error) {
	res, err := ctx.Orchestrator.GetDatabaseExport("")
	if err != nil {
		return nil, err
	}

	return res, nil
}
