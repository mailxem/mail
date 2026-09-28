package services

import (
	"context"
	"fmt"
	"kori/internal/models"
	"kori/internal/utils/logger"

	"gorm.io/gorm"
)

var automationLog = logger.New("AUTOMATION_SERVICE")

type AutomationService struct {
	base BaseService[models.Automation]
	db   *gorm.DB
}

func NewAutomationService(db *gorm.DB) *AutomationService {
	return &AutomationService{
		base: NewBaseService(db, models.Automation{}),
		db:   db,
	}
}

// Create creates a new automation
func (s *AutomationService) Create(ctx context.Context, automation *models.Automation, includes ...string) error {
	return s.base.Create(ctx, automation, includes...)
}

// Get retrieves an automation by ID
func (s *AutomationService) Get(ctx context.Context, id string, includes ...string) (*models.Automation, error) {
	return s.base.Get(ctx, id, includes...)
}

// List retrieves automations with filters
func (s *AutomationService) List(ctx context.Context, page, limit int, filters map[string]interface{}, excludeFields map[string]bool, sortFields []string, order string, includes ...string) ([]models.Automation, int64, error) {
	return s.base.List(ctx, page, limit, filters, excludeFields, sortFields, order, includes...)
}

// Update updates an automation
func (s *AutomationService) Update(ctx context.Context, id string, automation *models.Automation, includes ...string) error {
	return s.base.Update(ctx, id, automation, includes...)
}

// Delete deletes an automation
func (s *AutomationService) Delete(ctx context.Context, id string) error {
	return s.base.Delete(ctx, id)
}

// GetWithNodesAndEdges loads automation with full graph (nodes and edges)
func (s *AutomationService) GetWithNodesAndEdges(ctx context.Context, id string) (*models.Automation, error) {
	var automation models.Automation

	if err := s.db.WithContext(ctx).
		Preload("Nodes", "is_deleted = ?", false).
		Preload("Edges").
		Preload("Edges.Source").
		Preload("Edges.Target").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&automation).Error; err != nil {
		return nil, err
	}

	return &automation, nil
}

// ValidateGraph validates the automation graph structure
func (s *AutomationService) ValidateGraph(automation *models.Automation) error {
	if len(automation.Nodes) == 0 {
		return fmt.Errorf("automation must have at least one node")
	}

	// Check for START node
	startCount := 0
	startID := ""
	nodeIDs := make(map[string]bool)

	for _, node := range automation.Nodes {
		if node.ID == "" || nodeIDs[node.ID] {
			return fmt.Errorf("node IDs must be unique")
		}
		nodeIDs[node.ID] = true
		if node.Type == models.NodeTypeStart {
			startCount++
			startID = node.ID
		}
	}

	if startCount != 1 {
		return fmt.Errorf("automation must have exactly one START node")
	}

	// Validate edges reference existing nodes
	for _, edge := range automation.Edges {
		if !nodeIDs[edge.SourceID] {
			return fmt.Errorf("edge references non-existent source node: %s", edge.SourceID)
		}
		if !nodeIDs[edge.TargetID] {
			return fmt.Errorf("edge references non-existent target node: %s", edge.TargetID)
		}
	}

	// Every step must be reachable from the single trigger.
	reached := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		if reached[id] {
			return
		}
		reached[id] = true
		for _, e := range automation.Edges {
			if e.SourceID == id {
				visit(e.TargetID)
			}
		}
	}
	visit(startID)
	if len(reached) != len(automation.Nodes) {
		return fmt.Errorf("all steps must be reachable from the trigger")
	}
	// Check for cycles using DFS
	if hasCycle := s.detectCycle(automation.Nodes, automation.Edges); hasCycle {
		return fmt.Errorf("automation graph contains cycles")
	}

	// Check for orphaned nodes (except START and EXIT)
	if err := s.checkOrphanedNodes(automation.Nodes, automation.Edges); err != nil {
		return err
	}

	return nil
}

// detectCycle uses DFS to detect cycles in the graph
func (s *AutomationService) detectCycle(nodes []models.AutomationNode, edges []models.AutomationNodeEdge) bool {
	// Build adjacency list
	adjList := make(map[string][]string)
	for _, edge := range edges {
		adjList[edge.SourceID] = append(adjList[edge.SourceID], edge.TargetID)
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycleUtil func(nodeID string) bool
	hasCycleUtil = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, neighbor := range adjList[nodeID] {
			if !visited[neighbor] {
				if hasCycleUtil(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[nodeID] = false
		return false
	}

	for _, node := range nodes {
		if !visited[node.ID] {
			if hasCycleUtil(node.ID) {
				return true
			}
		}
	}

	return false
}

// checkOrphanedNodes checks for nodes that are not connected to the graph
func (s *AutomationService) checkOrphanedNodes(nodes []models.AutomationNode, edges []models.AutomationNodeEdge) error {
	if len(nodes) == 1 {
		return nil // Single node is valid
	}

	// Build sets of connected nodes
	connectedNodes := make(map[string]bool)
	for _, edge := range edges {
		connectedNodes[edge.SourceID] = true
		connectedNodes[edge.TargetID] = true
	}

	// Check each node
	for _, node := range nodes {
		// START nodes must have outgoing edges
		// EXIT nodes must have incoming edges
		// Other nodes must be connected
		if node.Type == models.NodeTypeStart || node.Type == models.NodeTypeExit {
			if !connectedNodes[node.ID] {
				return fmt.Errorf("node %s (%s) is not connected to the graph", node.ID, node.Type)
			}
		} else {
			if !connectedNodes[node.ID] {
				return fmt.Errorf("orphaned node detected: %s (%s)", node.ID, node.Type)
			}
		}
	}

	return nil
}

// Activate activates an automation after validation
func (s *AutomationService) Activate(ctx context.Context, id string) error {
	automation, err := s.GetWithNodesAndEdges(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load automation: %w", err)
	}

	// Validate graph structure
	if err := s.ValidateGraph(automation); err != nil {
		return fmt.Errorf("automation validation failed: %w", err)
	}

	if err := s.ValidateConfiguration(automation); err != nil {
		return err
	}

	// Update status
	if err := s.db.WithContext(ctx).
		Model(&models.Automation{}).
		Where("id = ?", id).
		Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to activate automation: %w", err)
	}

	automationLog.Success("Activated automation %s", id)
	return nil
}

// Deactivate safely deactivates a running automation
func (s *AutomationService) Deactivate(ctx context.Context, id string) error {

	// Update status
	if err := s.db.WithContext(ctx).
		Model(&models.Automation{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate automation: %w", err)
	}

	// Note: Running executions will continue, but new executions won't start
	automationLog.Info("Deactivated automation %s", id)
	return nil
}

// GetActiveAutomations returns all active automations for event-based triggers
func (s *AutomationService) GetActiveAutomations(ctx context.Context, teamID string) ([]models.Automation, error) {
	var automations []models.Automation

	if err := s.db.WithContext(ctx).
		Where("team_id = ? AND is_active = ? AND is_deleted = ?", teamID, true, false).
		Preload("Nodes", "is_deleted = ?", false).
		Preload("Edges").
		Find(&automations).Error; err != nil {
		return nil, err
	}

	return automations, nil
}
