package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/ai"
	"kori/internal/ai/agent"
	"kori/internal/automation"
	"kori/internal/config"
	"kori/internal/models"
	"strings"

	"gorm.io/gorm"
)

// AIDecisionProcessor handles AI_DECISION nodes
type AIDecisionProcessor struct {
	db             *gorm.DB
	aiClient       ai.AIClient
	knowledgeBase  *agent.KnowledgeBase
	cfg            *config.Config
}

// AIDecisionNodeData represents the data structure for AI_DECISION nodes
type AIDecisionNodeData struct {
	Prompt        string            `json:"prompt"`        // Decision prompt/question for AI
	DecisionPaths map[string]string `json:"decisionPaths"` // Mapping of decision keys to node IDs
	Model         string            `json:"model"`         // Optional model override
	MaxTokens     int               `json:"maxTokens"`     // Max tokens for response
	Temperature   float64           `json:"temperature"`   // Temperature for randomness
}

// NewAIDecisionProcessor creates a new AI_DECISION node processor
func NewAIDecisionProcessor(db *gorm.DB, aiClient ai.AIClient, cfg *config.Config) *AIDecisionProcessor {
	return &AIDecisionProcessor{
		db:            db,
		aiClient:      aiClient,
		knowledgeBase: agent.NewKnowledgeBase(db),
		cfg:           cfg,
	}
}

// Type returns the node type this processor handles
func (p *AIDecisionProcessor) Type() models.NodeType {
	return models.NodeTypeAIDecision
}

// Validate checks if the AI_DECISION node data is valid
func (p *AIDecisionProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeAIDecision {
		return fmt.Errorf("invalid node type for AIDecisionProcessor")
	}

	var data AIDecisionNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid ai_decision node data: %w", err)
	}

	if data.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(data.DecisionPaths) == 0 {
		return fmt.Errorf("at least one decision path is required")
	}

	return nil
}

// Process executes the AI_DECISION node logic
func (p *AIDecisionProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	// Check if AI is enabled
	if !p.cfg.AI.Enabled {
		return nil, fmt.Errorf("AI is not enabled in configuration")
	}

	var data AIDecisionNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse ai_decision node data: %w", err)
	}

	// Build knowledge context
	knowledgeContext, err := p.knowledgeBase.BuildDecisionContext(context.Background(), ctx.ContactID, ctx.AutomationID)
	if err != nil {
		return nil, fmt.Errorf("failed to build knowledge context: %w", err)
	}

	// Build decision options from paths
	pathOptions := make([]string, 0, len(data.DecisionPaths))
	for key := range data.DecisionPaths {
		pathOptions = append(pathOptions, key)
	}

	// Create AI prompt
	systemPrompt := buildSystemPrompt()
	userPrompt := buildDecisionPrompt(data.Prompt, pathOptions, knowledgeContext, ctx.Variables)

	// Call AI for decision
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	maxTokens := data.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1000 // Decisions should be concise
	}

	response, err := p.aiClient.CompleteWithMessages(context.Background(), messages, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	// Parse AI response to extract decision
	decision := extractDecision(response.Content, pathOptions)

	// Get the node ID for this decision
	nextNodeID, exists := data.DecisionPaths[decision]
	if !exists {
		// Fall back to first available path if AI decision doesn't match
		for key, nodeID := range data.DecisionPaths {
			decision = key
			nextNodeID = nodeID
			break
		}
	}

	// Find the actual next node (it might be missing)
	if nextNodeID == "" {
		// Try to find via edges instead
		var edges []models.AutomationNodeEdge
		if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
			return nil, fmt.Errorf("failed to load edges: %w", err)
		}

		if len(edges) > 0 {
			nextNodeID = edges[0].TargetID
		} else {
			return &automation.ProcessResult{
				Complete: true,
				Message:  fmt.Sprintf("AI decided '%s' but no next node found", decision),
			}, nil
		}
	}

	return &automation.ProcessResult{
		NextNodeIDs: []string{nextNodeID},
		Message:     fmt.Sprintf("AI decided: %s", decision),
		UpdateVars: map[string]interface{}{
			"ai_decision":       decision,
			"ai_decision_raw":   response.Content,
			"ai_tokens_used":    response.TokensUsed,
			"ai_model":          response.Model,
		},
		Data: map[string]interface{}{
			"decision":       decision,
			"reasoning":      response.Content,
			"tokensUsed":     response.TokensUsed,
			"availablePaths": pathOptions,
			"model":          response.Model,
		},
	}, nil
}

// buildSystemPrompt creates the system prompt for AI decisions
func buildSystemPrompt() string {
	return `You are an expert email marketing strategist helping to make decisions in an automated email workflow.

Your task is to analyze the provided contact information and context, then choose the best path forward based on engagement patterns, behavior, and performance data.

You must respond with ONLY the decision key (one of the provided options) followed by a brief explanation on a new line.

Format your response exactly as:
DECISION: [decision_key]
REASONING: [brief explanation]

Be concise and data-driven in your decisions.`
}

// buildDecisionPrompt creates the user prompt with context
func buildDecisionPrompt(prompt string, options []string, knowledge string, variables map[string]interface{}) string {
	var parts []string

	parts = append(parts, "=== DECISION PROMPT ===")
	parts = append(parts, prompt)
	parts = append(parts, "")

	parts = append(parts, "=== AVAILABLE OPTIONS ===")
	for i, option := range options {
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, option))
	}
	parts = append(parts, "")

	if knowledge != "" {
		parts = append(parts, "=== CONTACT ANALYTICS ===")
		parts = append(parts, knowledge)
		parts = append(parts, "")
	}

	if len(variables) > 0 {
		parts = append(parts, "=== CURRENT VARIABLES ===")
		for key, value := range variables {
			parts = append(parts, fmt.Sprintf("%s: %v", key, value))
		}
		parts = append(parts, "")
	}

	parts = append(parts, "Based on the above information, which option should we choose?")

	return strings.Join(parts, "\n")
}

// extractDecision parses the AI response to extract the decision key
func extractDecision(response string, validOptions []string) string {
	// Look for "DECISION: <key>" pattern
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "DECISION:") {
			decision := strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(line), "DECISION:"))
			decision = strings.Trim(decision, "\"'")

			// Check if it's a valid option
			for _, option := range validOptions {
				if strings.EqualFold(decision, option) {
					return option
				}
			}
		}
	}

	// Fallback: check if any option is mentioned in the response
	responseLower := strings.ToLower(response)
	for _, option := range validOptions {
		if strings.Contains(responseLower, strings.ToLower(option)) {
			return option
		}
	}

	// Final fallback: return first option
	if len(validOptions) > 0 {
		return validOptions[0]
	}

	return ""
}
