package processors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kori/internal/automation"
	"kori/internal/models"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// WebhookProcessor handles WEBHOOK nodes
type WebhookProcessor struct {
	db *gorm.DB
}

// WebhookNodeData represents the data structure for WEBHOOK nodes
type WebhookNodeData struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // GET, POST, PUT, DELETE
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"` // JSON body for POST/PUT
	Timeout int               `json:"timeout"` // Timeout in seconds (default: 30)
}

// NewWebhookProcessor creates a new WEBHOOK node processor
func NewWebhookProcessor(db *gorm.DB) *WebhookProcessor {
	return &WebhookProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *WebhookProcessor) Type() models.NodeType {
	return models.NodeTypeWebhook
}

// Validate checks if the WEBHOOK node data is valid
func (p *WebhookProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeWebhook {
		return fmt.Errorf("invalid node type for WebhookProcessor")
	}

	var data WebhookNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid webhook node data: %w", err)
	}

	if data.URL == "" {
		return fmt.Errorf("url is required")
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	if data.Method == "" {
		data.Method = "POST"
	}
	valid := false
	for _, m := range validMethods {
		if data.Method == m {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("method must be one of: GET, POST, PUT, DELETE, PATCH")
	}

	return nil
}

// Process executes the WEBHOOK node logic
func (p *WebhookProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data WebhookNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse webhook node data: %w", err)
	}

	// Default method
	if data.Method == "" {
		data.Method = "POST"
	}

	// Default timeout
	timeout := 30 * time.Second
	if data.Timeout > 0 {
		timeout = time.Duration(data.Timeout) * time.Second
	}

	// Prepare request body
	var reqBody io.Reader
	if data.Body != nil && (data.Method == "POST" || data.Method == "PUT" || data.Method == "PATCH") {
		// If body is a string, use it as template and replace variables
		if bodyStr, ok := data.Body.(string); ok {
			// Replace variables in body
			for key, val := range ctx.Variables {
				placeholder := fmt.Sprintf("{{%s}}", key)
				bodyStr = replaceAll(bodyStr, placeholder, fmt.Sprintf("%v", val))
			}
			reqBody = bytes.NewBufferString(bodyStr)
		} else {
			// Otherwise, marshal as JSON
			bodyBytes, err := json.Marshal(data.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal webhook body: %w", err)
			}
			reqBody = bytes.NewBuffer(bodyBytes)
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request
	req, err := http.NewRequest(data.Method, data.URL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Posthoot-Automation/1.0")
	for key, value := range data.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook returned non-success status: %d - %s", resp.StatusCode, string(respBody))
	}

	// Try to parse response as JSON
	var respData interface{}
	if err := json.Unmarshal(respBody, &respData); err == nil {
		// Successfully parsed as JSON, store in variables
		if respMap, ok := respData.(map[string]interface{}); ok {
			for key, val := range respMap {
				ctx.UpdateVariable(fmt.Sprintf("webhook_%s", key), val)
			}
		}
	}

	// Get next nodes
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	nextNodeIDs := make([]string, len(edges))
	for i, edge := range edges {
		nextNodeIDs[i] = edge.TargetID
	}

	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Message:     fmt.Sprintf("Webhook %s %s completed with status %d", data.Method, data.URL, resp.StatusCode),
		UpdateVars: map[string]interface{}{
			"webhook_status_code": resp.StatusCode,
			"webhook_duration_ms": duration.Milliseconds(),
		},
		Data: map[string]interface{}{
			"url":          data.URL,
			"method":       data.Method,
			"statusCode":   resp.StatusCode,
			"durationMs":   duration.Milliseconds(),
			"responseBody": string(respBody),
		},
	}, nil
}

// replaceAll is a helper function to replace all occurrences
func replaceAll(str, old, new string) string {
	result := str
	for {
		newStr := bytes.Replace([]byte(result), []byte(old), []byte(new), 1)
		if string(newStr) == result {
			break
		}
		result = string(newStr)
	}
	return result
}
