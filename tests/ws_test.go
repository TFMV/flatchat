package main

import (
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/TFMV/flatchat/flatbuffers/flatchat"
	flatbuffers "github.com/google/flatbuffers/go"
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

	// Build and send a single Flatbuffers message
	builder := flatbuffers.NewBuilder(1024)
	id := builder.CreateString("1")
	user := builder.CreateString("client")
	content := builder.CreateString("Hello from client!")
	timestamp := uint64(time.Now().Unix())

	flatchat.MessageStart(builder)
	flatchat.MessageAddId(builder, id)
	flatchat.MessageAddUser(builder, user)
	flatchat.MessageAddContent(builder, content)
	flatchat.MessageAddTimestamp(builder, timestamp)
	flatMsg := flatchat.MessageEnd(builder)

	builder.Finish(flatMsg)
	serializedMsg := builder.FinishedBytes()

	err = c.WriteMessage(websocket.BinaryMessage, serializedMsg)
	if err != nil {
		t.Logf("write: %v", err)
		return
	}

	select {
	case <-done:
	case <-time.After(time.Second):
	}

	log.Println("interrupt")

	// Cleanly close the connection by sending a close message and then waiting for the server to close the connection.
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		t.Logf("write close: %v", err)
		return
	}
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}
