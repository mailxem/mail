package ai

import "context"

// AIClient defines the interface for AI providers
type AIClient interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// CompleteWithMessages generates a completion using a conversation format
	CompleteWithMessages(ctx context.Context, messages []Message, maxTokens int) (*CompletionResponse, error)

	// Stream generates a streamed completion
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)

	// GetModel returns the model being used
	GetModel() string
}
