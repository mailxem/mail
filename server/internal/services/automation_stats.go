package services

import (
	"context"
	"time"
)

// StepStats counts contacts who started the automation in the range, per step.
// Email counts are per message; rates are over sent. Delivered is nil when no
// message from the step went through a sender that reports delivery.
type StepStats struct {
	NodeID          string  `json:"nodeId"`
	Entered         int64   `json:"entered"`
	Waiting         int64   `json:"waiting"`
	MovedOn         int64   `json:"movedOn"`
	Exited          int64   `json:"exited"`
	Conversions     int64   `json:"conversions"`
	Sent            int64   `json:"sent"`
	Delivered       *int64  `json:"delivered"`
	Opens           int64   `json:"opens"`
	Clicks          int64   `json:"clicks"`
	Unsubscribes    int64   `json:"unsubscribes"`
	OpenRate        float64 `json:"openRate"`
	ClickRate       float64 `json:"clickRate"`
	UnsubscribeRate float64 `json:"unsubscribeRate"`
	DeliveryKnown   int64   `json:"-"`
	DeliveredCount  int64   `json:"-"`
}

// Executions log each step they pass; automation emails carry the delivery
// key automation:<execution>:<step>, which hits the unique index on emails.
const stepStatsSQL = `WITH x AS (
 SELECT id, status, current_node_id, execution_log FROM automation_executions
 WHERE automation_id = @automation AND is_deleted = false AND started_at >= @from AND started_at < @to
), visits AS (
 SELECT DISTINCT x.id, l->>'nodeId' node_id FROM x CROSS JOIN LATERAL jsonb_array_elements(x.execution_log) l
), converted AS (
 SELECT DISTINCT execution_id FROM goal_conversions WHERE automation_id = @automation AND execution_id IS NOT NULL AND is_deleted = false
), steps AS (
 SELECT v.node_id,
 COUNT(*) entered,
 COUNT(*) FILTER (WHERE x.status = 'WAITING' AND x.current_node_id::text = v.node_id) waiting,
 COUNT(*) FILTER (WHERE x.status IN ('COMPLETED', 'FAILED') AND x.current_node_id::text = v.node_id) exited,
 COUNT(c.execution_id) conversions
 FROM visits v JOIN x ON x.id = v.id LEFT JOIN converted c ON c.execution_id = v.id GROUP BY v.node_id
), messages AS (
 SELECT v.node_id, e.id, e.status FROM visits v JOIN emails e ON e.delivery_key = 'automation:' || v.id::text || ':' || v.node_id
), outcomes AS (
 SELECT m.node_id,
 COUNT(*) FILTER (WHERE m.status IN ('SENT', 'OPENED', 'CLICKED', 'BOUNCED')) sent,
 COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM email_trackings t WHERE t.email_id = m.id AND t.event = 'open' AND t.is_deleted = false)) opens,
 COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM email_trackings t WHERE t.email_id = m.id AND t.event = 'click' AND t.is_deleted = false)) clicks,
 COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM email_trackings t WHERE t.email_id = m.id AND t.event = 'unsubscribe' AND t.is_deleted = false)) unsubscribes,
 COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM managed_messages mm WHERE mm.email_id = m.id::text)) delivery_known,
 COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM managed_messages mm WHERE mm.email_id = m.id::text AND mm.status IN ('DELIVERED', 'COMPLAINED'))) delivered_count
 FROM messages m GROUP BY m.node_id
)
SELECT s.node_id, s.entered, s.waiting, s.exited, s.conversions,
 COALESCE(o.sent, 0) sent, COALESCE(o.opens, 0) opens, COALESCE(o.clicks, 0) clicks, COALESCE(o.unsubscribes, 0) unsubscribes,
 COALESCE(o.delivery_known, 0) delivery_known, COALESCE(o.delivered_count, 0) delivered_count
FROM steps s LEFT JOIN outcomes o ON o.node_id = s.node_id`

// StepStats returns results for every step that contacts reached. PostgreSQL only.
func (s *AutomationService) StepStats(ctx context.Context, automationID string, from, to time.Time) ([]StepStats, error) {
	stats := []StepStats{}
	err := s.db.WithContext(ctx).Raw(stepStatsSQL, map[string]interface{}{"automation": automationID, "from": from, "to": to}).Scan(&stats).Error
	for i := range stats {
		st := &stats[i]
		st.MovedOn = st.Entered - st.Waiting - st.Exited
		if st.DeliveryKnown > 0 {
			st.Delivered = &st.DeliveredCount
		}
		if st.Sent > 0 {
			sent := float64(st.Sent)
			st.OpenRate, st.ClickRate, st.UnsubscribeRate = float64(st.Opens)/sent, float64(st.Clicks)/sent, float64(st.Unsubscribes)/sent
		}
	}
	return stats, err
}
