package chat

import (
	"fmt"
	"sync"
	"time"
)

type ChatRepository struct {
	mu       sync.Mutex
	messages []Message
}

func NewChatRepository() *ChatRepository {
	return &ChatRepository{
		messages: make([]Message, 0),
	}
}

// SaveMessage saves a chat message to the repository.
func (repo *ChatRepository) SaveMessage(message Message) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.messages = append(repo.messages, message)
}

// GetMessages retrieves all chat messages from the repository.
func (repo *ChatRepository) GetMessages() []Message {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	return append([]Message(nil), repo.messages...) // Return a copy of the messages slice
}

// GenerateID generates a unique ID for a new message.
func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
