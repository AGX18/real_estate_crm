package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const embeddingsURL = "https://api.openai.com/v1/embeddings"

type OpenAIEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

type openAIRequest struct {
	Model          string `json:"model"`
	Input          string `json:"input"`
	EncodingFormat string `json:"encoding_format"`
	Dimensions     int    `json:"dimensions,omitempty"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewOpenAIEmbedder(apiKey, model string, dimensions int) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	if e.apiKey == "" {
		return nil, nil
	}

	body, err := json.Marshal(openAIRequest{
		Model:          e.model,
		Input:          text,
		EncodingFormat: "float",
		Dimensions:     e.dimensions,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, embeddingsURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if parsed.Error != nil {
			return nil, fmt.Errorf("openai embeddings error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("openai embeddings error: status %d", resp.StatusCode)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("openai embeddings error: empty response")
	}

	return parsed.Data[0].Embedding, nil
}
