// msgkit is a basic wrapper for gorillas websocket package. It makes it
// quick and easy to write a websocket server using traditional http style
// request/message handlers and is completely thread safe

package msgkit

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"github.com/tile38/msgkit/safews"
)

// HandlerFunc is a type of function that is used for handling websocket
// messages
type HandlerFunc func(context *Context)

// EventFunc is a type of function that handles basic websocket connection
// events such as onOpen and onClose
type EventFunc func(connID string, conn *safews.Conn)

// Server is a package of all required dependencies to run a msgkit websocket
// server
type Server struct {
	Router   *mux.Router
	Upgrader *websocket.Upgrader
	Conns    *safews.Map

	// Event handlers for all connections
	Handlers map[string]HandlerFunc
	onOpen   EventFunc
	onClose  EventFunc
}

// New creates a new msgkit Server and binds the passed path
func New(wsPath string) *Server {
	// Create the new fully populated Server
	s := &Server{
		Router:   mux.NewRouter(),
		Upgrader: &websocket.Upgrader{},
		Handlers: make(map[string]HandlerFunc),
		Conns:    safews.New(),
	}

	// Bind the websocket path and return the Server
	s.Router.HandleFunc(wsPath, s.serveWs)
	return s
}

// OnOpen binds an on-open handler to the server which will be triggered every
// time a connection is made
func (s *Server) OnOpen(handler EventFunc) { s.onOpen = handler }

// OnClose binds an on-close handler to the server which will trigger every
// time a connection is closed
func (s *Server) OnClose(handler EventFunc) { s.onClose = handler }

// Handle adds a HandlerFunc to the map of websocket message handlers
func (s *Server) Handle(name string, handler HandlerFunc) {
	s.Handlers[name] = handler
}

// Static binds the passed directory path to the prefix path
func (s *Server) Static(prefix, path string) {
	fs := http.FileServer(http.Dir(path))
	h := http.StripPrefix(prefix, fs)
	s.Router.PathPrefix(prefix).Handler(h)
}

// Listen binds the standard websocket handler containing all handling logic
// and begins listening for messages from connected websocket clients
func (s *Server) Listen(addr string) error {
	// Bind the websocket handler to the passed path
	return http.ListenAndServe(addr, s.Router)
}

// serveWs is the primary websocket handler method which
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	// Generate an ID for the new websocket client
	connID := uuid.Must(uuid.NewV4()).String()

	// Upgrade the connection to a thread safe websocket connection
	conn, err := safews.NewConn(s.Upgrader, w, r)
	if err != nil {
		log.Println("Error upgrading :", err)
		return
	}
	s.Conns.Set(connID, conn)

	// Trigger the onOpen handler if one is defined
	if s.onOpen != nil {
		s.onOpen(connID, conn)
	}

	defer s.Conns.Delete(connID) // Defer delete the connection from the map
	defer conn.Close()           // Defer close the websocket connection

	// Trigger the onClose handler if one is defined
	if s.onClose != nil {
		defer s.onClose(connID, conn)
	}

	// For every message that comes through on the connection
	for {
		// Read the next message on the connection
		_, bmsg, err := conn.ReadMessage()
		if err != nil {
			conn.Send(jsonError("Error Reading : %s", err.Error()))
			break
		}

		// JSON decode the type from the json formatted message
		msgType := gjson.GetBytes(bmsg, "type").String()

		// If a handler exists for the message type, handle it
		if h, ok := s.Handlers[msgType]; ok {
			h(&Context{
				Server:  s,
				Conn:    conn,
				ConnID:  connID,
				Message: bmsg,
			})
		} else {
			conn.Send(jsonError("Unsupported message type : %s", string(bmsg)))
		}
	}
}
