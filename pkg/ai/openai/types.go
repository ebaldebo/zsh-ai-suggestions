package openai

import "net/http"

type Request struct {
	Model    string         `json:"model"`
	Messages []InputMessage `json:"messages"`
}

type Response struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type InputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Message struct {
	Content string `json:"content"`
}

type OpenAIClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}
