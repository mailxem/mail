package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// KnowledgeBase builds context from analytics data for AI prompts
type KnowledgeBase struct {
	db         *gorm.DB
	aggregator *AnalyticsAggregator
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase(db *gorm.DB) *KnowledgeBase {
	return &KnowledgeBase{
		db:         db,
		aggregator: NewAnalyticsAggregator(db),
	}
}

// BuildContactContext builds comprehensive context about a contact for AI decision-making
func (kb *KnowledgeBase) BuildContactContext(ctx context.Context, contactID string) (string, error) {
	insights, err := kb.aggregator.GetContactInsights(ctx, contactID)
	if err != nil {
		return "", fmt.Errorf("failed to get contact insights: %w", err)
	}

	var contextParts []string

	// Contact basic info
	contextParts = append(contextParts, fmt.Sprintf("Contact: %s", insights.Email))

	// Engagement metrics
	contextParts = append(contextParts, fmt.Sprintf("Engagement Score: %.1f/100", insights.EngagementScore))
	contextParts = append(contextParts, fmt.Sprintf("Total Emails Received: %d", insights.TotalEmailsReceived))
	contextParts = append(contextParts, fmt.Sprintf("Emails Opened: %d (%.1f%%)",
		insights.EmailsOpened,
		safePercent(insights.EmailsOpened, insights.TotalEmailsReceived)))
	contextParts = append(contextParts, fmt.Sprintf("Emails Clicked: %d (%.1f%%)",
		insights.EmailsClicked,
		safePercent(insights.EmailsClicked, insights.TotalEmailsReceived)))

	// Behavioral insights
	if insights.PreferredDevice != "" {
		contextParts = append(contextParts, fmt.Sprintf("Preferred Device: %s", insights.PreferredDevice))
	}
	if insights.PreferredTime != "" {
		contextParts = append(contextParts, fmt.Sprintf("Most Active Time: %s", insights.PreferredTime))
	}

	// Recency
	if insights.LastOpenedAt != nil {
		timeSince := time.Since(*insights.LastOpenedAt)
		contextParts = append(contextParts, fmt.Sprintf("Last Opened: %s ago", formatDuration(timeSince)))
	}
	if insights.LastClickedAt != nil {
		timeSince := time.Since(*insights.LastClickedAt)
		contextParts = append(contextParts, fmt.Sprintf("Last Clicked: %s ago", formatDuration(timeSince)))
	}

	// Segment classification
	segment := classifyEngagement(insights.EngagementScore)
	contextParts = append(contextParts, fmt.Sprintf("Engagement Segment: %s", segment))

	return strings.Join(contextParts, "\n"), nil
}

// BuildCampaignContext builds context about campaign performance
func (kb *KnowledgeBase) BuildCampaignContext(ctx context.Context, campaignID string) (string, error) {
	analytics, err := kb.aggregator.GetCampaignAnalytics(ctx, campaignID)
	if err != nil {
		return "", fmt.Errorf("failed to get campaign analytics: %w", err)
	}

	var contextParts []string

	contextParts = append(contextParts, fmt.Sprintf("Campaign: %s", analytics.CampaignName))
	contextParts = append(contextParts, fmt.Sprintf("Total Contacts: %d", analytics.TotalContacts))
	contextParts = append(contextParts, fmt.Sprintf("Emails Sent: %d", analytics.EmailsSent))
	contextParts = append(contextParts, fmt.Sprintf("Open Rate: %.2f%%", analytics.OpenRate))
	contextParts = append(contextParts, fmt.Sprintf("Click Rate: %.2f%%", analytics.ClickRate))

	if analytics.TopPerformingTime != "" {
		contextParts = append(contextParts, fmt.Sprintf("Best Send Time: %s", analytics.TopPerformingTime))
	}
	if analytics.BestDevice != "" {
		contextParts = append(contextParts, fmt.Sprintf("Most Used Device: %s", analytics.BestDevice))
	}

	// Performance assessment
	performance := assessCampaignPerformance(analytics.OpenRate, analytics.ClickRate)
	contextParts = append(contextParts, fmt.Sprintf("Performance: %s", performance))

	return strings.Join(contextParts, "\n"), nil
}

// BuildTeamContext builds context about overall team performance
func (kb *KnowledgeBase) BuildTeamContext(ctx context.Context, teamID string, days int) (string, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	scope := AnalyticsScope{
		TeamID:    teamID,
		StartDate: startDate,
	}

	analytics, err := kb.aggregator.GetEmailAnalytics(ctx, scope)
	if err != nil {
		return "", fmt.Errorf("failed to get team analytics: %w", err)
	}

	var contextParts []string

	contextParts = append(contextParts, fmt.Sprintf("Team Performance (Last %d days):", days))
	contextParts = append(contextParts, fmt.Sprintf("Total Emails Sent: %d", analytics.TotalSent))
	contextParts = append(contextParts, fmt.Sprintf("Average Open Rate: %.2f%%", analytics.OpenRate))
	contextParts = append(contextParts, fmt.Sprintf("Average Click Rate: %.2f%%", analytics.ClickRate))
	contextParts = append(contextParts, fmt.Sprintf("Bounce Rate: %.2f%%", analytics.BounceRate))

	// Benchmarking
	benchmark := benchmarkPerformance(analytics.OpenRate, analytics.ClickRate)
	contextParts = append(contextParts, benchmark)

	return strings.Join(contextParts, "\n"), nil
}

// BuildDecisionContext combines multiple contexts for AI decision-making
func (kb *KnowledgeBase) BuildDecisionContext(ctx context.Context, contactID, automationID string) (string, error) {
	var contextParts []string

	// Contact context
	if contactID != "" {
		contactCtx, err := kb.BuildContactContext(ctx, contactID)
		if err == nil {
			contextParts = append(contextParts, "=== Contact Information ===")
			contextParts = append(contextParts, contactCtx)
		}
	}

	// Could add automation performance context here
	// if automationID != "" { ... }

	return strings.Join(contextParts, "\n\n"), nil
}

// Helper functions

func safePercent(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	} else {
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
}

func classifyEngagement(score float64) string {
	switch {
	case score >= 75:
		return "Highly Engaged"
	case score >= 50:
		return "Moderately Engaged"
	case score >= 25:
		return "Lightly Engaged"
	default:
		return "Disengaged"
	}
}

func assessCampaignPerformance(openRate, clickRate float64) string {
	// Industry benchmarks: ~20% open rate, ~2.5% click rate
	if openRate >= 25 && clickRate >= 3 {
		return "Excellent - Above Industry Average"
	} else if openRate >= 18 && clickRate >= 2 {
		return "Good - Near Industry Average"
	} else if openRate >= 10 {
		return "Average - Room for Improvement"
	} else {
		return "Below Average - Needs Optimization"
	}
}

func benchmarkPerformance(openRate, clickRate float64) string {
	context := []string{"Industry Benchmarks: Open Rate ~20%, Click Rate ~2.5%"}

	if openRate > 20 {
		context = append(context, "✓ Open rate is above industry average")
	} else {
		context = append(context, "✗ Open rate is below industry average")
	}

	if clickRate > 2.5 {
		context = append(context, "✓ Click rate is above industry average")
	} else {
		context = append(context, "✗ Click rate is below industry average")
	}

	return strings.Join(context, "\n")
}
