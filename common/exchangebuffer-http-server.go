package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (o *Orchestrator) CreateWrappedHttpRequestExchangeBuffer(r *http.Request) int {
	return o.RegisterExchangeBuffer(WrapHttpReaderAsExchangeBuffer(r))
}

func (o *Orchestrator) CreateWrappedHttpResponseWriterExchangeBuffer(w http.ResponseWriter) int {
	return o.RegisterExchangeBuffer(WrapHttpWriterAsExchangeBuffer(w))
}

type HttpReaderExchangeBuffer struct {
	r           *http.Request
	headersRead bool
	headers     map[string]string
}

func WrapHttpReaderAsExchangeBuffer(r *http.Request) *HttpReaderExchangeBuffer {
	return &HttpReaderExchangeBuffer{
		r:           r,
		headersRead: false,
		headers:     make(map[string]string),
	}
}

func (b *HttpReaderExchangeBuffer) GetHeader(name string) (string, bool) {
	b.ensureHeadersReadFromRequest()
	val, ok := b.headers[name]
	return val, ok
}

func (b *HttpReaderExchangeBuffer) SetHeader(name string, value string) {
	b.ensureHeadersReadFromRequest()
	b.headers[strings.ToLower(name)] = value
}

func (b *HttpReaderExchangeBuffer) GetHeadersCount() int {
	b.ensureHeadersReadFromRequest()
	return len(b.headers)
}

func (b *HttpReaderExchangeBuffer) GetHeaders(cb func(name string, value string)) {
	b.ensureHeadersReadFromRequest()
	for k, v := range b.headers {
		cb(k, v)
	}
}

func (b *HttpReaderExchangeBuffer) GetStatusCode() int {
	fmt.Printf("ERROR cannot call GetStatusCode on HttpReaderExchangeBuffer instance\n")
	return -1
}

func (b *HttpReaderExchangeBuffer) GetBuffer() []byte {
	b.ensureHeadersReadFromRequest()
	body, err := ioutil.ReadAll(b.r.Body)
	if err == nil {
		if body != nil && len(body) == 0 {
			body = nil
		}

		return body
	} else {
		fmt.Printf("ERROR http wrapped request CANNOT READ BODY from request with method='%s', url='%s', maybe tried to read it twice ? (%v)\n", b.r.Method, b.r.RequestURI, err)
		return nil
	}
}

func (b *HttpReaderExchangeBuffer) WriteStatusCode(statusCode int) {
	fmt.Printf("ERROR cannot call WriteStatusCode on HttpReaderExchangeBuffer instance\n")
}

func (b *HttpReaderExchangeBuffer) Write(buffer []byte) (int, error) {
	return -1, fmt.Errorf("cannot write on a wrapped http request")
}

func (b *HttpReaderExchangeBuffer) Close() int {
	return 0
}

func (b *HttpReaderExchangeBuffer) ensureHeadersReadFromRequest() {
	if b.headersRead {
		return
	}

	b.headersRead = true

	for k, v := range b.r.Header {
		b.headers[strings.ToLower(k)] = v[0]
	}
}

/*
	WRITER
*/

type HttpWriterExchangeBuffer struct {
	w http.ResponseWriter
}

func WrapHttpWriterAsExchangeBuffer(w http.ResponseWriter) *HttpWriterExchangeBuffer {
	return &HttpWriterExchangeBuffer{
		w: w,
	}
}

func (b *HttpWriterExchangeBuffer) GetHeader(name string) (string, bool) {
	return b.w.Header().Get(name), true
}

func (b *HttpWriterExchangeBuffer) GetStatusCode() int {
	fmt.Printf("ERROR cannot call GetStatusCode on HttpWriterExchangeBuffer instance\n")
	return -1
}

func (b *HttpWriterExchangeBuffer) SetHeader(name string, value string) {
	b.w.Header().Set(name, value)
}

func (b *HttpWriterExchangeBuffer) GetHeadersCount() int {
	fmt.Printf("ERROR cannot call GetHeadersCount on HttpWriterExchangeBuffer instance\n")
	return -1
}

func (b *HttpWriterExchangeBuffer) GetHeaders(cb func(name string, value string)) {
	fmt.Printf("ERROR cannot call GetHeaders on HttpWriterExchangeBuffer instance\n")
}

func (b *HttpWriterExchangeBuffer) GetBuffer() []byte {
	fmt.Printf("ERROR cannot call GetBuffer on HttpWriterExchangeBuffer instance\n")
	return nil
}

func (b *HttpWriterExchangeBuffer) WriteStatusCode(statusCode int) {
	b.w.WriteHeader(statusCode)
}

func (b *HttpWriterExchangeBuffer) Write(buffer []byte) (int, error) {
	_, err := b.w.Write(buffer)
	if err != nil {
		return -1, err
	}

	return len(buffer), nil
}

func (b *HttpWriterExchangeBuffer) Close() int {
	return 0
}
