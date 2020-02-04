package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RegisterFunctionRequest struct {
	Name      string `json:"name"`
	WasmBytes string `json:"wasm_bytes"`
}

type RegisterFunctionResponse struct {
	Status        bool   `json:"status"`
	TechId        string `json:"tech_id"`
	Name          string `json:"name"`
	WasmBytesSize int    `json:"wasm_bytes_size"`
}

func httpResponse(w http.ResponseWriter, code int, contentType string, body string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func unauthenticatedResponse(w http.ResponseWriter) {
	w.WriteHeader(401)
	w.Write(nil)
}

func unauthorizedResponse(w http.ResponseWriter) {
	w.WriteHeader(403)
	w.Write(nil)
}

func redirectResponse(w http.ResponseWriter, location string) {
	w.Header().Set("Location", location)
	w.WriteHeader(301)
	w.Write(nil)
}

func jsonResponse(w http.ResponseWriter, code int, value interface{}) {
	body, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(w, "{ \"message\": \"error 98AAGGD\" }")
		return
	}

	httpResponse(w, code, "application/json", string(body))
}

func messageResponse(w http.ResponseWriter, message string) {
	jsonResponse(w, 200, MessageResponse{message})
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	jsonResponse(w, code, ErrorResponse{message})
}

func extractBodyAsJSON(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("CANNOT READ BODY\n")
		return err
	}

	return json.Unmarshal(body, v)
}

func handlerRegisterFunction(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	bodyReq := RegisterFunctionRequest{}
	err := extractBodyAsJSON(r, &bodyReq)
	if err != nil {
		fmt.Println(err)
		errorResponse(w, 500, "cannot read/parse your body")
		return
	}

	wasmBytes, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(bodyReq.WasmBytes)
	if err != nil {
		panic(err)
	}

	techID := server.orchestrator.RegisterFunction(bodyReq.Name, wasmBytes)

	response := RegisterFunctionResponse{
		Status:        true,
		TechId:        techID,
		Name:          bodyReq.Name,
		WasmBytesSize: len(wasmBytes),
	}

	jsonResponse(w, 200, response)
}

type CallFunctionRequest struct {
	Name string `json:"name"`
	// 'direct' or 'posix'
	Mode  string  `json:"mode"`
	Input *string `json:"input,omitempty"`
	Trace bool    `json:"trace,omitempty"`
}

type WASICallFunctionRequest struct {
	WasiFilename string   `json:"wasi_file_name"`
	Arguments    []string `json:"arguments"`
}

type DirectCallFunctionRequest struct {
	StartFunction string `json:"start_function"`
	Arguments     []int  `json:"arguments"`
}

type CallFunctionResponse struct {
	Result int    `json:"result"`
	Output string `json:"output"`
	Error  bool   `json:"error"`
}

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

	if mode == "direct" {
		bodyReq := DirectCallFunctionRequest{}
		if json.Unmarshal(body, &bodyReq) != nil {
			errorResponse(w, 400, "cannot read/parse your body for DIRECT mode")
			return
		}

		startFunction = bodyReq.StartFunction
		arguments = bodyReq.Arguments
	}

	portID := server.orchestrator.CreateOutputPort()
	wctx := CreateWasmContext(server.orchestrator, mode, baseReq.Name, startFunction, input, portID)
	if wctx == nil {
		errorResponse(w, 404, "not found function, maybe you forgot to register it ?")
		return
	}

	wctx.Trace = baseReq.Trace

	wctx.AddAPIPlugin(NewMyOwnClusterAPIPlugin())
	wctx.AddAPIPlugin(NewTinyGoAPIPlugin())

	switch mode {
	case "direct":
		break

	case "posix":
		{
			bodyReq := WASICallFunctionRequest{}
			if json.Unmarshal(body, &bodyReq) != nil {
				errorResponse(w, 400, "cannot read/parse your body POSIX mode")
				return
			}

			wctx.AddAPIPlugin(NewWASIHostPlugin(bodyReq.WasiFilename, bodyReq.Arguments, map[int]VirtualFile{
				0: CreateStdInVirtualFile(wctx, wctx.InputBuffer),
				1: wctx.Orchestrator.GetOutputPort(wctx.OutputPortID), // CreateStdOutVirtualFile(wctx)
				2: CreateStdErrVirtualFile(wctx),
			}))
			break
		}
	default:
		errorResponse(w, 400, "execution mode not specified, aborting")
		return
	}

	wctx.Run(arguments)

	outputBuffer := wctx.Orchestrator.GetOutputPort(wctx.OutputPortID).GetBuffer()
	outputString := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(outputBuffer)

	jsonResponse(w, 200, CallFunctionResponse{
		Result: wctx.Result,
		Output: outputString,
		Error:  false,
	})
}

// injects the WebServer context in http-router handler
func (server *WebServer) makeHandler(handler func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p, server)
	}
}

func (server *WebServer) init(router *httprouter.Router) {
	router.POST("/api/functions/register", server.makeHandler(handlerRegisterFunction))
	router.POST("/api/functions/call", server.makeHandler(handlerCallFunction))
}

type WebServer struct {
	name         string
	orchestrator *Orchestrator
}

// Start runs a webserver hosting the application
func StartWebServer(port int, workingDir string, orchestrator *Orchestrator) {
	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := &WebServer{
		name:         "my-own-cluster",
		orchestrator: orchestrator,
	}

	server.init(router)

	fmt.Printf("listening on port %d\n", port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", port), filepath.Join(workingDir, "tls.cert.pem"), filepath.Join(workingDir, "tls.key.pem"), router))
}
