// msgkit is a simple websocket json message handling package. It makes it
// quick and easy to write a websocket server using traditional http style
// request/message handlers.

package msgkit

import (
	"log"
	"net/http"

	"github.com/tidwall/gjson"
)

// Handler is a package of all required dependencies to run a msgkit websocket
// server
type Handler struct {
	conns conns

	// Event handlers for all connections
	handlers map[string]func(id, msg string)

	// OnOpen binds an on-open handler to the server which will be triggered
	// every time a connection is made
	OnOpen func(id string)

	// OnClose binds an on-close handler to the server which will trigger every
	// time a connection is closed
	OnClose func(id string)
}

// Handle adds a HandlerFunc to the map of websocket message handlers
func (h *Handler) Handle(name string, handler func(id, msg string)) {
	if h.handlers == nil {
		h.handlers = make(map[string]func(id, msg string))
	}
	h.handlers[name] = handler
}

// Send a message to a websocket.
func (h *Handler) Send(id string, message string) {
	h.conns.send(id, message)
}

// // IDs returns all connection IDs
// func (h *Handler) IDs() []string {
// 	return h.conns.IDs()
// }

// ServeHTTP is the primary websocket handler method and conforms to the
// http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Open and register the websocket
	id, err := h.conns.register(w, r)
	if err != nil {
		log.Println("register:", err)
		return
	}
	defer h.conns.unregister(id) // Defer close and unregister the websocket

	// Trigger the OnOpen handler if one is defined
	if h.OnOpen != nil {
		h.OnOpen(id)
	}

	if h.OnClose != nil {
		// Defer trigger the OnClose handler if one is defined
		defer h.OnClose(id)
	}

	// For every message that comes through on the connection
	for {
		// Read the next message on the connection
		msg, err := h.conns.receive(id)
		if err != nil {
			return
		}

		// JSON decode the type from the json formatted message
		msgType := gjson.Get(msg, "type").String()

		// If a handler exists for the message type, handle it
		if fn, ok := h.handlers[msgType]; ok {
			fn(id, msg)
		} else {
			// Send an error back to the client letting them know that the
			// incoming type is unknown
			h.conns.send(id, `{"type":"Error","message":"Unknown type"}`)
		}
	}
}
