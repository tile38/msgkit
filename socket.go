package msgkit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Socket is msgkit socket connection containing context about the connection
type Socket struct {
	id  string
	req *http.Request

	mu   sync.Mutex
	ctx  interface{}
	conn *websocket.Conn
}

// newSocket upgrades the passed http connection to a websocket connection,
// and returns the connection bundled in a msgkit Socket
func newSocket(u *websocket.Upgrader, w http.ResponseWriter,
	r *http.Request) (*Socket, error) {
	// Upgrade the websocket connection
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	// Generate a unique identifier for the connection
	var b [12]byte
	rand.Read(b[:])

	// Assemble and return a fully populated msgkit Socket
	return &Socket{
		id:   hex.EncodeToString(b[:]),
		req:  r,
		conn: conn,
	}, nil
}

// SetContext applies the passed context interface to the Socket
func (s *Socket) SetContext(ctx interface{}) {
	s.mu.Lock()
	s.ctx = ctx
	s.mu.Unlock()
}

// Context returns the context on the Socket
func (s *Socket) Context() interface{} {
	s.mu.Lock()
	ctx := s.ctx
	s.mu.Unlock()
	return ctx
}

// Request returns the original request used to create the Socket
func (s *Socket) Request() *http.Request { return s.req }

// Send broadcasts a message over the socket
func (s *Socket) Send(name string, msgs ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(msgs) == 0 {
		s.conn.WriteMessage(1, []byte(fmt.Sprintf(`{"type":"%s"}`, name)))
	} else {
		for _, msg := range msgs {
			b, _ := json.Marshal(msg)
			s.conn.WriteMessage(1,
				[]byte(fmt.Sprintf(`{"type":"%s","data":%s}`, name, string(b))))
		}
	}
}

// close closes the websocket connection
func (s *Socket) close() { s.conn.Close() }

// readMessage reads the next message off of the connection, returning the type
// and data decoded from the message
func (s *Socket) readMessage() (string, error) {
	// Read the next message off of the connection
	_, msgb, err := s.conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(msgb), nil
}
