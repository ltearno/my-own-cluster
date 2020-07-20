package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"my-own-cluster/common"
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
		httpResponse(w, code, "application/json", "{ \"message\": \"error 98AAGER\" }")
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

func handlerGetGeneric(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	path := p.ByName("path")

	found, plugType, plug, boundParameters := server.orchestrator.GetPlugFromPath(r.Method, path)
	if !found {
		if server.trace {
			fmt.Printf("received not found query for path '%s'\n", path)
		}

		errorResponse(w, 404, fmt.Sprintf("sorry, unbound resource '%s', method:%s", path, r.Method))
		return
	}

	switch plugType {
	case "function":
		pluggedFunction := plug.(*common.PluggedFunction)

		if server.trace {
			fmt.Printf("received plugged function request, path:'%s', type:%s, name:%s, start_function:%s\n", path, plugType, pluggedFunction.Name, pluggedFunction.StartFunction)
		}

		// create exchange buffers and provide informations about current http request
		outputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()
		inputExchangeBufferID := server.orchestrator.CreateExchangeBuffer()
		inputExchangeBuffer := server.orchestrator.GetExchangeBuffer(inputExchangeBufferID)

		for k, v := range r.Header {
			// TODO why not support multiple values ? would add complexity and one header with clear syntax parsing should be enough
			inputExchangeBuffer.SetHeader(strings.ToLower(k), v[0])
		}
		inputExchangeBuffer.SetHeader("x-moc-host", r.Host)
		inputExchangeBuffer.SetHeader("x-moc-method", r.Method)
		inputExchangeBuffer.SetHeader("x-moc-proto", r.Proto)
		inputExchangeBuffer.SetHeader("x-moc-remote-addr", r.RemoteAddr)
		inputExchangeBuffer.SetHeader("x-moc-request-uri", r.RequestURI)
		inputExchangeBuffer.SetHeader("x-moc-url-path", r.URL.Path)
		inputExchangeBuffer.SetHeader("x-moc-url-query", r.URL.RawQuery)
		for k, v := range boundParameters {
			inputExchangeBuffer.SetHeader(fmt.Sprintf("x-moc-path-param-%s", strings.ToLower(k)), v)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			inputExchangeBuffer.Write(body)
		} else {
			fmt.Printf("CANNOT READ BODY\n")
		}

		// create a function execution context ...
		fctx := server.orchestrator.NewFunctionExecutionContext(
			pluggedFunction.Name,
			pluggedFunction.StartFunction,
			[]int{},
			server.trace,
			"direct",
			nil,
			nil,
			inputExchangeBufferID,
			outputExchangeBufferID,
		)

		// ... and run it
		err = fctx.Run()
		if err != nil {
			errorResponse(w, 500, fmt.Sprintf("error while executing the function: %v", err))
			return
		}

		// copy output exchange buffer headers to the response headers
		outputExchangeBuffer := server.orchestrator.GetExchangeBuffer(outputExchangeBufferID)
		outputExchangeBuffer.GetHeaders(func(name string, value string) {
			w.Header().Set(name, value)
		})
		w.WriteHeader(fctx.Result)

		// copy output exchange buffer content to response body
		w.Write(outputExchangeBuffer.GetBuffer())

		// release exchange buffers
		server.orchestrator.ReleaseExchangeBuffer(inputExchangeBufferID)
		server.orchestrator.ReleaseExchangeBuffer(outputExchangeBufferID)

		return

	case "file":
		pluggedFile := plug.(*common.PluggedFile)

		if server.trace {
			fmt.Printf("received plugged file request, path:'%s', type:%s, name:%s\n", path, plugType, pluggedFile.Name)
		}

		fileTechID, err := server.orchestrator.GetBlobTechIDFromReference(pluggedFile.Name)
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
		contentType := fileAbstract.ContentType
		if strings.HasPrefix(contentType, "text/") {
			contentType = contentType + "; charset=utf-8"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(200)
		w.Write(fileBytes)

		return
	}

	errorResponse(w, 404, "sorry, nothing found")
	return
}

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
