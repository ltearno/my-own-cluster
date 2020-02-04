package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

const BASE_SERVER_URL = "https://localhost:8443"

func main() {
	var printUsage = false
	var help = flag.Bool("help", false, "show this help")
	var port = flag.Int("port", 8443, "webserver listening port")

	flag.Parse()

	if *help {
		printUsage = true
	}

	verbs := flag.Args()

	if len(verbs) == 0 {
		fmt.Println("not enough parameters, use '-help' !")
		printUsage = true
	}

	printHelp := func() {
		fmt.Printf("\nmy-own-cluster usage :\n\n  my-own-cluster [OPTIONS] verbs...\n\nOPTIONS :\n\n")
		flag.PrintDefaults()
		fmt.Printf("\nVERBS :\n\n")
		fmt.Printf("  serve                   start the web server\n")
		fmt.Printf("  push NAME WASM_FILE     push a was code to the server\n")
	}

	if printUsage {
		printHelp()
		return
	}

	// execute the verb
	switch verbs[0] {
	case "serve":
		relativeWorkdir := "."
		if len(verbs) > 1 {
			relativeWorkdir = verbs[1]
		}

		workingDir, err := filepath.Abs(relativeWorkdir)
		if err != nil {
			fmt.Printf("cannot find working directory, abort (%v)\n", err)
			return
		}

		orchestrator := NewOrchestrator()

		StartWebServer(*port, workingDir, orchestrator)
		break

	case "push":
		if len(verbs) != 3 {
			printHelp()
			return
		}

		functionName := verbs[1]
		wasmFileName := verbs[2]

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

		break

	default:
		printHelp()
	}
}
