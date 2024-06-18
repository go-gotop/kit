package gorilla

import (
	"net/http"
	"time"

	gwebsocket "github.com/gorilla/websocket"
)

func NewGorillaWebSocketConn() *GorillaWebSocketConn {
	return &GorillaWebSocketConn{}
}

type GorillaWebSocketConn struct {
	conn *gwebsocket.Conn
}

func (g *GorillaWebSocketConn) Dial(endpoint string, requestHeader http.Header) error {
	dialer := gwebsocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(endpoint, requestHeader)
	if err != nil {
		return err
	}
	conn.SetReadLimit(655350)
	g.conn = conn
	return nil
}

func (g *GorillaWebSocketConn) ReadMessage() (int, []byte, error) {
	return g.conn.ReadMessage()
}

func (g *GorillaWebSocketConn) WriteMessage(messageType int, data []byte) error {
	return g.conn.WriteMessage(messageType, data)
}

func (g *GorillaWebSocketConn) SetPingHandler(h func(appData string) error) {
	g.conn.SetPingHandler(h)
}

func (g *GorillaWebSocketConn) SetPongHandler(h func(appData string) error) {
	g.conn.SetPongHandler(h)
}

func (g *GorillaWebSocketConn) Close() error {
	return g.conn.Close()
}
