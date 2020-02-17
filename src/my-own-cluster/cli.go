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
	"strconv"
	"strings"
)

func CliPushFunction(verbs []Verb) {
	// skip the verb that triggered us, it is given to us in case it contains options
	baseUrl := getAPIBaseURL(verbs[0])
	verbs = verbs[1:]

	functionName := verbs[0].Name
	wasmFileName := verbs[1].Name

	wasmBytes, err := ioutil.ReadFile(wasmFileName)
	if err != nil {
		fmt.Printf("cannot read file '%s'\n", wasmFileName)
		return
	}

	encodedWasmBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(wasmBytes)

	reqBody := &RegisterFunctionRequest{
		Name:      functionName,
		WasmBytes: encodedWasmBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
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
	resp, err := client.Post(baseUrl+"/api/function/register", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &RegisterFunctionResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if response.Status {
		fmt.Printf("registered function '%s' size:%d techID:%s\n", response.Name, response.WasmBytesSize, response.TechID)
	} else {
		fmt.Printf("ERROR while registration of '%s' size:%d techID:%s\n", response.Name, response.WasmBytesSize, response.TechID)
	}
}

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
		defaultUrl = "https://localhost:8444"
	}

	return verb.GetOptionOr("baseUrl", defaultUrl)
}

func getAPIBaseURL(verb Verb) string {
	baseURL := getBaseUrl(verb)

	return baseURL + "/my-own-cluster"
}

func CliUploadDir(verbs []Verb) {
	baseUrl := getAPIBaseURL(verbs[0])
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
			resp, err := uploadFile(baseUrl, method, urlPath, detectContentTypeFromFileName(path), path)
			if err != nil || !resp.Status {
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

	response := &PlugFunctionResponse{}
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

func CliUploadFile(verbs []Verb) {
	baseUrl := getAPIBaseURL(verbs[0])
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

	response, err := uploadFile(baseUrl, method, path, contentType, fileName)
	if err != nil {
		fmt.Printf("error while doing things, %v\n", err)
		return
	}

	if response.Status {
		fmt.Printf("registered file '%s' '%s' size:%d content_type:%s techID:%s\n", fileName, response.Path, response.BytesSize, response.ContentType, response.TechID)
	} else {
		fmt.Printf("ERROR while registration '%s' of '%s' size:%d content_type:%s techID:%s\n", fileName, response.Path, response.BytesSize, response.ContentType, response.TechID)
	}
}

func uploadFile(serverBaseUrl string, method string, path string, contentType string, fileName string) (*PlugFileResponse, error) {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("cannot read file '%s'\n", fileName)
		return nil, err
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(fileBytes)

	reqBody := &PlugFileRequest{
		Method:      method,
		Path:        path,
		ContentType: contentType,
		Bytes:       encodedBytes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("cannot marshal json\n")
		return nil, err
	}

	bodyReader := bytes.NewReader(bodyBytes)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	resp, err := client.Post(serverBaseUrl+"/api/file/register", "application/json", bodyReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &PlugFileResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return nil, err
	}

	return response, nil
}
