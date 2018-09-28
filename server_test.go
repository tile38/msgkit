package msgkit

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

func TestHandler(t *testing.T) {
	const addr = "localhost:17892"
	const connsN = 10  // number of concurrent sockets
	const msgsN = 1000 // number of messages per socket

	s := NewServer(nil)

	// create handlers
	s.On("h0", func(so *Socket, msg string) { so.Send("h0", msg) })
	s.On("h1", func(so *Socket, msg string) { so.Send("h1", msg) })
	s.On("h2", func(so *Socket, msg string) { so.Send("h2", msg) })

	// count the number of opens
	var opened int32
	s.On("connected", func(_ *Socket, _ string) { atomic.AddInt32(&opened, 1) })

	// count/wait on all closes
	var cwg sync.WaitGroup
	cwg.Add(connsN)
	s.On("disconnected", func(_ *Socket, _ string) { cwg.Done() })

	srv := &http.Server{Addr: addr}
	http.Handle("/ws", s)

	var swg sync.WaitGroup
	swg.Add(1)
	go func() {
		defer swg.Done()
		if err := srv.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				panic(err)
			}
		}
	}()
	var wg sync.WaitGroup
	wg.Add(connsN)
	for i := 0; i < connsN; i++ {
		go func(i int) {
			defer wg.Done()
			u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				panic(err)
			}
			defer c.Close()

			// send and receive back basic messages
			msgm := make(map[string]bool)
			for j := 0; j < msgsN; j++ {
				msg := fmt.Sprintf(`{"type":"h%d","data":"%d%d"}`, j%3, j, i)
				c.WriteMessage(1, []byte(msg))
				msgm[msg] = true
			}
			for j := 0; j < msgsN; j++ {
				_, msgb, _ := c.ReadMessage()
				if !msgm[string(msgb)] {
					panic("bad read")
				}
				delete(msgm, string(msgb))
			}
			// send an invalid type
			c.WriteMessage(1, []byte(`{"type":"invalid"}`))
			_, msgb, _ := c.ReadMessage()
			if gjson.GetBytes(msgb, "type").String() != "error" {
				panic("expected error")
			}
		}(i)
	}
	wg.Wait()
	if err := srv.Shutdown(nil); err != nil {
		t.Fatal(err)
	}
	swg.Wait()
	cwg.Wait()
	if opened != connsN {
		t.Fatalf("expected '%v', got '%v'", connsN, opened)
	}
}
