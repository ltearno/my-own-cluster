package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpReaderExchangeBuffer struct {
	r           *http.Request
	headersRead bool
	headers     map[string]string
	bodyRead    bool
	body        []byte
}

func WrapHttpReaderAsExchangeBuffer(r *http.Request) *HttpReaderExchangeBuffer {
	return &HttpReaderExchangeBuffer{
		r:           r,
		headersRead: false,
		headers:     make(map[string]string),
		bodyRead:    false,
		body:        nil,
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

func (b *HttpReaderExchangeBuffer) GetBuffer() []byte {
	b.ensureBodyReadFromRequest()
	return b.body
}

func (b *HttpReaderExchangeBuffer) Read(buffer []byte) int {
	fmt.Printf("WARNING: Read(...) called on a wrapped http request, not implemented, should use GetBuffer()")
	return 0
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

	fmt.Printf("http req - read headers\n")

	for k, v := range b.r.Header {
		// TODO why not support multiple values ? would add complexity and one header with clear syntax parsing should be enough
		b.headers[strings.ToLower(k)] = v[0]
	}
}

func (b *HttpReaderExchangeBuffer) ensureBodyReadFromRequest() {
	if b.bodyRead {
		return
	}

	body, err := ioutil.ReadAll(b.r.Body)

	b.bodyRead = true

	fmt.Printf("http req - read body\n")

	if err == nil {
		b.body = body
	} else {
		fmt.Printf("http wrapped request CANNOT READ BODY\n")
		b.body = nil
	}
}
