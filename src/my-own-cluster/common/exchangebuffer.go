package common

import (
	"sync"
)

/*
	ExchangeBuffer is a byte buffer together with a set of headers.

	It is a merged abstraction between sequential file and http request.

	It is used all over the place to transfer data between functions.

	Of course, headers are case-insensitive.
*/
type ExchangeBuffer interface {
	GetHeader(name string) (string, bool)
	SetHeader(name string, value string)
	GetHeadersCount() int
	GetHeaders(cb func(name string, value string))
	GetBuffer() []byte
	Read(buffer []byte) int
	Write(buffer []byte) (int, error)
	Close() int
}

/*
	In memory implementation of ExchangeBuffer
*/
type InMemoryExchangeBuffer struct {
	headers map[string]string
	buffer  []byte
}

func (o *Orchestrator) CreateExchangeBuffer() int {
	var lock sync.Mutex

	lock.Lock()
	bufferID := o.nextExchangeBufferID
	o.nextExchangeBufferID++
	o.exchangeBuffers[int(bufferID)] = NewMemoryExchangeBuffer()
	lock.Unlock()

	return int(bufferID)
}

func (o *Orchestrator) GetExchangeBuffer(bufferID int) ExchangeBuffer {
	return o.exchangeBuffers[bufferID]
}

func (o *Orchestrator) ReleaseExchangeBuffer(bufferID int) {
	delete(o.exchangeBuffers, bufferID)
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
