package common

type ExchangeBuffer struct {
	buffer []byte
}

func (p *ExchangeBuffer) GetBuffer() []byte {
	return p.buffer
}

func (p *ExchangeBuffer) Read(buffer []byte) int {
	return 0
}

func (p *ExchangeBuffer) Write(buffer []byte) int {
	// TODO WARNING THIS DOES NOT TAKE WRITE POS IN ACCOUNT !!!
	p.buffer = appendSlice(p.buffer, buffer)

	return len(buffer)
}

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
