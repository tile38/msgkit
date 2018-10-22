package msgkit

import "github.com/tidwall/gjson"

// Message is the structural representation of a msgkit message
type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// NewMessage produces a new Message reference from the passed type and data
func NewMessage(t string, d ...string) *Message {
	if len(d) == 0 {
		d = []string{""}
	}
	return &Message{Type: t, Data: d[0]}
}

// ParseMessage parses and returns a new fully populated Message type
func ParseMessage(msgb []byte) *Message {
	return NewMessage(gjson.GetBytes(msgb, "type").String(),
		gjson.GetBytes(msgb, "data").String())
}
