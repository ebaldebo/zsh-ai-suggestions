package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/prompt"
)

type OllamaClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

func New(httpClient *http.Client) *OllamaClient {
	return &OllamaClient{
		httpClient: httpClient,
		baseURL:    env.Get(envURL, defaultURL),
		model:      env.Get(envModel, defaultModel),
	}
}

func (c *OllamaClient) Suggest(ctx context.Context, input string) (string, error) {
	payload := Request{
		Model:  c.model,
		Prompt: prompt.Get(input),
		Stream: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read error body: %w", err)
		}

		return "", fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, errorBody)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Response, nil
}
