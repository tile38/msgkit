package main

import "github.com/tile38/gows"

func main() {
	// Initialize a gows server
	s := gows.New("/ws")

	// Bind a response handler to any JSON message that contains a "type" of "ID"
	s.Handle("ID", func(c gows.Context) error {
		return c.Send(c.ConnID())
	})

	// Listen for requests on port 8000
	s.Listen(":8000")
}
