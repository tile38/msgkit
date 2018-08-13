package safews

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Conn is a thread safe websocket connection that only exposes the ability to
// write to the websocket connection
type Conn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// NewConn generates a new, wrapped websocket connection
func NewConn(u *websocket.Upgrader, w http.ResponseWriter,
	r *http.Request) (*Conn, error) {
	// Upgrade the connection to a websocket connection
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	// Return the wrapped websocket connection as a Conn
	return &Conn{conn: conn}, nil
}

// ReadMessage reads the next message from the websocket connection
func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	return c.conn.ReadMessage()
}

// Send writes the passed bytes to the connection
func (c *Conn) Send(data string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, []byte(data))
}

// Close closes the thread safe websocket connection
func (c *Conn) Close() {
	c.mu.Lock()
	c.conn.Close()
	c.mu.Unlock()
}
