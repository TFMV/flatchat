package main

import (
	"log"
	"net/http"
	"os"

	"github.com/TFMV/flatchat/api/handlers"
	"github.com/TFMV/flatchat/internal/chat"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from a .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Initialize the chat repository
	chatRepo := chat.NewChatRepository()

	// Set up routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleChat(w, r, chatRepo)
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
