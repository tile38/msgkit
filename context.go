package gows

import (
	"encoding/json"
	"net"

	"github.com/gorilla/websocket"
)

// Context defines all functionality for handling a websocket message
type Context interface {
	// ConnID returns the ID of the client connection
	ConnID() string

	// RemoteAddr returns the net Address of the client connection
	RemoteAddr() net.Addr

	// Message returns the full message stored in the context
	Message() string

	// Bind JSON decodes the message on the context into the passed interface
	Bind(i interface{}) error

	// Send will write the passed string to the connection
	Send(s string) error

	// SendAll will write the passed string to all connections
	SendAll(s string)
}

// context contains all context about the websocket message

type context struct {
	server  *Server
	conn    *websocket.Conn
	connID  string
	message []byte
}

// ConnID returns the ID of the client connection
func (c *context) ConnID() string {
	return c.connID
}

// RemoteAddr returns the net Address of the client connection
func (c *context) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Message returns the full message stored in the context
func (c *context) Message() string {
	return string(c.message)
}

// Bind JSON decodes the message on the context into the passed interface
func (c *context) Bind(i interface{}) error {
	return json.Unmarshal(c.message, i)
}

// Send will write the passed string to the connection
func (c *context) Send(s string) error {
	return send(c.conn, s)
}

// SendAll will write the passed string to all connections
func (c *context) SendAll(s string) {
	c.server.Conns.Range(func(_, value interface{}) {
		if conn, ok := value.(*websocket.Conn); ok {
			send(conn, s)
		}
	})
}

// send will send the passed UTF8 message bytes to the passed connection
func send(conn *websocket.Conn, msg string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
