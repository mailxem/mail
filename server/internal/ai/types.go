package ai

// CompletionRequest represents a request to generate a completion
type CompletionRequest struct {
	Prompt      string
	MaxTokens   int
	Temperature float64
	SystemPrompt string
	StopSequences []string
}

// CompletionResponse represents the response from a completion request
type CompletionResponse struct {
	Content      string
	StopReason   string
	TokensUsed   int
	Model        string
	FinishReason string
}

// StreamChunk represents a chunk of streamed data
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}
