package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func CliCallFunction(verbs []string) {
	functionName := verbs[0]
	mode := strings.ToLower(verbs[1])

	var bodyReq interface{} = nil

	switch mode {
	case "posix":
		bodyReq = &WASICallFunctionRequest{
			CallFunctionRequest: CallFunctionRequest{
				Name:  functionName,
				Mode:  mode,
				Input: nil,
			},
			WasiFilename: "toto",
			Arguments:    []string{"kjhgkjhg"},
		}
		break

	case "direct":
		bodyReq = &DirectCallFunctionRequest{
			CallFunctionRequest: CallFunctionRequest{
				Name:  functionName,
				Mode:  mode,
				Input: nil,
			},
			Arguments:     []int{2, 66},
			StartFunction: "_start",
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
	resp, err := client.Post(BASE_SERVER_URL+"/api/functions/call", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	fmt.Print(string(bytes))
}

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
	resp, err := client.Post(BASE_SERVER_URL+"/api/functions/register", "application/json", bodyReader)
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
		fmt.Printf("registered function '%s' size:%d techID:%s\n", response.Name, response.WasmBytesSize, response.TechId)
	} else {
		fmt.Printf("ERROR while registration of '%s' size:%d techID:%s\n", response.Name, response.WasmBytesSize, response.TechId)
	}
}
