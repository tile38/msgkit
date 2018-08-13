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
type HandlerFunc func(context *Context) error

// Server is a package of all required dependencies to run a msgkit websocket
// server
type Server struct {
	Router   *mux.Router
	Upgrader *websocket.Upgrader
	Handlers map[string]HandlerFunc
	Conns    *safews.Map
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

	// Defer close the connection and delete it from the connection map
	defer s.Conns.Delete(connID)
	defer conn.Close()

	// For every message that comes through on the websocket
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
			if err := h(&Context{
				Server:  s,
				Conn:    conn,
				ConnID:  connID,
				Message: bmsg,
			}); err != nil {
				conn.Send(jsonError("Error Handling : %s", err.Error()))
			}
		} else {
			conn.Send(jsonError("Unsupported message type : %s", string(bmsg)))
		}
	}
}
