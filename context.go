package msgkit

import "github.com/tile38/msgkit/safews"

// Context contains all context about the websocket message
type Context struct {
	Server  *Server
	Conn    *safews.Conn
	ConnID  string
	Message []byte
}
