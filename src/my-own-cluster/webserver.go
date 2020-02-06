package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"my-own-cluster/common"
	"my-own-cluster/wasm"
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
	TechID        string `json:"tech_id"`
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

/*-----------------------------------------------------------------------------

Register File

-----------------------------------------------------------------------------*/

type RegisterFileRequest struct {
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
	Bytes       string `json:"bytes"`
}

type RegisterFileResponse struct {
	Status      bool   `json:"status"`
	TechID      string `json:"tech_id"`
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
	BytesSize   int    `json:"bytes_size"`
}

func handlerGetFile(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	path := p.ByName("path")

	techID, present := server.orchestrator.GetFileTechIDFromPath(path)
	if !present {
		errorResponse(w, 404, "sorry, unbound resource")
		return
	}

	contentType, present := server.orchestrator.GetFileContentType(techID)
	if !present {
		errorResponse(w, 404, "sorry, file content type not found")
		return
	}

	fileBytes, present := server.orchestrator.GetFileBytes(techID)
	if !present {
		errorResponse(w, 404, "sorry, file bytes not found")
		return
	}

	// TODO add the ETag header corresponding to the sha
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(200)
	w.Write(fileBytes)
}

func handlerRegisterFile(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	bodyReq := RegisterFileRequest{}
	err := extractBodyAsJSON(r, &bodyReq)
	if err != nil {
		fmt.Println(err)
		errorResponse(w, 500, "cannot read/parse your body")
		return
	}

	bytes, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(bodyReq.Bytes)
	if err != nil {
		panic(err)
	}

	techID := server.orchestrator.RegisterFile(bodyReq.Path, bodyReq.ContentType, bytes)

	response := RegisterFileResponse{
		Status:      true,
		TechID:      techID,
		Path:        bodyReq.Path,
		ContentType: bodyReq.ContentType,
		BytesSize:   len(bytes),
	}

	jsonResponse(w, 200, response)
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
		TechID:        techID,
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
}

type WASICallFunctionRequest struct {
	CallFunctionRequest
	WasiFilename string   `json:"wasi_file_name"`
	Arguments    []string `json:"arguments"`
}

type DirectCallFunctionRequest struct {
	CallFunctionRequest
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

	wctx, err := wasm.PorcelainPrepareWasm(
		server.orchestrator,
		mode,
		baseReq.Name,
		startFunction,
		wasmBytes,
		input,
		outputExchangeBufferID,
		server.trace)
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

	outputExchangeBuffer := wctx.Orchestrator.GetExchangeBuffer(wctx.OutputExchangeBufferID).GetBuffer()

	jsonResponse(w, 200, CallFunctionResponse{
		Result: wctx.Result,
		Output: base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(outputExchangeBuffer),
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
	router.POST("/api/file/register", server.makeHandler(handlerRegisterFile))
	router.POST("/api/function/register", server.makeHandler(handlerRegisterFunction))
	router.POST("/api/function/call", server.makeHandler(handlerCallFunction))
	router.GET("/*path", server.makeHandler(handlerGetFile))
}

type WebServer struct {
	name         string
	orchestrator *common.Orchestrator
	trace        bool
}

// Start runs a webserver hosting the application
func StartWebServer(port int, workingDir string, orchestrator *common.Orchestrator, trace bool) {
	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := &WebServer{
		name:         "my-own-cluster",
		orchestrator: orchestrator,
		trace:        trace,
	}

	server.init(router)

	fmt.Printf("listening on port %d\n", port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", port), filepath.Join(workingDir, "tls.cert.pem"), filepath.Join(workingDir, "tls.key.pem"), router))
}
