package handlers

import (
	"log"
	"net/http"

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
		log.Println(err)
		return
	}
	defer conn.Close()

	chatService := chat.NewChatService(chatRepo)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Deserialize message using Flatbuffers
		flatMsg := flatchat.GetRootAsMessage(msg, 0)
		userMessage := chat.Message{
			Role:      "user",
			Content:   string(flatMsg.Content()),
			Timestamp: flatMsg.Timestamp(),
		}

		log.Printf("Received message: %s", userMessage.Content)

		// Get response from ChatGPT
		responseContent, err := chatService.ProcessMessage(userMessage)
		if err != nil {
			log.Println(err)
			break
		}

		// Send the ChatGPT response back to the client
		if err := sendMessage(conn, responseContent); err != nil {
			log.Println(err)
			break
		}
	}
}

func sendMessage(conn *websocket.Conn, content string) error {
	builder := flatbuffers.NewBuilder(1024)

	id := builder.CreateString("1")
	user := builder.CreateString("ChatGPT")
	contentStr := builder.CreateString(content)
	timestamp := uint64(0)

	flatchat.MessageStart(builder)
	flatchat.MessageAddId(builder, id)
	flatchat.MessageAddUser(builder, user)
	flatchat.MessageAddContent(builder, contentStr)
	flatchat.MessageAddTimestamp(builder, timestamp)
	flatMsg := flatchat.MessageEnd(builder)

	builder.Finish(flatMsg)
	serializedMsg := builder.FinishedBytes()

	return conn.WriteMessage(websocket.BinaryMessage, serializedMsg)
}
