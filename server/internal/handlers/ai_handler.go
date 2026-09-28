package handlers

import (
	"context"
	"fmt"
	"kori/internal/ai"
	"kori/internal/ai/agent"
	"kori/internal/config"
	"kori/internal/utils/logger"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var aiHandlerLog = logger.New("AI_HANDLER")

type AIHandler struct {
	db            *gorm.DB
	aiClient      ai.AIClient
	knowledgeBase *agent.KnowledgeBase
	cfg           *config.Config
}

func NewAIHandler(db *gorm.DB, aiClient ai.AIClient, cfg *config.Config) *AIHandler {
	return &AIHandler{
		db:            db,
		aiClient:      aiClient,
		knowledgeBase: agent.NewKnowledgeBase(db),
		cfg:           cfg,
	}
}

// QueryAnalytics handles POST /ai/query
func (h *AIHandler) QueryAnalytics(c echo.Context) error {
	if !h.cfg.AI.Enabled {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "AI features are not enabled",
		})
	}

	teamID := c.Get("teamID").(string)

	var req struct {
		Query      string                 `json:"query" validate:"required"`
		Scope      map[string]interface{} `json:"scope"` // Optional: campaignId, startDate, endDate
		MaxTokens  int                    `json:"maxTokens"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if req.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "query is required"})
	}

	aiHandlerLog.Info("Processing AI query: %s", req.Query)

	// Build analytics context based on scope
	var contextBuilder strings.Builder
	contextBuilder.WriteString("=== ANALYTICS DATA ===\n\n")

	// Team-level analytics
	days := 30 // Last 30 days by default
	if startDateStr, ok := req.Scope["startDate"].(string); ok {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			days = int(time.Since(startDate).Hours() / 24)
		}
	}

	teamContext, err := h.knowledgeBase.BuildTeamContext(context.Background(), teamID, days)
	if err == nil {
		contextBuilder.WriteString(teamContext)
		contextBuilder.WriteString("\n\n")
	}

	// Campaign-specific analytics if provided
	if campaignID, ok := req.Scope["campaignId"].(string); ok && campaignID != "" {
		campaignContext, err := h.knowledgeBase.BuildCampaignContext(context.Background(), campaignID)
		if err == nil {
			contextBuilder.WriteString("=== CAMPAIGN DETAILS ===\n")
			contextBuilder.WriteString(campaignContext)
			contextBuilder.WriteString("\n\n")
		}
	}

	// Build AI prompt
	systemPrompt := `You are an email marketing analytics expert. Analyze the provided data and answer the user's question with specific, actionable insights.

Be concise but thorough. Include relevant metrics, comparisons, and recommendations when appropriate.

If the data is insufficient to answer the question accurately, say so clearly.`

	userPrompt := fmt.Sprintf("%s\n\n=== USER QUESTION ===\n%s\n\nProvide a detailed answer based on the analytics data above.",
		contextBuilder.String(), req.Query)

	// Call AI
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	response, err := h.aiClient.CompleteWithMessages(context.Background(), messages, maxTokens)
	if err != nil {
		aiHandlerLog.Error("AI query failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process AI query",
		})
	}

	aiHandlerLog.Success("AI query completed (tokens: %d)", response.TokensUsed)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"answer":      response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"query":       req.Query,
	})
}

// OptimizeAutomation handles POST /ai/optimize
func (h *AIHandler) OptimizeAutomation(c echo.Context) error {
	if !h.cfg.AI.Enabled {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "AI features are not enabled",
		})
	}

	_ = c.Get("teamID").(string) // teamID for future ownership verification

	var req struct {
		AutomationID      string `json:"automationId" validate:"required"`
		OptimizationGoal  string `json:"optimizationGoal"` // e.g., "increase_open_rate", "reduce_unsubscribes"
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// TODO: Verify automation ownership by loading automation and checking team_id

	aiHandlerLog.Info("Generating optimization suggestions for automation %s", req.AutomationID)

	// Build optimization prompt
	systemPrompt := `You are an email marketing automation expert. Analyze the provided automation and suggest specific improvements to achieve the stated goal.

Provide actionable recommendations that can be implemented, such as:
- Timing adjustments
- Segmentation improvements
- Content personalization
- A/B testing opportunities
- Better targeting criteria

Be specific and prioritize recommendations by expected impact.`

	userPrompt := fmt.Sprintf(`Analyze this automation and suggest improvements to %s.

Automation ID: %s

Provide 3-5 specific, actionable recommendations.`,
		req.OptimizationGoal,
		req.AutomationID)

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := h.aiClient.CompleteWithMessages(context.Background(), messages, 2000)
	if err != nil {
		aiHandlerLog.Error("AI optimization failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate optimization suggestions",
		})
	}

	aiHandlerLog.Success("Optimization suggestions generated")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"suggestions":    response.Content,
		"tokensUsed":     response.TokensUsed,
		"model":          response.Model,
		"automationId":   req.AutomationID,
		"goal":           req.OptimizationGoal,
	})
}

// BuildAutomation handles POST /ai/build
func (h *AIHandler) BuildAutomation(c echo.Context) error {
	if !h.cfg.AI.Enabled {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "AI features are not enabled",
		})
	}

	var req struct {
		Description string   `json:"description" validate:"required"`
		Goals       []string `json:"goals"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if req.Description == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "description is required"})
	}

	aiHandlerLog.Info("Building automation from description: %s", req.Description)

	systemPrompt := `You are an expert at designing email automation workflows. Convert the user's description into a structured automation workflow.

Available node types:
- START: Entry point
- EMAIL: Send an email
- WAIT: Delay before next action
- CONDITION: Branching based on conditions
- ADD_TO_LIST: Add contact to a list
- TAG: Add/remove tags
- WEBHOOK: Call external URL
- UPDATE_SUBSCRIBER: Update contact fields
- AI_DECISION: AI-powered routing
- EXIT: End the workflow

Respond with a JSON structure defining the nodes and edges.

Format:
{
  "name": "Automation name",
  "description": "Brief description",
  "nodes": [
    {"id": "node1", "type": "START", "label": "Start"},
    {"id": "node2", "type": "EMAIL", "label": "Welcome Email", "data": {"templateId": "...", "smtpConfigId": "..."}}
  ],
  "edges": [
    {"source": "node1", "target": "node2"}
  ]
}`

	userPrompt := fmt.Sprintf(`Create an automation workflow for:

Description: %s

Goals: %s

Provide a complete, ready-to-use automation structure.`,
		req.Description,
		strings.Join(req.Goals, ", "))

	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := h.aiClient.CompleteWithMessages(context.Background(), messages, 3000)
	if err != nil {
		aiHandlerLog.Error("AI build failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate automation",
		})
	}

	aiHandlerLog.Success("Automation structure generated")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"automation":  response.Content,
		"tokensUsed":  response.TokensUsed,
		"model":       response.Model,
		"description": req.Description,
	})
}
