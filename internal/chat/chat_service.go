package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
)

type ChatService struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	chatRepo   *ChatRepository
}

type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp uint64 `json:"timestamp,omitempty"`
}

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func NewChatService(repo *ChatRepository) *ChatService {
	return &ChatService{
		apiKey:     os.Getenv("OPENAI_API_KEY"),
		apiURL:     "https://api.openai.com/v1/chat/completions",
		httpClient: &http.Client{Timeout: 10 * time.Second},
		chatRepo:   repo,
	}
}

func (s *ChatService) ProcessMessage(userMessage Message) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("OpenAI API key is not set")
	}

	// Save the user message to the repository
	s.chatRepo.SaveMessage(userMessage)

	chatGPTRequest := ChatGPTRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: userMessage.Content,
			},
		},
	}

	reqBody, err := json.Marshal(chatGPTRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get a valid response from OpenAI API")
	}

	var chatGPTResponse ChatGPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatGPTResponse); err != nil {
		return "", err
	}

	if len(chatGPTResponse.Choices) == 0 {
		return "", errors.New("no response from ChatGPT")
	}

	// Save the ChatGPT response to the repository
	responseMessage := Message{
		Role:      "assistant",
		Content:   chatGPTResponse.Choices[0].Message.Content,
		Timestamp: uint64(time.Now().Unix()),
	}
	s.chatRepo.SaveMessage(responseMessage)

	return chatGPTResponse.Choices[0].Message.Content, nil
}
