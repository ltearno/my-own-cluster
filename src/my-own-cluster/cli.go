package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func CliPushFunction(functionName string, wasmFileName string) {
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
	resp, err := client.Post(BASE_SERVER_URL+"/api/function/register", "application/json", bodyReader)
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
	resp, err := client.Post(BASE_SERVER_URL+"/api/function/call", "application/json", bodyReader)
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

	mimeType, ok := MimeTypes[extension]
	if !ok {
		return "application/octet-stream"
	}

	return mimeType
}

func CliUploadFile(verbs []Verb) {
	verbs = verbs[1:]

	path := verbs[0].Name
	fileName := verbs[1].Name
	contentType := ""
	if len(verbs) >= 3 {
		contentType = verbs[2].Name
	} else {
		contentType = detectContentTypeFromFileName(fileName)
	}

	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("cannot read file '%s'\n", fileName)
		return
	}

	encodedBytes := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(fileBytes)

	reqBody := &RegisterFileRequest{
		Path:        path,
		ContentType: contentType,
		Bytes:       encodedBytes,
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

	resp, err := client.Post(BASE_SERVER_URL+"/api/file/register", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	response := &RegisterFileResponse{}
	if json.Unmarshal(bytes, response) != nil {
		fmt.Printf("cannot unmarshall server response : %s\n", string(bytes))
		return
	}

	if response.Status {
		fmt.Printf("registered file '%s' '%s' size:%d content_type:%s techID:%s\n", fileName, response.Path, response.BytesSize, response.ContentType, response.TechID)
	} else {
		fmt.Printf("ERROR while registration '%s' of '%s' size:%d content_type:%s techID:%s\n", fileName, response.Path, response.BytesSize, response.ContentType, response.TechID)
	}
}
