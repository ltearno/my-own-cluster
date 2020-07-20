package common

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

func (o *Orchestrator) CreateExchangeBuffer() int {
	o.lock.Lock()
	bufferID := o.nextExchangeBufferID
	o.nextExchangeBufferID++
	o.exchangeBuffers[int(bufferID)] = NewMemoryExchangeBuffer()
	o.lock.Unlock()

	return int(bufferID)
}

func (o *Orchestrator) GetExchangeBuffer(bufferID int) ExchangeBuffer {
	o.lock.Lock()
	buffer := o.exchangeBuffers[bufferID]
	o.lock.Unlock()
	return buffer
}

func (o *Orchestrator) ReleaseExchangeBuffer(bufferID int) {
	o.lock.Lock()
	delete(o.exchangeBuffers, bufferID)
	o.lock.Unlock()
}

func NewMemoryExchangeBuffer() *InMemoryExchangeBuffer {
	return &InMemoryExchangeBuffer{
		headers: make(map[string]string),
		buffer:  []byte{},
	}
}
