package safews

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Map is a thread safe map of thread safe websocket connections
type Map struct {
	mu    sync.RWMutex
	conns map[string]*Conn
}

// New initializes and returns a new thread safe websocket map
func New() *Map {
	return &Map{conns: make(map[string]*Conn)}
}

// Set assigns a new value for the passed ID in the websocket connection map
func (m *Map) Set(id string, conn *Conn) {
	m.mu.Lock()
	m.conns[id] = conn
	m.mu.Unlock()
}

// SendText broadcasts a message to the connection
func (m *Map) SendText(id string, data []byte) {
	m.mu.Lock()
	if c, ok := m.conns[id]; ok {
		c.WriteMessage(websocket.TextMessage, data)
	}
	m.mu.Unlock()
}

// IDs returns all the connection ids in the websocket map
func (m *Map) IDs() []string {
	m.mu.RLock()
	var ids []string
	for id := range m.conns {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	return ids
}

// Delete deletes a connection from the websocket connection map
func (m *Map) Delete(id string) {
	m.mu.Lock()
	delete(m.conns, id)
	m.mu.Unlock()
}
