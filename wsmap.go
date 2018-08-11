package gows

import (
	"sync"

	"github.com/gorilla/websocket"
)

// wsmap is a thread safe map of websocket connections
type wsmap struct {
	sync.Mutex
	conns map[string]*websocket.Conn
}

// newWSMap initializes and returns a new thread safe wsmap
func newWSMap() *wsmap {
	return &wsmap{conns: make(map[string]*websocket.Conn)}
}

// Get retrieves and returns both a connection and an exists field for the
// passed connection ID
func (m *wsmap) Get(id string) (*websocket.Conn, bool) {
	m.Lock()
	defer m.Unlock()
	conn, ok := m.conns[id]
	return conn, ok
}

// Set assigns a new value for the passed ID in the websocket connection map
func (m *wsmap) Set(id string, conn *websocket.Conn) {
	m.Lock()
	defer m.Unlock()
	m.conns[id] = conn
}

// Delete deletes a connection from the websocket connection map
func (m *wsmap) Delete(id string) {
	m.Lock()
	defer m.Unlock()
	delete(m.conns, id)
}

// Range performs the passed handler function on all connections
func (m *wsmap) Range(f func(id, conn interface{})) {
	m.Lock()
	defer m.Unlock()
	for id, conn := range m.conns {
		f(id, conn)
	}
}
