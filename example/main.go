package main

import (
	"log"
	"net/http"

	"github.com/tile38/msgkit"
)

func main() {
	// Initialize a msgkit handler
	var h msgkit.Handler

	// Bind a response handler to any JSON message with the "type" of "Echo"
	h.Handle("Echo", func(id, msg string) {
		h.Send(id, msg)
	})

	// Bind the handler to url path "/ws"
	http.Handle("/ws", &h)

	// start serving on port 8000
	srv := &http.Server{Addr: ":8000"}
	log.Fatal(srv.ListenAndServe())
}
