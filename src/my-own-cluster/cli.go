package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"my-own-cluster/common"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RegisterBlobWithNameRequest struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Bytes       string `json:"bytes"`
}

type RegisterBlobWithNameResponse struct {
	Status bool   `json:"status"`
	TechID string `json:"tech_id"`
}

type RegisterBlobRequest struct {
	ContentType string `json:"content_type"`
	Bytes       string `json:"bytes"`
}

type RegisterBlobResponse struct {
	Status bool   `json:"status"`
	TechID string `json:"tech_id"`
}

type PlugFunctionRequest struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	Name          string `json:"name"`
	StartFunction string `json:"start_function"`
}

type PlugFileRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}

type PlugResponse struct {
	Status bool `json:"status"`
}

func registerBlobWithName(baseURL string, name string, contentType string, fileName string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("cannot read file '%s'", fileName)
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(contentBytes)

	reqBody := &RegisterBlobWithNameRequest{
		Name:        name,
		ContentType: contentType,
		Bytes:       encodedBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal json")
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	resp, err := client.Post(baseURL+"/api/blob/register", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &RegisterBlobWithNameResponse{}
	if json.Unmarshal(bytes, response) != nil {
		return "", fmt.Errorf("cannot unmarshall server response : %s", string(bytes))
	}

	if response.Status {
		fmt.Printf("registered blob '%s' content_type:%s techID:%s\n", name, contentType, response.TechID)
	} else {
		return "", fmt.Errorf("ERROR while registration of '%s' content_type:%s techID:%s", name, contentType, response.TechID)
	}

	return response.TechID, nil
}

func registerBlob(baseURL string, contentType string, fileName string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("cannot read file '%s'", fileName)
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(contentBytes)

	reqBody := &RegisterBlobRequest{
		ContentType: contentType,
		Bytes:       encodedBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("cannot marshal json")
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	resp, err := client.Post(baseURL+"/api/blob/register", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &RegisterBlobResponse{}
	if json.Unmarshal(bytes, response) != nil {
		return "", fmt.Errorf("cannot unmarshall server response : %s", string(bytes))
	}

	if response.Status {
		fmt.Printf("registered blob content_type:%s techID:%s\n", contentType, response.TechID)
	} else {
		return "", fmt.Errorf("ERROR while registration of content_type:%s techID:%s", contentType, response.TechID)
	}

	return response.TechID, nil
}

func CliPushFunction(verbs []Verb) {
	// skip the verb that triggered us, it is given to us in case it contains options
	baseURL := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	functionName := verbs[0].Name
	wasmFileName := verbs[1].Name

	var codeType string
	if strings.HasSuffix(wasmFileName, ".wasm") {
		codeType = "application/wasm"
	} else if strings.HasSuffix(wasmFileName, ".js") {
		codeType = "text/javascript"
	} else {
		codeType = detectContentTypeFromFileName(wasmFileName)
	}

	registerBlobWithName(baseURL, functionName, codeType, wasmFileName)
}

func CliUploadFile(verbs []Verb) {
	baseURL := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	verbs = verbs[1:]

	path := verbs[0].Name
	fileName := verbs[1].Name
	contentType := ""
	if len(verbs) >= 3 {
		contentType = verbs[2].Name
	} else {
		contentType = detectContentTypeFromFileName(fileName)
	}

	techID, err := uploadFile(baseURL, method, path, contentType, fileName)
	if err != nil {
		fmt.Printf("error while doing things, %v\n", err)
		return
	}

	fmt.Printf("registered file '%s' '%s' content_type:%s techID:%s\n", fileName, path, contentType, techID)
}

func CliUploadDir(verbs []Verb) {
	baseURL := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	verbs = verbs[1:]

	pathPrefix := verbs[0].Name
	directoryName := verbs[1].Name

	directoryName, err := filepath.Abs(directoryName)
	if err != nil {
		fmt.Printf("error getting absolute path (%v)\n", err)
		return
	}

	count := 0
	countError := 0

	filepath.Walk(directoryName, func(path string, info os.FileInfo, err error) error {
		urlPath := filepath.Join(pathPrefix, path[len(directoryName):])
		if !info.IsDir() {
			fmt.Printf("%s  =>  %s\n", path, urlPath)
			_, err := uploadFile(baseURL, method, urlPath, detectContentTypeFromFileName(path), path)
			if err != nil {
				countError++
			}
			count++
		}
		return nil
	})

	fmt.Printf("uploaded %d files (%d errors).\n", count, countError)
}

func CliPlugFunction(verbs []Verb) {
	serverBaseUrl := getAPIBaseURL(verbs[0])
	method := verbs[0].GetOptionOr("method", "get")
	verbs = verbs[1:]

	path := verbs[0].Name
	name := verbs[1].Name
	startFunction := verbs[2].Name

	reqBody := &PlugFunctionRequest{
		Method:        method,
		Path:          path,
		Name:          name,
		StartFunction: startFunction,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json (%v)\n", err)
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	resp, err := client.Post(serverBaseUrl+"/api/function/plug", "application/json", bodyReader)
	if err != nil {
		fmt.Printf("error during http request (%v)\n", err)
		return
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &PlugResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if response.Status {
		fmt.Printf("ok, done\n")
	} else {
		fmt.Printf("error (response:%v)!\n", response)
	}
}

func detectContentTypeFromFileName(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return "application/octet-stream"
	}

	extension := name[i+1:]

	mimeType, ok := common.MimeTypes[extension]
	if !ok {
		return "application/octet-stream"
	}

	if strings.HasPrefix(mimeType, "text/") {
		mimeType = mimeType + "; charset=utf-8"
	}

	return mimeType
}

func getBaseUrl(verb Verb) string {
	defaultUrl, ok := os.LookupEnv("MYOWNCLUSTER_SERVER_BASE_URL")
	if !ok {
		defaultUrl = "https://localhost:8443"
	}

	return verb.GetOptionOr("baseUrl", defaultUrl)
}

func getAPIBaseURL(verb Verb) string {
	baseURL := getBaseUrl(verb)

	return baseURL + "/my-own-cluster"
}

func uploadFile(serverBaseUrl string, method string, path string, contentType string, fileName string) (string, error) {
	techID, err := registerBlob(serverBaseUrl, contentType, fileName)
	if err != nil {
		return "", fmt.Errorf("cannot read file '%s' (%v)", fileName, err)
	}

	reqBody := &PlugFileRequest{
		Method: method,
		Path:   path,
		Name:   fmt.Sprintf("techID://%s", techID),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json\n")
		return "", err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	resp, err := client.Post(serverBaseUrl+"/api/file/plug", "application/json", bodyReader)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &PlugResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return "", err
	}

	if !response.Status {
		return "", fmt.Errorf("bad response status")
	}

	return techID, nil
}

type CallFunctionRequest struct {
	Name           string    `json:"name"`
	StartFunction  *string   `json:"start_function,omitempty"`  // defaults to "_start"
	Arguments      *[]int    `json:"arguments,omitempty"`       // defaults to []int{}
	Mode           *string   `json:"mode,omitempty"`            // "direct" or "posix" (only for webassembly code, not applicable for javascript)
	Input          *string   `json:"input,omitempty"`           // defaults to empty buffer
	POSIXFilename  *string   `json:"posix_file_name,omitempty"` // defaults to "a.out"
	POSIXArguments *[]string `json:"posix_arguments,omitempty"` // defaults to []string{}
}

type CallFunctionResponse struct {
	Result int    `json:"result"`
	Output string `json:"output"`
	Error  bool   `json:"error"`
}

/*
func handlerCallFunction(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "cannot read your body")
		return
	}

	baseReq := CallFunctionRequest{}
	if json.Unmarshal(body, &baseReq) != nil {
		errorResponse(w, 400, "cannot read/parse your body")
		return
	}

	var input []byte
	if baseReq.Input != nil {
		input = []byte(*baseReq.Input)
	}

	startFunction := "_start"
	var arguments []int = []int{}

	mode := strings.ToLower(baseReq.Mode)

	switch mode {
	case "direct":
		bodyReq := DirectCallFunctionRequest{}
		if json.Unmarshal(body, &bodyReq) != nil {
			errorResponse(w, 400, "cannot read/parse your body for DIRECT mode")
			return
		}

		startFunction = bodyReq.StartFunction
		arguments = bodyReq.Arguments
		break

	case "posix":
		break

	default:
		errorResponse(w, 400, fmt.Sprintf("invalid execution mode '%s', aborting", mode))
		return
	}

	wasmBytes, ok := server.orchestrator.GetFunctionBytesByFunctionName(baseReq.Name)
	if !ok {
		errorResponse(w, 400, fmt.Sprintf("can't find sub function bytes (%s)\n", baseReq.Name))
		return
	}

	outputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()

	inputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()
	inputExchangeBuffer := server.orchestrator.GetExchangeBuffer(inputExchangeBufferID)
	inputExchangeBuffer.Write(input)

	fctx := &common.FunctionExecutionContext{
		Orchestrator:           server.orchestrator,
		Name:                   baseReq.Name,
		StartFunction:          startFunction,
		HasFinishedRunning:     false,
		InputExchangeBufferID:  inputExchangeBufferID,
		OutputExchangeBufferID: outputExchangeBufferID,
		Result:                 0,
	}

	wctx, err := wasm.PorcelainPrepareWasm(fctx, mode, wasmBytes)
	if err != nil {
		errorResponse(w, 404, fmt.Sprintf("cannot create function: %v", err))
		return
	}

	if mode == "posix" {
		bodyReq := WASICallFunctionRequest{}
		if json.Unmarshal(body, &bodyReq) != nil {
			errorResponse(w, 400, "cannot read/parse your body POSIX mode")
			return
		}

		wasm.PorcelainAddWASIPlugin(wctx, bodyReq.WasiFilename, bodyReq.Arguments)
	}

	wctx.Run(arguments)

	// as seen in the previous comment, here we will have to wait for the output buffer to be released by
	// all components before returning the http response. If the buffer is not touched, we will respond
	// with some user well known 5xx code.
	// That's a kind of distributed GC for buffers...
	outputExchangeBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(wctx.Fctx.OutputExchangeBufferID).GetBuffer()

	jsonResponse(w, 200, CallFunctionResponse{
		Result: wctx.Fctx.Result,
		Output: base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(outputExchangeBuffer),
		Error:  false,
	})
}*/

/*
func CliCallFunction(verbs []Verb) {
	// skip the verb that triggered us, it is given to us in case it contains options
	baseUrl := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	functionName := verbs[0].Name
	mode := strings.ToLower(verbs[0].GetOptionOr("mode", "direct"))
	input := verbs[0].GetOptionOr("input", "")

	var bodyReq interface{} = nil

	switch mode {
	case "posix":
		arguments := []string{}
		for a := 1; a < len(verbs); a++ {
			arguments = append(arguments, verbs[a].Name)
		}

		bodyReq = &WASICallFunctionRequest{
			CallFunctionRequest: CallFunctionRequest{
				Name:  functionName,
				Mode:  mode,
				Input: &input,
			},
			WasiFilename: verbs[0].GetOptionOr("wasi_file_name", functionName),
			Arguments:    arguments,
		}
		break

	case "direct":
		arguments := []int{}
		for a := 1; a < len(verbs); a++ {
			an, err := strconv.Atoi(verbs[a].Name)
			if err != nil {
				fmt.Println("Bad argument !", verbs[a].Name)
				return
			}
			arguments = append(arguments, an)
		}

		bodyReq = &DirectCallFunctionRequest{
			CallFunctionRequest: CallFunctionRequest{
				Name:  functionName,
				Mode:  mode,
				Input: &input,
			},
			Arguments:     arguments,
			StartFunction: verbs[0].GetOptionOr("start_function", "_start"),
		}
		break

	default:
		fmt.Printf("unkwown run mode '%s'\n", mode)
		return
	}

	bodyBytes, err := json.Marshal(bodyReq)
	if err != nil {
		fmt.Printf("cannot marshal json\n")
		return
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	resp, err := client.Post(baseUrl+"/api/function/call", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	fmt.Print(string(bytes))
}
*/
