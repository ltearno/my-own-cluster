package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"my-own-cluster/common"
	"my-own-cluster/js"
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

func getTechIDFromPlugName(o *common.Orchestrator, name string) (string, error) {
	if strings.HasPrefix(name, "techID://") {
		return name[len("techID://"):], nil
	}

	techID, err := o.GetBlobTechIDFromName(name)
	if err != nil {
		return "", err
	}

	return techID, nil
}

func handlerGetGeneric(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	path := p.ByName("path")

	found, plugType, plug, boundParameters := server.orchestrator.GetPlugFromPath(r.Method, path)
	if !found {
		errorResponse(w, 404, fmt.Sprintf("sorry, unbound resource '%s', method:%s", path, r.Method))
		return
	}

	switch plugType {
	case "function":
		pluggedFunction := plug.(*common.PluggedFunction)

		pluggedFunctionTechID, err := getTechIDFromPlugName(server.orchestrator, pluggedFunction.Name)
		if err != nil {
			errorResponse(w, 400, fmt.Sprintf("can't find plugged function (%s)\n", pluggedFunction.Name))
			return
		}

		pluggedFunctionAbstract, err := server.orchestrator.GetBlobAbstractByTechID(pluggedFunctionTechID)
		if err != nil {
			errorResponse(w, 400, fmt.Sprintf("can't find plugged function abstract (%s)\n", pluggedFunction.Name))
			return
		}

		if pluggedFunctionAbstract.ContentType != "application/wasm" && pluggedFunctionAbstract.ContentType != "text/javascript" {
			errorResponse(w, 400, fmt.Sprintf("not supported function code type '%s'", pluggedFunctionAbstract.ContentType))
			return
		}

		wasmBytes, err := server.orchestrator.GetBlobBytesByTechID(pluggedFunctionTechID)
		if err != nil {
			errorResponse(w, 400, fmt.Sprintf("can't find plugged function bytes (%s)\n", pluggedFunction.Name))
			return
		}

		outputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()
		inputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()
		inputExchangeBuffer := server.orchestrator.GetExchangeBuffer(inputExchangeBufferID)
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			inputExchangeBuffer.Write(body)
		} else {
			fmt.Printf("CANNOT READ BODY\n")
		}
		for k, v := range r.Header {
			// TODO why not support multiple values ?
			inputExchangeBuffer.SetHeader(k, v[0])
		}
		inputExchangeBuffer.SetHeader("x-moc-host", r.Host)
		inputExchangeBuffer.SetHeader("x-moc-method", r.Method)
		inputExchangeBuffer.SetHeader("x-moc-proto", r.Proto)
		inputExchangeBuffer.SetHeader("x-moc-remote-addr", r.RemoteAddr)
		inputExchangeBuffer.SetHeader("x-moc-request-uri", r.RequestURI)
		inputExchangeBuffer.SetHeader("x-moc-url-path", r.URL.Path)
		for k, v := range boundParameters {
			inputExchangeBuffer.SetHeader(fmt.Sprintf("x-moc-path-param-%s", k), v)
		}

		mode := "direct"
		arguments := []int{}
		posixFileName := "a.out"
		posixArguments := []string{"42", "33"}

		fctx := &common.FunctionExecutionContext{
			Orchestrator:  server.orchestrator,
			Name:          pluggedFunction.Name,
			StartFunction: pluggedFunction.StartFunction,
			Trace:         server.trace,

			HasFinishedRunning:     false,
			InputExchangeBufferID:  inputExchangeBufferID,
			OutputExchangeBufferID: outputExchangeBufferID,
			Result:                 -1,
		}

		switch pluggedFunctionAbstract.ContentType {
		case "application/wasm":
			wctx, err := wasm.PorcelainPrepareWasm(fctx, mode, wasmBytes)
			if err != nil || wctx == nil {
				errorResponse(w, 404, fmt.Sprintf("cannot create wasm context for function: %v", err))
				return
			}

			if mode == "posix" {
				wctx.AddAPIPlugin(wasm.NewWASIHostPlugin(posixFileName, posixArguments, map[int]wasm.VirtualFile{
					0: wasm.CreateStdInVirtualFile(wctx, inputExchangeBuffer.GetBuffer()),
					1: wctx.Fctx.Orchestrator.GetExchangeBuffer(wctx.Fctx.OutputExchangeBufferID),
					2: wasm.CreateStdErrVirtualFile(wctx),
				}))
			}

			err = wctx.Run(arguments)
			if err != nil {
				errorResponse(w, 500, fmt.Sprintf("execution error in function: %v", err))
				return
			}

			break

		case "text/javascript":
			jsctx, err := js.PorcelainPrepareJs(fctx, wasmBytes)
			if err != nil || jsctx == nil {
				errorResponse(w, 404, fmt.Sprintf("cannot create js context for function: %v", err))
				return
			}

			err = jsctx.Run(arguments)
			if err != nil {
				errorResponse(w, 500, fmt.Sprintf("execution error in function: %v", err))
				return
			}

			break

		default:
			errorResponse(w, 400, fmt.Sprintf("unknown function type '%s'\n", pluggedFunctionAbstract.ContentType))
			return
		}

		if fctx.Trace {
			fmt.Printf(" -> result:%d\n", fctx.Result)
		}

		/*
			Instead of waiting for the end of the call, we should count references to the exchange buffer
			and wait for the last reference to dissappear. At this moment, the http response is complete and
			can be sent back to the client. This allows the first callee to transfer its output exchange
			buffer to another function and exit. The other function will then do whatever it wants to do
			(fan out, fan in and so on...).
		*/
		// here we will have to wait for the output buffer to be released by
		// all components before returning the http response. If the buffer is not touched, we will respond
		// with some user well known 5xx code.
		// That's a kind of distributed GC for buffers...
		outputExchangeBuffer := server.orchestrator.GetExchangeBuffer(outputExchangeBufferID)

		// copy output exchange buffer headers to the response headers
		outputExchangeBuffer.GetHeaders(func(name string, value string) {
			w.Header().Set(name, value)
		})
		w.WriteHeader(200)

		// copy output exchange buffer content to response body
		w.Write(outputExchangeBuffer.GetBuffer())

		server.orchestrator.ReleaseExchangeBuffer(inputExchangeBufferID)
		server.orchestrator.ReleaseExchangeBuffer(outputExchangeBufferID)

		return

	case "file":
		pluggedFile := plug.(*common.PluggedFile)

		fileTechID, err := getTechIDFromPlugName(server.orchestrator, pluggedFile.Name)
		if err != nil {
			errorResponse(w, 404, fmt.Sprintf("sorry, file techID '%s' not found", fileTechID))
			return
		}

		fileAbstract, err := server.orchestrator.GetBlobAbstractByTechID(fileTechID)
		if err != nil {
			errorResponse(w, 404, "sorry, file abstract type not found")
			return
		}

		fileBytes, err := server.orchestrator.GetBlobBytesByTechID(fileTechID)
		if err != nil {
			errorResponse(w, 404, "sorry, file bytes not found")
			return
		}

		// TODO add the ETag header corresponding to the sha
		w.Header().Set("Content-Type", fileAbstract.ContentType)
		w.WriteHeader(200)
		w.Write(fileBytes)

		return
	}

	errorResponse(w, 404, "sorry, nothing found")
	return
}

/*
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

// injects the WebServer context in http-router handler
func (server *WebServer) makeHandler(handler func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p, server)
	}
}

func initRouting(server *WebServer) {
	server.router.GET("/*path", server.makeHandler(handlerGetGeneric))
	server.router.POST("/*path", server.makeHandler(handlerGetGeneric))
}

type WebServer struct {
	name         string
	orchestrator *common.Orchestrator
	trace        bool
	router       *httprouter.Router
}

// StartWebServer runs a webserver hosting the application
func StartWebServer(port int, workingDir string, orchestrator *common.Orchestrator, trace bool) {
	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := &WebServer{
		name:         "my-own-cluster",
		orchestrator: orchestrator,
		trace:        trace,
		router:       router,
	}

	initRouting(server)

	endSignal := make(chan bool, 1)

	go func() {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", port), filepath.Join(workingDir, "tls.cert.pem"), filepath.Join(workingDir, "tls.key.pem"), router))
		endSignal <- true
	}()

	fmt.Printf("listening on port %d\n", port)

	<-endSignal
}
