package msgkit

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// shared websocket upgrader
var upgrader = new(websocket.Upgrader)

// conns is a map of connections
type conns struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

// makeID makes a unique identifier
func makeID() string {
	var b [12]byte
	if n, err := rand.Read(b[:]); err != nil || n != len(b) {
		panic("random error")
	}
	return hex.EncodeToString(b[:])
}

// register and upgrade to a new websocket. Each connection get a unique
// identifier.
func (m *conns) register(w http.ResponseWriter, r *http.Request) (
	id string, err error,
) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return "", err
	}
	id = makeID()
	m.mu.Lock()
	if m.conns == nil {
		// lazy initialization
		m.conns = make(map[string]*websocket.Conn)
	}
	m.conns[id] = conn
	m.mu.Unlock()
	return id, nil
}

// unregister and close a connection
func (m *conns) unregister(id string) {
	m.mu.Lock()
	conn := m.conns[id]
	delete(m.conns, id)
	m.mu.Unlock()
	conn.Close()
}

// receive a websocket message.
func (m *conns) receive(id string) (msg string, err error) {
	m.mu.RLock()
	conn := m.conns[id]
	m.mu.RUnlock()
	_, msgb, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(msgb), nil
}

// send a websocket message. This operation is protected by a lock because
// other goroutines may be writing to the same websocket.
func (m *conns) send(id, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.conns[id]; ok {
		conn.WriteMessage(1, []byte(msg))
	}
}
