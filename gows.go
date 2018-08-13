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
type HandlerFunc func(context Context) error

// Server is a package of all required dependencies to run a msgkit websocket
// server
type Server struct {
	router   *mux.Router
	upgrader *websocket.Upgrader
	handlers map[string]HandlerFunc
	conns    *safews.Map
}

// New creates a new msgkit Server and binds the passed path
func New(wsPath string) *Server {
	// Create the new fully populated Server
	s := &Server{
		router:   mux.NewRouter(),
		upgrader: &websocket.Upgrader{},
		handlers: make(map[string]HandlerFunc),
		conns:    safews.New(),
	}

	// Bind the websocket path and return the Server
	s.router.HandleFunc(wsPath, s.serveWs)
	return s
}

// Handle adds a HandlerFunc to the map of websocket message handlers
func (s *Server) Handle(name string, handler HandlerFunc) {
	s.handlers[name] = handler
}

// Static binds the passed directory path to the prefix path
func (s *Server) Static(prefix, path string) {
	fs := http.FileServer(http.Dir(path))
	h := http.StripPrefix(prefix, fs)
	s.router.PathPrefix(prefix).Handler(h)
}

// Listen binds the standard websocket handler containing all handling logic
// and begins listening for messages from connected websocket clients
func (s *Server) Listen(addr string) error {
	// Bind the websocket handler to the passed path
	return http.ListenAndServe(addr, s.router)
}

// serveWs is the primary websocket handler method which
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	// Generate an ID for the new websocket client
	connID := uuid.Must(uuid.NewV4()).String()

	// Upgrade the connection to a thread safe websocket connection
	conn, err := safews.NewConn(s.upgrader, w, r)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	s.conns.Set(connID, conn)

	// Defer close the connection and delete it from the connection map
	defer s.conns.Delete(connID)
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
		if h, ok := s.handlers[msgType]; ok {
			if err := h(&context{
				server:  s,
				conn:    conn,
				connID:  connID,
				message: bmsg,
			}); err != nil {
				log.Println("handle:", err)
				break
			}
		} else {
			log.Println("unsupported message:", string(bmsg))
		}
	}
}
