package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type WebSocketExchangeBuffer struct {
	r       *http.Request
	c       *websocket.Conn
	headers map[string]string
}

func WrapWebSocketAsExchangeBuffer(r *http.Request, c *websocket.Conn) *WebSocketExchangeBuffer {
	b := &WebSocketExchangeBuffer{
		r:       r,
		c:       c,
		headers: make(map[string]string),
	}

	b.readHeadersFromRequest()

	return b
}

func (b *WebSocketExchangeBuffer) GetHeader(name string) (string, bool) {
	val, ok := b.headers[name]
	return val, ok
}

func (b *WebSocketExchangeBuffer) SetHeader(name string, value string) {
	b.headers[strings.ToLower(name)] = value
}

func (b *WebSocketExchangeBuffer) GetHeadersCount() int {
	return len(b.headers)
}

func (b *WebSocketExchangeBuffer) GetHeaders(cb func(name string, value string)) {
	for k, v := range b.headers {
		cb(k, v)
	}
}

func (b *WebSocketExchangeBuffer) GetStatusCode() int {
	fmt.Printf("WARNING: GetStatusCode() called on a wrapped web socket, not implemented, should use GetBuffer()")
	return 101
}

func (b *WebSocketExchangeBuffer) GetBuffer() []byte {
	mt, buf, err := b.c.ReadMessage()
	if err != nil {
		return nil
	}

	fmt.Printf("websocket : just read message (%d) from client : %s\n", mt, string(buf))

	return buf
}

func (b *WebSocketExchangeBuffer) WriteStatusCode(statusCode int) {
	fmt.Printf("ERROR cannot call WriteStatusCode on WebSocketExchangeBuffer instance\n")
}

func (b *WebSocketExchangeBuffer) Write(buffer []byte) (int, error) {
	b.c.WriteMessage(websocket.BinaryMessage, buffer)

	fmt.Printf("websocket : just wrote message to client : %s\n", string(buffer))

	return len(buffer), nil
}

func (b *WebSocketExchangeBuffer) Close() int {
	b.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	b.c.Close()
	return 0
}

func (b *WebSocketExchangeBuffer) readHeadersFromRequest() {
	for k, v := range b.r.Header {
		b.headers[strings.ToLower(k)] = v[0]
	}
}
