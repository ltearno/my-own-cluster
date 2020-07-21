package common

import (
	"crypto/tls"
	"fmt"
	"io"
	"my-own-cluster/tools"
	"net/http"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/gorilla/websocket"
)

/*
req, err := NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)*/

/*
	to send data to the remote server
*/
type requestWrapper struct {
	request *http.Request
	headers map[string]string

	queue          *queue.Queue
	availableBytes []byte
}

func newRequestWrapper(method string, url string, headers map[string]string) (*requestWrapper, error) {
	w := &requestWrapper{
		headers: make(map[string]string),
		queue:   queue.New(1),
	}

	req, err := http.NewRequest(method, url, w)
	if err != nil {
		return nil, err
	}

	w.request = req

	for k, v := range headers {
		w.headers[k] = v
		req.Header.Set(k, v)
	}

	return w, nil
}

// called by http client when sending the request body
func (w *requestWrapper) Read(buffer []byte) (int, error) {
	//fmt.Printf("rw read\n")
	for {
		if w.availableBytes != nil {
			// first purge available bytes
			toGive := tools.Min(len(buffer), len(w.availableBytes))
			copy(buffer[0:toGive], w.availableBytes[0:toGive])
			w.availableBytes = w.availableBytes[toGive:]
			if len(w.availableBytes) == 0 {
				w.availableBytes = nil
			}
			//fmt.Printf("rw give %d bytes\n", toGive)
			return toGive, nil
		} else if w.queue.Disposed() {
			// if finished, say it
			//fmt.Printf("rw says EOF\n")
			return 0, io.EOF
		} else {
			// or else wait for next available bytes
			//fmt.Printf("rw waits on queue\n")
			b, _ := w.queue.Get(1)
			//fmt.Printf("rw wait is finished with %v\n", b)
			if b != nil && len(b) > 0 && len(b[0].([]byte)) > 0 {
				w.availableBytes = b[0].([]byte)
			}
		}
	}
}

func (w *requestWrapper) GetHeader(name string) (string, bool) {
	v, ok := w.headers[name]
	return v, ok
}

func (w *requestWrapper) GetHeadersCount() int {
	return len(w.headers)
}

func (w *requestWrapper) GetHeaders(cb func(name string, value string)) {
	for k, v := range w.headers {
		cb(k, v)
	}
}

func (w *requestWrapper) GetStatusCode() int {
	fmt.Printf("err 93876KHJG\n")
	return -1
}

func (w *requestWrapper) GetBuffer() []byte {
	fmt.Printf("err JKHZ?NBV\n")
	return nil
}

func (w *requestWrapper) SetHeader(name string, value string) {
	w.headers[name] = value
}

func (w *requestWrapper) WriteStatusCode(statusCode int) {
	fmt.Printf("err KJKENEBKJ\n")
}

func (w *requestWrapper) Write(buffer []byte) (int, error) {
	o := make([]byte, len(buffer))
	copy(o, buffer)
	w.queue.Put(o)
	return len(buffer), nil
}

func (w *requestWrapper) Close() int {
	w.queue.Dispose()
	return 0
}

/*
	to receive data from the remote server
*/
type responseWrapper struct {
	response *http.Response
	headers  map[string]string
}

func newResponseWrapper(r *http.Response) (*responseWrapper, error) {
	w := &responseWrapper{
		headers:  make(map[string]string),
		response: r,
	}

	for k, v := range r.Header {
		if len(v) != 1 {
			fmt.Printf("WARNING : header '%s' has %d values, skipping all but the first...\n", k, len(v))
		}

		if len(v) > 0 {
			w.headers[k] = v[0]
		}
	}

	return w, nil
}

func (w *responseWrapper) GetHeader(name string) (string, bool) {
	h, ok := w.headers[name]
	return h, ok
}

func (w *responseWrapper) GetHeadersCount() int {
	return len(w.headers)
}

func (w *responseWrapper) GetHeaders(cb func(name string, value string)) {
	for k, v := range w.headers {
		cb(k, v)
	}
}

func (w *responseWrapper) GetStatusCode() int {
	return w.response.StatusCode
}

func (w *responseWrapper) GetBuffer() []byte {
	buffer := make([]byte, 1024*1024)
	n, err := w.response.Body.Read(buffer)
	if n > 0 {
		return buffer[0:n]
	}
	if err != nil {
		//fmt.Printf("Error while reading http response '%v'\n", err)
	}
	return nil
}

func (w *responseWrapper) SetHeader(name string, value string) {
	w.headers[name] = value
}

func (w *responseWrapper) WriteStatusCode(statusCode int) {
	fmt.Printf("ERROR cannot call WriteStatusCode on client response wrapper\n")
}

func (w *responseWrapper) Write(buffer []byte) (int, error) {
	return -1, fmt.Errorf("ERROR cannot call WriteStatusCode on client response wrapper")
}

func (w *responseWrapper) Close() int {
	w.response.Body.Close()
	return 0
}

func (o *Orchestrator) CreateExchangeBuffersFromHttpClientRequest(method string, url string, headers map[string]string) (int, int, error) {
	requestWrapper, err := newRequestWrapper(method, url, headers)
	if err != nil {
		return -1, -1, err
	}

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	response, err := client.Do(requestWrapper.request)
	if err != nil {
		return -1, -1, err
	}

	responseWrapper, err := newResponseWrapper(response)
	if err != nil {
		return -1, -1, err
	}

	return o.RegisterExchangeBuffer(requestWrapper), o.RegisterExchangeBuffer(responseWrapper), nil
}

func (o *Orchestrator) CreateExchangeBuffersFromWebSocketClient(method string, url string, headers map[string]string) (int, int, error) {
	h := make(map[string][]string)
	for k, v := range headers {
		h[k] = []string{v}
	}

	//websocket.DefaultDialer.TLSClientConfig.InsecureSkipVerify = true
	/*trace := &httptrace.ClientTrace{
		WroteHeaderField: func(k string, v []string) { fmt.Printf("trace set header %s: %s\n", k, v[0]) },
		WroteRequest:     func(i httptrace.WroteRequestInfo) { fmt.Printf("trace sent req %v\n", i.Err) },
	}
	ctx := httptrace.WithClientTrace(context.Background(), trace)*/

	con, response, err := websocket.DefaultDialer.Dial(url, h)
	if err != nil {
		return -1, -1, err
	}

	requestWrapper := WrapWebSocketAsExchangeBuffer(headers, con)
	responseWrapper := WrapWebSocketAsExchangeBuffer(tools.SimplifyHeaders(response.Header), con)

	return o.RegisterExchangeBuffer(requestWrapper), o.RegisterExchangeBuffer(responseWrapper), nil
}
