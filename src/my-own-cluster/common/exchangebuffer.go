package common

import "sync/atomic"

/*
	ExchangeBuffer is a byte buffer together with a set of headers.

	It is a merged abstraction between sequential file and http request.

	It is used all over the place to transfer data between functions.

	Of course, headers are case-insensitive.
*/
type ExchangeBuffer struct {
	headers map[string]string
	buffer  []byte
}

func (o *Orchestrator) CreateExchangeBuffer() int {
	bufferID := atomic.AddInt32(&o.nextExchangeBufferID, 1)
	o.exchangeBuffers[int(bufferID)] = newExchangeBuffer()
	return int(bufferID)
}

func (o *Orchestrator) GetExchangeBuffer(bufferID int) *ExchangeBuffer {
	return o.exchangeBuffers[bufferID]
}

func (o *Orchestrator) ReleaseExchangeBuffer(bufferID int) {
	delete(o.exchangeBuffers, bufferID)
}

func newExchangeBuffer() *ExchangeBuffer {
	return &ExchangeBuffer{
		headers: make(map[string]string),
		buffer:  []byte{},
	}
}

func (p *ExchangeBuffer) GetHeader(name string) (string, bool) {
	s, ok := p.headers[name]
	return s, ok
}

func (p *ExchangeBuffer) SetHeader(name string, value string) {
	p.headers[name] = value
}

func (p *ExchangeBuffer) GetHeadersCount() int {
	return len(p.headers)
}

func (p *ExchangeBuffer) GetHeaders(cb func(name string, value string)) {
	for name, value := range p.headers {
		cb(name, value)
	}
}

func (p *ExchangeBuffer) GetBuffer() []byte {
	return p.buffer
}

func (p *ExchangeBuffer) Read(buffer []byte) int {
	return 0
}

func (p *ExchangeBuffer) Write(buffer []byte) (int, error) {
	// TODO WARNING THIS DOES NOT TAKE WRITE POS IN ACCOUNT !!!
	p.buffer = appendSlice(p.buffer, buffer)

	return len(buffer), nil
}

/**
TODO : should be called 'Release' and trigger the exchange buffer GC
*/
func (p *ExchangeBuffer) Close() int {
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
