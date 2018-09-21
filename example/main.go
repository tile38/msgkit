package main

import (
	"log"
	"net/http"

	"github.com/tile38/msgkit"
)

func main() {
	// Initialize a msgkit handler
	s := msgkit.NewServer(nil)

	// Bind a response handler to any JSON message with the "type" of "Echo"
	s.On("echo", func(so *msgkit.Socket, msg string) {
		so.Send("echo", "Hello World!")
	})

	// Bind the handler to url path "/ws"
	http.Handle("/ws", s)

	// start serving on port 8000
	srv := &http.Server{Addr: ":8000"}
	log.Fatal(srv.ListenAndServe())
}
