package msgkit

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/tile38/msgkit/safews"
)

// Context defines all functionality for handling a websocket message
type Context interface {
	// ConnID returns the ID of the client connection
	ConnID() string

	// Message returns the full message stored in the context
	Message() string

	// Bind JSON decodes the message on the context into the passed interface
	Bind(i interface{}) error

	// Send will write the passed string to the connection
	Send(s string) error
}

// context contains all context about the websocket message
type context struct {
	server  *Server
	conn    *safews.Conn
	connID  string
	message []byte
}

// ConnID returns the ID of the client connection
func (c *context) ConnID() string {
	return c.connID
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

// send will send the passed UTF8 message bytes to the passed connection
func send(conn *safews.Conn, msg string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
