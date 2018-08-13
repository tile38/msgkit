package main

import (
	"log"

	"github.com/tile38/msgkit"
)

func main() {
	// Initialize a msgkit server
	s := msgkit.New("/ws")

	// Bind a response handler to any JSON message that contains a "type" of "ID"
	s.Handle("ID", func(c *msgkit.Context) error {
		return c.Conn.Send(c.ConnID)
	})

	// Listen for requests on port 8000
	log.Println(s.Listen(":8000"))
}
