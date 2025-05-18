package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/prompt"
)

type GeminiClient struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

func New(httpClient *http.Client) *GeminiClient {
	apiKey := env.Get(envAPIKey, "")
	if apiKey == "" {
		log.Fatal("gemini api key is required")
	}

	return &GeminiClient{
		httpClient: httpClient,
		apiKey:     apiKey,
		model:      env.Get(envModel, defaultModel),
	}
}

func (c *GeminiClient) Suggest(ctx context.Context, input string) (string, error) {
	request := Request{
		SystemInstruction: Instruction{
			Parts: []Part{
				{Text: prompt.Get(input)},
			},
		},
		Contents: Instruction{
			Parts: []Part{
				{Text: input},
			},
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(baseUrl, c.model), bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	query := req.URL.Query()
	query.Add("key", c.apiKey)
	req.URL.RawQuery = query.Encode()

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
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read error body: %w", err)
		}
		return "", fmt.Errorf("gemini API error (%d): %s", resp.StatusCode, errorBody)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Candidates) == 0 || response.Candidates[0].Content.Parts[0].Text == "" {
		return "", fmt.Errorf("no suggestion from gemini")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}
