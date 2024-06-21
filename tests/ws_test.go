package main

import (
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketServer(t *testing.T) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := "ws://localhost:8080/ws"

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				t.Logf("read: %v", err)
				return
			}
			t.Logf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case tick := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("Hello from client!"))
			if err != nil {
				t.Logf("write: %v", err)
				return
			}
			t.Logf("tick: %v", tick)
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then waiting for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				t.Logf("write close: %v", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
