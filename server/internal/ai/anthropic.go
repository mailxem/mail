package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kori/internal/utils/logger"
	"net/http"
	"time"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

var aiLog = logger.New("AI_CLIENT")

// AnthropicClient implements AIClient for Anthropic's Claude
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// Anthropic API request/response types
type anthropicRequest struct {
	Model       string              `json:"model"`
	MaxTokens   int                 `json:"max_tokens"`
	Messages    []anthropicMessage  `json:"messages"`
	System      string              `json:"system,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
	StopSequences []string          `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []anthropicContent  `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	StopSequence string              `json:"stop_sequence"`
	Usage        anthropicUsage      `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewAnthropicClient creates a new Anthropic AI client
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Complete generates a completion for the given prompt
func (c *AnthropicClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Convert simple prompt to messages format
	messages := []Message{
		{Role: "user", Content: req.Prompt},
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return c.CompleteWithMessages(ctx, messages, maxTokens)
}

// CompleteWithMessages generates a completion using a conversation format
func (c *AnthropicClient) CompleteWithMessages(ctx context.Context, messages []Message, maxTokens int) (*CompletionResponse, error) {
	aiLog.Info("Generating completion with Anthropic Claude (%s)", c.model)

	// Convert to Anthropic message format
	anthropicMessages := make([]anthropicMessage, 0, len(messages))
	var systemPrompt string

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, anthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// Build request
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages:  anthropicMessages,
		System:    systemPrompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var apiErr anthropicError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	var content string
	if len(anthropicResp.Content) > 0 {
		content = anthropicResp.Content[0].Text
	}

	aiLog.Success("Completion generated in %s (tokens: %d)", duration, anthropicResp.Usage.OutputTokens)

	return &CompletionResponse{
		Content:      content,
		StopReason:   anthropicResp.StopReason,
		TokensUsed:   anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		Model:        anthropicResp.Model,
		FinishReason: anthropicResp.StopReason,
	}, nil
}

// Stream generates a streamed completion
func (c *AnthropicClient) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// For now, return a channel with a single chunk (full response)
	// Streaming can be implemented later if needed
	ch := make(chan StreamChunk, 1)

	go func() {
		defer close(ch)

		resp, err := c.Complete(ctx, req)
		if err != nil {
			ch <- StreamChunk{Error: err, Done: true}
			return
		}

		ch <- StreamChunk{Content: resp.Content, Done: true}
	}()

	return ch, nil
}

// GetModel returns the model being used
func (c *AnthropicClient) GetModel() string {
	return c.model
}
