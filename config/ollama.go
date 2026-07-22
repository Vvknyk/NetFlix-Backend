package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// request struct sent to Ollama
type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// response struct received from Ollama
type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func GenerateEmbedding(text string) ([]float64, error) {
	// Step 1 — build request body
	reqBody := embeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	// Step 2 — marshal to JSON
	marshalled, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Step 3 — send POST request to Ollama
	resp, err := http.Post(
		"http://localhost:11434/api/embeddings",
		"application/json",
		bytes.NewBuffer(marshalled),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama: %v", err)
	}
	defer resp.Body.Close()

	// Step 4 — read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Step 5 — unmarshal response
	var embResp embeddingResponse
	err = json.Unmarshal(body, &embResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return embResp.Embedding, nil
}
