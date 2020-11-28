package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"my-own-cluster/common"
	"my-own-cluster/tools"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{}

type WebServer struct {
	name         string
	orchestrator *common.Orchestrator
	trace        bool
}

func (server *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.orchestrator.StatIncrement(common.STAT_NB_REQUESTS_RECEIVED)

	path := r.URL.Path
	method := strings.ToLower(r.Method)

	server.orchestrator.StatIncrement(common.StatName(fmt.Sprintf("hit_count_%s_%s", method, path)))

	if server.trace {
		fmt.Printf("WEB HANDLER METHOD='%s' PATH='%s'\n", method, path)
	}

	found, plugType, plug, boundParameters := server.orchestrator.GetPlugFromPath(r.Method, path)
	if !found {
		if server.trace {
			fmt.Printf("received not found query for path '%s'\n", path)
		}

		errorResponse(w, 404, fmt.Sprintf("sorry, unbound resource '%s', method:%s", path, r.Method))
		return
	}

	var outputExchangeBufferID int
	var inputExchangeBufferID int

	if r.Header.Get("Upgrade") == "websocket" {
		fmt.Printf("WE ARE ON A WEBSOCKET CONNECTION !!!\n")

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		inputExchangeBufferID, outputExchangeBufferID = server.orchestrator.CreateWrappedWebSocketExchangeBuffers(tools.SimplifyHeaders(r.Header), c)
	} else {
		// create exchange buffers
		outputExchangeBufferID = server.orchestrator.CreateWrappedHttpResponseWriterExchangeBuffer(w)
		inputExchangeBufferID = server.orchestrator.CreateWrappedHttpRequestExchangeBuffer(r)
	}

	// safe guard :  release exchange buffers
	defer server.orchestrator.ReleaseExchangeBuffer(inputExchangeBufferID)
	defer server.orchestrator.ReleaseExchangeBuffer(outputExchangeBufferID)

	// provide informations about current http request in the inputExchangeBuffer
	inputExchangeBuffer := server.orchestrator.GetExchangeBuffer(inputExchangeBufferID)
	for k, v := range r.Header {
		inputExchangeBuffer.SetHeader(strings.ToLower(k), v[0])
	}
	inputExchangeBuffer.SetHeader("x-moc-host", r.Host)
	inputExchangeBuffer.SetHeader("x-moc-method", method)
	inputExchangeBuffer.SetHeader("x-moc-proto", r.Proto)
	inputExchangeBuffer.SetHeader("x-moc-remote-addr", r.RemoteAddr)
	inputExchangeBuffer.SetHeader("x-moc-request-uri", r.RequestURI)
	inputExchangeBuffer.SetHeader("x-moc-url-path", r.URL.Path)
	inputExchangeBuffer.SetHeader("x-moc-url-query", r.URL.RawQuery)
	for k, v := range boundParameters {
		inputExchangeBuffer.SetHeader(fmt.Sprintf("x-moc-path-param-%s", strings.ToLower(k)), v)
	}

	filters := server.orchestrator.GetFilters()
	for _, filter := range filters {
		// create a function execution context ...
		fctx := server.orchestrator.NewFunctionExecutionContext(
			filter.Name,
			filter.StartFunction,
			[]int{},
			server.trace,
			"direct",
			nil,
			nil,
			inputExchangeBufferID,
			outputExchangeBufferID,
		)

		// ... and run it
		err := fctx.Run()
		if err != nil {
			errorResponse(w, 500, fmt.Sprintf("error while executing the function: '%v'", err))
			return
		}
	}

	switch plugType {
	case "function":
		pluggedFunction := plug.(*common.PluggedFunction)

		fmt.Printf("TAGS: %v\n", pluggedFunction.Tags)

		inputExchangeBuffer.SetHeader("x-moc-plug-data", pluggedFunction.Data)

		if server.trace {
			fmt.Printf("received plugged function request, path:'%s', type:%s, name:%s, start_function:%s\n", path, plugType, pluggedFunction.Name, pluggedFunction.StartFunction)
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
		err := fctx.Run()
		if err != nil {
			errorResponse(w, 500, fmt.Sprintf("error while executing the function: '%v'", err))
			return
		}

		return

	case "file":
		if method != "get" {
			errorResponse(w, 404, "sorry, nothing found.")
			return
		}

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

		outputExchangeBuffer := server.orchestrator.GetExchangeBuffer(outputExchangeBufferID)

		outputExchangeBuffer.SetHeader("Content-Type", contentType)
		outputExchangeBuffer.WriteStatusCode(200)
		outputExchangeBuffer.Write(fileBytes)

		return
	}

	errorResponse(w, 404, "sorry, nothing found")
	return
}

// StartWebServer runs a webserver hosting the application
func StartWebServer(port int, workingDir string, orchestrator *common.Orchestrator, trace bool) {
	webServer := &WebServer{
		name:         "my-own-cluster",
		orchestrator: orchestrator,
		trace:        trace,
	}

	endSignal := make(chan bool, 1)

	go func() {
		address := fmt.Sprintf(":%d", port)
		certPath := filepath.Join(workingDir, "tls.cert.pem")
		keyPath := filepath.Join(workingDir, "tls.key.pem")

		server := &http.Server{
			Addr:    address,
			Handler: webServer,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
				MaxVersion:         tls.VersionTLS12,
			},
		}

		log.Fatal(server.ListenAndServeTLS(certPath, keyPath))

		endSignal <- true
	}()

	fmt.Printf("listening on port %d\n", port)

	<-endSignal
}
