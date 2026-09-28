package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
	"kori/internal/utils/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var engineLog = logger.New("AUTOMATION_ENGINE")

// Engine orchestrates automation execution
type Engine struct {
	db         *gorm.DB
	taskClient tasks.AutomationEnqueuer
	registry   *ProcessorRegistry
	eventBus   *events.EventBus
}

// NewEngine creates a new automation engine
func NewEngine(db *gorm.DB, taskClient tasks.AutomationEnqueuer) *Engine {
	engine := &Engine{
		db:         db,
		taskClient: taskClient,
		registry:   NewProcessorRegistry(),
		eventBus:   events.NewEventBus(),
	}

	return engine
}

// RegisterProcessor adds a node processor to the engine
func (e *Engine) RegisterProcessor(processor NodeProcessor) {
	e.registry.Register(processor)
	engineLog.Info("Registered processor for node type: %s", processor.Type())
}

// Execute starts or resumes an automation execution
func (e *Engine) Execute(ctx context.Context, automationID, contactID string, triggerData map[string]interface{}, currentNodeID string) error {
	// Load automation with nodes and edges
	var automation models.Automation
	if err := e.db.WithContext(ctx).
		Preload("Nodes", "is_deleted = ?", false).
		Preload("Edges").
		Where("id = ? AND is_deleted = ?", automationID, false).
		First(&automation).Error; err != nil {
		return fmt.Errorf("failed to load automation: %w", err)
	}

	// Check if automation is active
	if !automation.IsActive && currentNodeID == "" {
		return fmt.Errorf("automation is not active")
	}

	// Load contact
	var contact models.Contact
	if err := e.db.WithContext(ctx).Where("id = ? AND team_id = ? AND is_deleted = ?", contactID, automation.TeamID, false).First(&contact).Error; err != nil {
		return fmt.Errorf("failed to load contact: %w", err)
	}

	// Load or create execution record
	execution, isNew, err := e.getOrCreateExecution(ctx, automationID, contactID, triggerData, currentNodeID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}

	if execution.Status == models.ExecutionStatusCompleted || (execution.Status == models.ExecutionStatusWaiting && currentNodeID == "") {
		return nil
	}

	// Create execution context
	execCtx := e.buildExecutionContext(execution, &contact, &automation)

	// Emit start event for new executions
	if isNew {
		events.Emit("automation.execution.started", execution)
	}

	// Find starting node
	var startNode *models.AutomationNode
	if currentNodeID != "" {
		// Resume from specific node
		for i := range automation.Nodes {
			if automation.Nodes[i].ID == currentNodeID {
				startNode = &automation.Nodes[i]
				break
			}
		}
	} else {
		// Start from START node
		for i := range automation.Nodes {
			if automation.Nodes[i].Type == models.NodeTypeStart {
				startNode = &automation.Nodes[i]
				break
			}
		}
	}

	if startNode == nil {
		return fmt.Errorf("could not find start node")
	}

	// Execute the workflow
	if err := e.executeWorkflow(ctx, execCtx, startNode, &automation); err != nil {
		// Mark execution as failed
		execution.Status = models.ExecutionStatusFailed
		execution.Error = err.Error()
		execution.CompletedAt = time.Now()
		if saveErr := e.saveExecution(ctx, execution, execCtx); saveErr != nil {
			return fmt.Errorf("%v; save execution: %w", err, saveErr)
		}
		events.Emit("automation.execution.failed", execution)
		return err
	}

	return nil
}

// executeWorkflow processes nodes in the workflow
func (e *Engine) executeWorkflow(ctx context.Context, execCtx *ExecutionContext, currentNode *models.AutomationNode, automation *models.Automation) error {
	for currentNode != nil {
		engineLog.Info("Processing node %s (type: %s)", currentNode.ID, currentNode.Type)

		// Get processor for this node type
		processor, exists := e.registry.Get(currentNode.Type)
		if !exists {
			return fmt.Errorf("no processor found for node type: %s", currentNode.Type)
		}

		// Update current node
		execCtx.CurrentNode = currentNode
		execCtx.Execution.CurrentNodeID = currentNode.ID

		// Process the node
		startTime := time.Now()
		result, err := processor.Process(execCtx, currentNode)
		duration := time.Since(startTime)

		// Log the execution
		if err != nil {
			execCtx.AddLog(currentNode.ID, currentNode.Type, "failed", "Node processing failed", err)
			return fmt.Errorf("node %s processing failed: %w", currentNode.ID, err)
		}

		// Add success log
		execCtx.AddLog(currentNode.ID, currentNode.Type, "success", result.Message, nil)
		execCtx.ExecutionLog[len(execCtx.ExecutionLog)-1].DurationMs = duration.Milliseconds()

		// Emit node executed event
		events.Emit("automation.node.executed", map[string]interface{}{
			"automationId": automation.ID,
			"nodeId":       currentNode.ID,
			"nodeType":     currentNode.Type,
			"contactId":    execCtx.ContactID,
			"duration":     duration,
		})

		// Update variables
		if result.UpdateVars != nil {
			for key, value := range result.UpdateVars {
				execCtx.UpdateVariable(key, value)
			}
		}

		// Check if workflow is complete
		if result.Complete {
			execCtx.Execution.Status = models.ExecutionStatusCompleted
			execCtx.Execution.CompletedAt = time.Now()
			if err := e.saveExecution(ctx, execCtx.Execution, execCtx); err != nil {
				return err
			}
			events.Emit("automation.execution.completed", execCtx.Execution)
			engineLog.Success("Automation execution completed: %s", execCtx.Execution.ID)
			return nil
		}

		// Check if we need to wait
		if result.Wait != nil && *result.Wait > 0 {
			// Save current state
			execCtx.Execution.Status = models.ExecutionStatusWaiting
			if err := e.saveExecution(ctx, execCtx.Execution, execCtx); err != nil {
				return err
			}

			// Schedule resume task
			if len(result.NextNodeIDs) > 0 {
				if err := e.scheduleResume(ctx, automation.ID, execCtx.ContactID, result.NextNodeIDs[0], *result.Wait, execCtx.Execution.ID); err != nil {
					return err
				}
			}

			if len(result.NextNodeIDs) == 0 {
				return fmt.Errorf("delay step has no continuation")
			}
			engineLog.Info("Execution paused for %s, will resume at node %s", result.Wait, result.NextNodeIDs[0])
			return nil
		}

		// Find next node
		if len(result.NextNodeIDs) == 0 {
			// No next node, workflow complete
			execCtx.Execution.Status = models.ExecutionStatusCompleted
			execCtx.Execution.CompletedAt = time.Now()
			if err := e.saveExecution(ctx, execCtx.Execution, execCtx); err != nil {
				return err
			}
			events.Emit("automation.execution.completed", execCtx.Execution)
			return nil
		}

		// Get the next node (for now, just take the first one)
		// In future, CONDITION nodes may return multiple paths
		nextNodeID := result.NextNodeIDs[0]
		currentNode = e.findNodeByID(automation.Nodes, nextNodeID)

		if currentNode == nil {
			return fmt.Errorf("next node not found: %s", nextNodeID)
		}

		// Save intermediate state
		execCtx.Execution.Status = models.ExecutionStatusRunning
		if err := e.saveExecution(ctx, execCtx.Execution, execCtx); err != nil {
			return err
		}
	}

	return nil
}

// getOrCreateExecution retrieves existing execution or creates a new one
func (e *Engine) getOrCreateExecution(ctx context.Context, automationID, contactID string, triggerData map[string]interface{}, currentNodeID string) (*models.AutomationExecution, bool, error) {
	// Queue retries and wait continuations retain their execution identity.
	var execution models.AutomationExecution
	if id, _ := triggerData["_execution_id"].(string); id != "" {
		if uuid.Validate(id) != nil {
			return nil, false, fmt.Errorf("invalid execution ID")
		}
		data, err := json.Marshal(triggerData)
		if err != nil {
			return nil, false, err
		}
		execution = models.AutomationExecution{Base: models.Base{ID: id}, AutomationID: automationID, ContactID: contactID, Status: models.ExecutionStatusRunning, CurrentNodeID: currentNodeID, ExecutionLog: []byte("[]"), Variables: data, TriggerData: data, StartedAt: time.Now(), TriggerType: "event"}
		result := e.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&execution)
		if result.Error != nil {
			return nil, false, result.Error
		}
		created := result.RowsAffected > 0
		if err := e.db.WithContext(ctx).Where("id = ? AND automation_id = ? AND contact_id = ?", id, automationID, contactID).First(&execution).Error; err != nil {
			return nil, false, err
		}
		return &execution, created, nil
	}
	// Compatibility for callers created before durable execution IDs.
	err := e.db.WithContext(ctx).
		Where("automation_id = ? AND contact_id = ? AND status IN ?",
			automationID, contactID, []string{string(models.ExecutionStatusRunning), string(models.ExecutionStatusWaiting)}).
		First(&execution).Error

	if err == nil {
		// Found existing execution
		return &execution, false, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, false, err
	}

	// Create new execution
	triggerDataJSON, _ := json.Marshal(triggerData)
	execution = models.AutomationExecution{
		AutomationID:  automationID,
		ContactID:     contactID,
		Status:        models.ExecutionStatusRunning,
		CurrentNodeID: currentNodeID,
		ExecutionLog:  []byte("[]"),
		Variables:     triggerDataJSON,
		StartedAt:     time.Now(),
		TriggerType:   "manual",
		TriggerData:   triggerDataJSON,
	}

	if err := e.db.WithContext(ctx).Create(&execution).Error; err != nil {
		return nil, false, err
	}

	return &execution, true, nil
}

// buildExecutionContext creates execution context from execution record
func (e *Engine) buildExecutionContext(execution *models.AutomationExecution, contact *models.Contact, automation *models.Automation) *ExecutionContext {
	var variables map[string]interface{}
	json.Unmarshal(execution.Variables, &variables)

	var executionLog []LogEntry
	json.Unmarshal(execution.ExecutionLog, &executionLog)

	execCtx := &ExecutionContext{
		AutomationID: execution.AutomationID,
		ContactID:    execution.ContactID,
		TeamID:       automation.TeamID,
		Variables:    variables,
		ExecutionLog: executionLog,
		Contact:      contact,
		Execution:    execution,
	}
	execCtx.RefreshContactVariables()
	return execCtx
}

// saveExecution persists the execution state
func (e *Engine) saveExecution(ctx context.Context, execution *models.AutomationExecution, execCtx *ExecutionContext) error {
	// Serialize variables and log
	variablesJSON, _ := json.Marshal(execCtx.Variables)
	logJSON, _ := json.Marshal(execCtx.ExecutionLog)

	execution.Variables = variablesJSON
	execution.ExecutionLog = logJSON
	execution.UpdatedAt = time.Now()

	return e.db.WithContext(ctx).Save(execution).Error
}

// scheduleResume schedules a task to resume execution after a delay
func (e *Engine) scheduleResume(ctx context.Context, automationID, contactID, nextNodeID string, delay time.Duration, executionID string) error {
	task := tasks.AutomationExecuteTask{
		ExecutionID:   executionID,
		AutomationID:  automationID,
		ContactID:     contactID,
		TriggerData:   make(map[string]interface{}),
		CurrentNodeID: nextNodeID,
	}

	return e.taskClient.EnqueueAutomationTask(ctx, task, delay)
}

// findNodeByID finds a node in the automation by ID
func (e *Engine) findNodeByID(nodes []models.AutomationNode, nodeID string) *models.AutomationNode {
	for i := range nodes {
		if nodes[i].ID == nodeID {
			return &nodes[i]
		}
	}
	return nil
}

// GetNextNodes returns the nodes connected to the given node via edges
func (e *Engine) GetNextNodes(nodeID string, edges []models.AutomationNodeEdge) []string {
	var nextNodes []string
	for _, edge := range edges {
		if edge.SourceID == nodeID {
			nextNodes = append(nextNodes, edge.TargetID)
		}
	}
	return nextNodes
}
