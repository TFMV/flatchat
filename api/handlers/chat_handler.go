package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/TFMV/flatchat/flatbuffers/flatchat"
	"github.com/TFMV/flatchat/internal/chat"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleChat(w http.ResponseWriter, r *http.Request, chatRepo *chat.ChatRepository) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	chatService := chat.NewChatService(chatRepo)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			break
		}

		log.Printf("Received raw message: %v", msg)

		// Check buffer length to ensure it contains at least the offset size
		if len(msg) < flatbuffers.SizeUOffsetT {
			log.Println("Buffer too short for Flatbuffers message")
			continue
		}

		// Deserialize message using Flatbuffers
		flatMsg := flatchat.GetRootAsMessage(msg, 0)
		if flatMsg == nil {
			log.Println("Failed to get root message")
			continue
		}

		content := flatMsg.Content()
		if content == nil {
			log.Println("Content is nil")
			continue
		}

		userMessage := chat.Message{
			Role:      "user",
			Content:   string(content),
			Timestamp: flatMsg.Timestamp(),
		}

		log.Printf("Deserialized message: %+v", userMessage)

		// Get response from ChatGPT
		responseContent, err := chatService.ProcessMessage(userMessage)
		if err != nil {
			log.Println("ProcessMessage error:", err)
			break
		}

		log.Printf("Sending response: %s", responseContent)

		// Send the ChatGPT response back to the client
		if err := sendMessage(conn, responseContent); err != nil {
			log.Println("sendMessage error:", err)
			break
		}
	}
}

func sendMessage(conn *websocket.Conn, content string) error {
	builder := flatbuffers.NewBuilder(1024)

	id := builder.CreateString("1")
	user := builder.CreateString("ChatGPT")
	contentStr := builder.CreateString(content)
	timestamp := uint64(time.Now().Unix())

	flatchat.MessageStart(builder)
	flatchat.MessageAddId(builder, id)
	flatchat.MessageAddUser(builder, user)
	flatchat.MessageAddContent(builder, contentStr)
	flatchat.MessageAddTimestamp(builder, timestamp)
	flatMsg := flatchat.MessageEnd(builder)

	builder.Finish(flatMsg)
	serializedMsg := builder.FinishedBytes()

	log.Printf("Serialized response message: %v", serializedMsg)

	return conn.WriteMessage(websocket.BinaryMessage, serializedMsg)
}
