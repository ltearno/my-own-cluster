package apicore

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"my-own-cluster/common"
	"my-own-cluster/enginejs"
	"my-own-cluster/enginewasm"
	"net/http"
	"strings"
	"sync"
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
		BindMyOwnClusterFunctionsWASM(wctx, p)
	}

	jsctx, ok := ctx.(*enginejs.JSProcessContext)
	if ok {
		BindMyOwnClusterFunctionsJs(*jsctx, p)
	}
}

/**
 * Implementation of core functions that execution engines can use to exposes functionality to their runtimes
 */

func GetInputBufferID(ctx *common.FunctionExecutionContext, cookie interface{}) (int, error) {
	return ctx.InputExchangeBufferID, nil
}

func GetOutputBufferID(ctx *common.FunctionExecutionContext, cookie interface{}) (int, error) {
	return ctx.OutputExchangeBufferID, nil
}

func FreeBuffer(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int) (int, error) {
	ctx.Orchestrator.ReleaseExchangeBuffer(int(bufferID))
	return 0, nil
}

func PlugFunction(ctx *common.FunctionExecutionContext, cookie interface{}, method string, path string, name string, startFunction string, data string, tagsJSON string) (int, error) {
	ctx.Orchestrator.PlugFunction(method, path, name, startFunction, data, tagsJSON)
	return 0, nil
}

func PlugFile(ctx *common.FunctionExecutionContext, cookie interface{}, method string, path string, name string, tagsJSON string) (int, error) {
	ctx.Orchestrator.PlugFile(method, path, name, tagsJSON)
	return 0, nil
}

func UnplugPath(ctx *common.FunctionExecutionContext, cookie interface{}, method string, path string) (int, error) {
	ctx.Orchestrator.UnplugPath(method, path)
	return 0, nil
}

func PlugFilter(ctx *common.FunctionExecutionContext, cookie interface{}, name string, startFunction string, data string) (string, error) {
	filterID, err := ctx.Orchestrator.PlugFilter(name, startFunction, data)
	return filterID, err
}

func UnplugFilter(ctx *common.FunctionExecutionContext, cookie interface{}, id string) (int, error) {
	ctx.Orchestrator.UnplugFilter(id)
	return 0, nil
}

func RegisterBlobWithName(ctx *common.FunctionExecutionContext, cookie interface{}, name string, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlobWithName(name, contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func RegisterBlob(ctx *common.FunctionExecutionContext, cookie interface{}, contentType string, contentBytes []byte) (string, error) {
	techID, err := ctx.Orchestrator.RegisterBlob(contentType, contentBytes)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func Base64Decode(ctx *common.FunctionExecutionContext, cookie interface{}, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

func Base64Encode(ctx *common.FunctionExecutionContext, cookie interface{}, b []byte) (string, error) {
	return base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(b), nil
}

func WriteExchangeBuffer(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int, content []byte) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.Write(content)

	return len(content), nil
}

func WriteExchangeBufferHeader(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int, name string, value string) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(int(bufferID))
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.SetHeader(string(name), string(value))

	return 0, nil
}

func WriteExchangeBufferStatusCode(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int, statusCode int) (int, error) {
	exchangeBuffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if exchangeBuffer == nil {
		fmt.Printf("GET EXCHANGE BUFFER FOR UNKNOWN BUFFER %d\n", bufferID)
		return 0, nil
	}

	exchangeBuffer.WriteStatusCode(statusCode)

	return 0, nil
}

func ReadExchangeBuffer(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int) ([]byte, error) {
	buffer := ctx.Orchestrator.GetExchangeBuffer(bufferID)
	if buffer == nil {
		fmt.Printf("buffer %d not found for reading\n", bufferID)
		return nil, nil
	}

	bufferBytes := buffer.GetBuffer()

	return bufferBytes, nil
}

func PersistenceSet(ctx *common.FunctionExecutionContext, cookie interface{}, key []byte, value []byte) (int, error) {
	ok := ctx.Orchestrator.PersistenceSet(key, value)
	if !ok {
		return 0, errors.New("cannot persist key !")
	}

	return 0, nil
}

func CreateExchangeBuffer(ctx *common.FunctionExecutionContext, cookie interface{}) (int, error) {
	return ctx.Orchestrator.CreateExchangeBuffer(), nil
}

func ReadExchangeBufferHeaders(ctx *common.FunctionExecutionContext, cookie interface{}, bufferID int) (map[string]string, error) {
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

func GetUrl(ctx *common.FunctionExecutionContext, cookie interface{}, url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	return bytes, nil
}

func PersistenceGet(ctx *common.FunctionExecutionContext, cookie interface{}, key []byte) ([]byte, error) {
	value, present := ctx.Orchestrator.PersistenceGet(key)
	if !present {
		return nil, nil
	}

	return value, nil
}

func PrintDebug(ctx *common.FunctionExecutionContext, cookie interface{}, text string) (int, error) {
	fmt.Printf("[print_debug %s::%s] %s\n", ctx.Name, ctx.StartFunction, text)

	return 0, nil
}

func GetTime(ctx *common.FunctionExecutionContext, cookie interface{}, destBuffer []byte) (int, error) {
	*(*int64)(unsafe.Pointer(&destBuffer[0])) = time.Now().UnixNano()

	return 0, nil
}

func GetBlobTechIdFromName(ctx *common.FunctionExecutionContext, cookie interface{}, name string) (string, error) {
	techID, err := ctx.Orchestrator.GetBlobTechIDFromName(name)

	return techID, err
}

func GetBlobBytesAsString(ctx *common.FunctionExecutionContext, cookie interface{}, techID string) (string, error) {
	contentBytes, err := ctx.Orchestrator.GetBlobBytesByTechID(techID)

	return string(contentBytes), err
}

func GetStatus(ctx *common.FunctionExecutionContext, cookie interface{}) (string, error) {
	return ctx.Orchestrator.GetStatus(), nil
}

func CallFunction(ctx *common.FunctionExecutionContext, cookie interface{}, name string, startFunction string, arguments []int, mode string, inputExchangeBufferID int, outputExchangeBufferID int, posixFileName string, posixArguments []string) (int, error) {
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

func PersistenceGetSubset(ctx *common.FunctionExecutionContext, cookie interface{}, prefix string) (map[string]string, error) {
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

func ExportDatabase(ctx *common.FunctionExecutionContext, cookie interface{}) ([]byte, error) {
	res, err := ctx.Orchestrator.GetDatabaseExport("")
	if err != nil {
		return nil, err
	}

	return res, nil
}

type ProxySpec struct {
	Method                 string            `json:"method"`
	Url                    string            `json:"url"`
	Headers                map[string]string `json:"headers"`
	InputExchangeBufferID  int               `json:"inputBufferId"`
	OutputExchangeBufferID int               `json:"outputBufferId"`
}

func BetaWebProxy(ctx *common.FunctionExecutionContext, cookie interface{}, proxySpecJSON string) (int, error) {
	spec := &ProxySpec{}
	err := json.Unmarshal([]byte(proxySpecJSON), &spec)
	if err != nil {
		return -1, err
	}

	if ctx.Trace {
		fmt.Printf("BETA PROXY to %s\n", spec.Url)
	}

	var reqID, respID int

	if strings.HasPrefix(spec.Url, "http") {
		reqID, respID, err = ctx.Orchestrator.CreateExchangeBuffersFromHttpClientRequest(spec.Method, spec.Url, spec.Headers)
		if err != nil {
			fmt.Printf("ERROR 555 %v\n", err)
			return -1, err
		}
	} else if strings.HasPrefix(spec.Url, "ws") {
		reqID, respID, err = ctx.Orchestrator.CreateExchangeBuffersFromWebSocketClient(spec.Method, spec.Url, spec.Headers)
		if err != nil {
			fmt.Printf("ERROR 444 %v\n", err)
			return -1, err
		}
	} else {
		fmt.Printf("ERROR 2321 %s\n", spec.Url)
		return -1, fmt.Errorf("unknown url '%s'", spec.Url)
	}

	req := ctx.Orchestrator.GetExchangeBuffer(reqID)
	resp := ctx.Orchestrator.GetExchangeBuffer(respID)
	input := ctx.Orchestrator.GetExchangeBuffer(spec.InputExchangeBufferID)
	output := ctx.Orchestrator.GetExchangeBuffer(spec.OutputExchangeBufferID)

	var wg sync.WaitGroup
	wg.Add(2)

	if ctx.Trace {
		fmt.Printf("LAUNCH LOOPS\n")
	}

	go func() {
		defer wg.Done()
		for {
			i := input.GetBuffer()
			if i == nil || len(i) == 0 {
				if ctx.Trace {
					fmt.Printf("INPUT FINISHED\n")
				}
				req.Close()
				return
			}

			if ctx.Trace {
				fmt.Printf("READ %d FROM INPUT\n", len(i))
			}
			req.Write(i)
		}
	}()

	go func() {
		defer wg.Done()

		resp.GetHeaders(func(k string, v string) {
			if ctx.Trace {
				fmt.Printf("transmitting backend header '%s'\n", k)
			}
			output.SetHeader(k, v)
		})

		output.WriteStatusCode(resp.GetStatusCode())

		for {
			o := resp.GetBuffer()
			if o == nil {
				if ctx.Trace {
					fmt.Printf("RESPONSE FINISHED\n")
				}
				output.Close()
				return
			}

			if ctx.Trace {
				fmt.Printf("READ %d FROM RESPONSE\n", len(o))
			}
			output.Write(o)
		}
	}()

	wg.Wait()

	if ctx.Trace {
		fmt.Printf("LOOPS FINISHED\n")
	}

	return 0, nil
}

func IsTrace(ctx *common.FunctionExecutionContext, cookie interface{}) (int, error) {
	if ctx.Trace {
		return 1, nil
	} else {
		return 0, nil
	}
}
