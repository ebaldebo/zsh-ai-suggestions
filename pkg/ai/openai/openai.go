package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/prompt"
)

type OpenAIClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

func New() *OpenAIClient {
	apiKey := env.Get(envAPIKey, "")
	if apiKey == "" {
		log.Fatal("openai api key is required")
	}

	return &OpenAIClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
		model:      env.Get(envModel, defaultModel),
	}
}

func (c *OpenAIClient) Suggest(ctx context.Context, input string) (string, error) {
	request := Request{
		Model: c.model,
		Messages: []InputMessage{
			{
				Role: roleSystem, Content: prompt.Get(input),
			},
			{Role: roleUser, Content: input},
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API error (%d): %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no suggestions returned by openai")
	}

	return response.Choices[0].Message.Content, nil
}
