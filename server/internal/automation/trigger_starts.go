package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kori/internal/models"
	"kori/internal/services"
	"kori/internal/tasks"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Triggers that start one contact at a time, matched against the START step's settings.
const (
	TriggerTagAdded       = "contact.tag_added"
	TriggerListJoined     = "list.subscriber_added"
	TriggerSegmentEntered = "segment.entered"
	TriggerDateReached    = "date.reached"
	TriggerEvent          = "event.received"
)

// DateFieldPattern limits date triggers to createdAt or a custom field key.
var DateFieldPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)

// matchingAutomations returns the team's live automations on trigger whose START
// step settings satisfy accepts.
func (tm *TriggerManager) matchingAutomations(ctx context.Context, teamID, trigger string, accepts func(models.Automation, map[string]interface{}) bool) ([]models.Automation, error) {
	var automations []models.Automation
	if err := tm.db.WithContext(ctx).Preload("Nodes", "type = ? AND is_deleted = ?", models.NodeTypeStart, false).
		Where("team_id = ? AND trigger_event = ? AND is_active = ? AND is_deleted = ?", teamID, trigger, true, false).
		Order("id").Find(&automations).Error; err != nil {
		return nil, err
	}
	matched := automations[:0]
	for _, automation := range automations {
		if start, ok := startSettings(automation); ok && accepts(automation, start) {
			matched = append(matched, automation)
		}
	}
	return matched, nil
}

func startSettings(automation models.Automation) (map[string]interface{}, bool) {
	start := map[string]interface{}{}
	if len(automation.Nodes) != 1 {
		return nil, false
	}
	if len(automation.Nodes[0].Data) > 0 && json.Unmarshal(automation.Nodes[0].Data, &start) != nil {
		return nil, false
	}
	return start, true
}

// setting reports whether the START step names exactly value for key. An
// explicit choice is required so a trigger never fires workspace-wide by accident.
func setting(start map[string]interface{}, key, value string) bool {
	chosen, _ := start[key].(string)
	return chosen != "" && chosen == value
}

// startFor enqueues the automation for the contact. The execution ID is derived
// from key, so one occurrence never starts the same automation twice for a
// contact, however often it is delivered.
func (tm *TriggerManager) startFor(ctx context.Context, automation models.Automation, contactID, key string, data map[string]interface{}) error {
	variables := map[string]interface{}{"event": automation.TriggerEvent}
	for name, value := range data {
		variables[name] = value
	}
	return tm.taskClient.EnqueueAutomationTask(ctx, tasks.AutomationExecuteTask{
		ExecutionID:  uuid.NewSHA1(uuid.NameSpaceOID, []byte("xem:"+key+":"+automation.ID)).String(),
		AutomationID: automation.ID,
		ContactID:    contactID,
		TriggerData:  variables,
	}, 0)
}

func (tm *TriggerManager) logFailure(err error) {
	if err != nil {
		triggerLog.Error("Failed to start automations: %v", err)
	}
}

func (tm *TriggerManager) handleListJoined(ctx context.Context, contact *models.Contact) error {
	if contact.ListID == "" {
		return nil
	}
	automations, err := tm.matchingAutomations(ctx, contact.TeamID, TriggerListJoined, func(_ models.Automation, start map[string]interface{}) bool {
		return setting(start, "listId", contact.ListID)
	})
	if err != nil {
		return err
	}
	key := fmt.Sprintf("list:%s:%s:%d", contact.ID, contact.ListID, contact.UpdatedAt.UnixNano())
	var failures []error
	for _, automation := range automations {
		failures = append(failures, tm.startFor(ctx, automation, contact.ID, key, map[string]interface{}{"list_id": contact.ListID}))
	}
	return errors.Join(failures...)
}

func (tm *TriggerManager) handleTagAdded(ctx context.Context, added *models.TagAdded) error {
	automations, err := tm.matchingAutomations(ctx, added.Contact.TeamID, TriggerTagAdded, func(_ models.Automation, start map[string]interface{}) bool {
		return setting(start, "tagId", added.Tag.ID)
	})
	if err != nil {
		return err
	}
	key := fmt.Sprintf("tag:%s:%s:%d", added.Contact.ID, added.Tag.ID, added.At.UnixNano())
	var failures []error
	for _, automation := range automations {
		failures = append(failures, tm.startFor(ctx, automation, added.Contact.ID, key, map[string]interface{}{"tag_id": added.Tag.ID, "tag_name": added.Tag.Name}))
	}
	return errors.Join(failures...)
}

func (tm *TriggerManager) handleSegmentEntered(ctx context.Context, entry *models.SegmentEntry) error {
	automations, err := tm.matchingAutomations(ctx, entry.TeamID, TriggerSegmentEntered, func(automation models.Automation, start map[string]interface{}) bool {
		// Only automations already live when the entry window began; otherwise
		// switching one on would start it for everyone already in the segment.
		return setting(start, "segmentId", entry.SegmentID) && !automation.UpdatedAt.After(entry.Since)
	})
	if err != nil {
		return err
	}
	var failures []error
	for _, contactID := range entry.ContactIDs {
		key := fmt.Sprintf("segment:%s:%s:%d", entry.SegmentID, contactID, entry.Since.UnixNano())
		for _, automation := range automations {
			failures = append(failures, tm.startFor(ctx, automation, contactID, key, map[string]interface{}{"segment_id": entry.SegmentID}))
		}
	}
	return errors.Join(failures...)
}

func (tm *TriggerManager) handleContactAppEvent(ctx context.Context, event *models.ContactEvent) error {
	automations, err := tm.matchingAutomations(ctx, event.TeamID, TriggerEvent, func(_ models.Automation, start map[string]interface{}) bool {
		name, _ := start["eventName"].(string)
		return name != "" && strings.EqualFold(strings.TrimSpace(name), event.Name)
	})
	if err != nil || len(automations) == 0 {
		return err
	}
	data := map[string]interface{}{"event_name": event.Name, "event_id": event.ID}
	var properties map[string]interface{}
	if len(event.Properties) > 0 && json.Unmarshal(event.Properties, &properties) != nil {
		return fmt.Errorf("invalid properties on event %s", event.ID)
	}
	// Scalar properties become event_<name> variables for conditions and merge tags.
	for name, value := range properties {
		variable := "event_" + name
		if _, reserved := data[variable]; reserved {
			continue
		}
		switch v := value.(type) {
		case string:
			data[variable] = v
		case float64:
			data[variable] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			data[variable] = strconv.FormatBool(v)
		}
	}
	var failures []error
	for _, automation := range automations {
		failures = append(failures, tm.startFor(ctx, automation, event.ContactID, "event:"+event.ID, data))
	}
	return errors.Join(failures...)
}

// RefreshSegmentTriggers re-evaluates every dynamic segment that a live
// "entered segment" automation watches. Each refresh announces its new members.
func (tm *TriggerManager) RefreshSegmentTriggers(ctx context.Context) error {
	var automations []models.Automation
	if err := tm.db.WithContext(ctx).Preload("Nodes", "type = ? AND is_deleted = ?", models.NodeTypeStart, false).
		Where("trigger_event = ? AND is_active = ? AND is_deleted = ?", TriggerSegmentEntered, true, false).Find(&automations).Error; err != nil {
		return err
	}
	segments := map[string]bool{}
	for _, automation := range automations {
		if start, ok := startSettings(automation); ok {
			if id, _ := start["segmentId"].(string); id != "" {
				var dynamic int64
				if err := tm.db.WithContext(ctx).Model(&models.Segment{}).Where("id = ? AND team_id = ? AND is_dynamic = ? AND is_deleted = ?", id, automation.TeamID, true, false).Count(&dynamic).Error; err != nil {
					return err
				}
				segments[id] = dynamic == 1
			}
		}
	}
	service := services.NewSegmentService(tm.db)
	var failures []error
	for id, dynamic := range segments {
		if dynamic {
			failures = append(failures, service.RefreshSegment(ctx, id))
		}
	}
	return errors.Join(failures...)
}

// RunDateTriggers starts "date reached" automations for contacts whose date
// field, shifted by the step's offset, is today in the workspace's time zone.
// It runs hourly; the first run on or after 09:00 local time starts them, and the
// execution ID (automation, contact, date) keeps later runs from starting them again.
func (tm *TriggerManager) RunDateTriggers(ctx context.Context, now time.Time) error {
	var automations []models.Automation
	if err := tm.db.WithContext(ctx).Preload("Nodes", "type = ? AND is_deleted = ?", models.NodeTypeStart, false).
		Where("trigger_event = ? AND is_active = ? AND is_deleted = ?", TriggerDateReached, true, false).Order("id").Find(&automations).Error; err != nil {
		return err
	}
	var failures []error
	for _, automation := range automations {
		start, ok := startSettings(automation)
		if !ok {
			continue
		}
		field, _ := start["field"].(string)
		offset, _ := start["offsetDays"].(float64)
		yearly, _ := start["yearly"].(bool)
		if !DateFieldPattern.MatchString(field) || offset != float64(int(offset)) || offset < -365 || offset > 365 {
			continue
		}
		local := now.In(tm.workspaceLocation(ctx, automation.TeamID))
		if local.Hour() < 9 {
			continue
		}
		// "3 days after" fires today for dates 3 days ago.
		anchor := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -int(offset))
		value := "SUBSTR(CAST(contacts.created_at AS TEXT), 1, 10)"
		args := []interface{}{}
		if field != "createdAt" {
			value = "SUBSTR(contacts.metadata ->> CAST(? AS TEXT), 1, 10)"
			args = append(args, field)
		}
		query := tm.db.WithContext(ctx).Model(&models.Contact{}).
			Where("team_id = ? AND is_deleted = ? AND status = ?", automation.TeamID, false, models.SubscriberStatusActive)
		day := anchor.Format("2006-01-02")
		if yearly {
			days := []string{anchor.Format("01-02")}
			if anchor.Month() == time.February && anchor.Day() == 28 && anchor.AddDate(0, 0, 1).Day() == 1 {
				days = append(days, "02-29") // leap-day dates come round on 28 February
			}
			query = query.Where("SUBSTR("+value+", 6, 5) IN ?", append(append([]interface{}{}, args...), days)...).
				Where(value+" <= ?", append(append([]interface{}{}, args...), day)...)
		} else {
			query = query.Where(value+" = ?", append(args, day)...)
		}
		var ids []string
		if err := query.Order("id").Pluck("id", &ids).Error; err != nil {
			failures = append(failures, err)
			continue
		}
		for _, id := range ids {
			key := fmt.Sprintf("date:%s:%s", id, day)
			failures = append(failures, tm.startFor(ctx, automation, id, key, map[string]interface{}{"date_field": field, "date": day}))
		}
	}
	return errors.Join(failures...)
}

// workspaceLocation is the team's default sending time zone, or UTC.
func (tm *TriggerManager) workspaceLocation(ctx context.Context, teamID string) *time.Location {
	var zones []string
	tm.db.WithContext(ctx).Model(&models.TeamSendTimeDefaults{}).Where("team_id = ?", teamID).Limit(1).Pluck("default_timezone", &zones)
	if len(zones) == 1 && zones[0] != "" {
		if loc, err := time.LoadLocation(zones[0]); err == nil {
			return loc
		}
	}
	return time.UTC
}
