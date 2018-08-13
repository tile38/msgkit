package msgkit

import (
	"encoding/json"
	"fmt"
)

// Error is a generic websocket error type
type Error struct {
	Type    string `json:"type"`
	Message string `json:"Message"`
}

// NewError generates a new msgkit.Error from the passed error
func jsonError(format string, a ...interface{}) string {
	b, _ := json.Marshal(&Error{"Error", fmt.Sprintf(format, a)})
	return string(b)
}
