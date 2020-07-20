package common

import "fmt"

/*
	In memory implementation of ExchangeBuffer
*/
type InMemoryExchangeBuffer struct {
	headers    map[string]string
	buffer     []byte
	statusCode int
}

func NewMemoryExchangeBuffer() *InMemoryExchangeBuffer {
	return &InMemoryExchangeBuffer{
		headers: make(map[string]string),
		buffer:  []byte{},
	}
}

func (p *InMemoryExchangeBuffer) GetHeader(name string) (string, bool) {
	s, ok := p.headers[name]
	return s, ok
}

func (p *InMemoryExchangeBuffer) SetHeader(name string, value string) {
	p.headers[name] = value
}

func (p *InMemoryExchangeBuffer) GetHeadersCount() int {
	return len(p.headers)
}

func (p *InMemoryExchangeBuffer) GetHeaders(cb func(name string, value string)) {
	for name, value := range p.headers {
		cb(name, value)
	}
}

func (p *InMemoryExchangeBuffer) GetBuffer() []byte {
	return p.buffer
}

func (p *InMemoryExchangeBuffer) Read(buffer []byte) int {
	return 0
}

func (p *InMemoryExchangeBuffer) WriteStatusCode(statusCode int) {
	fmt.Printf("ERROR cannot call WriteStatusCode on InMemoryExchangeBuffer instance\n")
}

func (p *InMemoryExchangeBuffer) Write(buffer []byte) (int, error) {
	// TODO WARNING THIS DOES NOT TAKE WRITE POS IN ACCOUNT !!!
	p.buffer = appendSlice(p.buffer, buffer)

	return len(buffer), nil
}

/*
	TODO : should be called 'Release' and trigger the exchange buffer GC
*/
func (p *InMemoryExchangeBuffer) Close() int {
	return 0
}

func appendSlice(slice []byte, data []byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]byte, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}
