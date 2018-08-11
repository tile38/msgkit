// gows is a basic wrapper for gorillas websocket package. It makes it quick and
// easy to write a websocket server using traditional http style request/message
// handlers

package gows

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

// HandlerFunc is a type of function that is used for handling websocket
// messages
type HandlerFunc func(context Context) error

// Server is a package of all required dependencies to run a gows websocket
// server
type Server struct {
	Router   *mux.Router
	Upgrader *websocket.Upgrader
	Handlers map[string]HandlerFunc
	Conns    *wsmap
}

// New creates a new gows Server and binds the passed path
func New(wsPath string) *Server {
	// Create the new fully populated Server
	s := &Server{
		Router:   mux.NewRouter(),
		Upgrader: &websocket.Upgrader{},
		Handlers: make(map[string]HandlerFunc),
		Conns:    newWSMap(),
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

	// Upgrade the connection for use with websockets
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	s.Conns.Set(connID, conn)

	// Defer close the connection and delete it from the conns map
	defer s.Conns.Delete(connID)
	defer conn.Close()

	// For every message that comes through on the websocket
	for {
		// Read the next message on the connection
		_, bmsg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// JSON decode the type from the json formatted message
		msgType := gjson.GetBytes(bmsg, "type").String()

		// If a handler exists for the message type, handle it
		if h, ok := s.Handlers[msgType]; ok {
			if err := h(&context{
				server:  s,
				connID:  connID,
				conn:    conn,
				message: bmsg,
			}); err != nil {
				log.Println("handle:", err)
				break
			}
		} else {
			log.Println("Unsupported message:", string(bmsg))
		}
	}
}
