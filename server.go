// msgkit is a simple websocket json message handling package. It makes it
// quick and easy to write a websocket server using traditional http style
// request/message handlers.

package msgkit

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	// EventConnected is an event name that is fired when a client connects
	EventConnected = "connected"
	// EventDisconnected is an event name that is fired when a client disconnects
	EventDisconnected = "disconnected"
)

// Server contains all required dependencies to run a msgkit websocket server
type Server struct {
	sockets      sync.Map               // Map of Sockets
	upgrader     *websocket.Upgrader    // Shared upgrader
	connected    HandlerFunc            // Handler fired on a connection
	disconnected HandlerFunc            // Handler fired on a disconnection
	handlers     map[string]HandlerFunc // All user-defined event handlers
}

// HandlerFunc is a type that defines the function signature of a msgkit request
// handler
type HandlerFunc func(so *Socket, msg string)

// NewServer creates a new Server using the passed custom websocket upgrader
func NewServer(u *websocket.Upgrader) *Server {
	if u == nil {
		u = &websocket.Upgrader{}
	}
	return &Server{
		upgrader: u,
		handlers: make(map[string]HandlerFunc),
	}
}

// Handle binds a handler for a specified type
func (s *Server) Handle(name string, handler HandlerFunc) {
	switch name {
	case EventConnected:
		s.connected = handler
	case EventDisconnected:
		s.disconnected = handler
	default:
		if s.handlers == nil {
			s.handlers = make(map[string]HandlerFunc)
		}
		s.handlers[name] = handler
	}
}

// Broadcast sends the passed message to all clients
func (s *Server) Broadcast(name, msg string) {
	s.sockets.Range(func(_, soi interface{}) bool {
		if so, ok := soi.(*Socket); ok {
			so.Send(name, msg)
		}
		return true
	})
}

// ServeHTTP is the primary websocket handler method and conforms to the
// http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a Socket for the connection
	so, err := newSocket(s.upgrader, w, r)
	if err != nil {
		log.Println("register:", err)
		return
	}
	s.sockets.Store(so.id, so)    // Store the new socket reference
	defer so.close()              // Defer close the connection
	defer s.sockets.Delete(so.id) // Defer un-register the connection

	// Trigger a connected listener if one is defined
	if s.connected != nil {
		s.connected(so, `{"type":"connected"}`)
	}

	// Trigger an internal disconnected listener if one is defined
	if s.disconnected != nil {
		defer s.disconnected(so, `{"type":"disconnected"}`)
	}

	// For every message that comes through on the connection
	for {
		// Read the message off of the connection
		t, d, err := so.readMessage()
		if err != nil {
			so.Send("error", "Failed to read message")
			return
		}

		// If a handler exists for the message type, handle it, otherwise emit
		// an error
		if fn, ok := s.handlers[t]; ok {
			fn(so, d)
		} else {
			so.Send("error", "Unknown type")
		}
	}
}
