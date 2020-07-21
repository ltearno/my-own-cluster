package common

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
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
	request    *http.Request
	headers    map[string]string
	headersSet bool
}

func newRequestWrapper(method string, url string, headers map[string]string) *requestWrapper {
	w := &requestWrapper{
		headers:    make(map[string]string),
		headersSet: false,
	}

	req, err := http.NewRequest(method, url, w)
	if err != nil {
		return nil
	}

	w.request = req

	for k, v := range headers {
		w.headers[k] = v
	}

	return w
}

// called by http client when sending the request body
func (w *requestWrapper) Read(buffer []byte) (int, error) {
	fmt.Printf("ADHGKADJHAGDKJHADGKJHADG THE THING CALLED OUR READ WRAPPER")
	return 0, io.EOF
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
	w.ensureHeadersSet()
	return len(buffer), nil
}

func (w *requestWrapper) Close() int {
	return -1
}

func (w *requestWrapper) ensureHeadersSet() {
	if !w.headersSet {
		w.headersSet = true

		for k, v := range w.headers {
			w.request.Header.Set(k, v)
		}
	}
}

/*
	to receive data from the remote server
*/
type responseWrapper struct {
	headers map[string]string
}

func (o *Orchestrator) CreateExchangeBuffersFromHttpClientRequest(method string, url string, headers map[string]string) {
	rw := newRequestWrapper(method, url, headers)
	if rw == nil {
		return
	}

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	client.Do(rw.request)
}
