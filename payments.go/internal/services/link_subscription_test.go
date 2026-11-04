package services

import (
	"context"
	"testing"

	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
	"github.com/mailxem/payments.go/pkg/dodo"
	"gorm.io/gorm"
)

type fakeSubscriptions struct {
	repositories.SubscriptionRepository
	byID map[string]*models.Subscription
}

func (f *fakeSubscriptions) GetByDodoSubscriptionID(_ context.Context, dodoID string) (*models.Subscription, error) {
	for _, s := range f.byID {
		if s.DodoSubscriptionID != nil && *s.DodoSubscriptionID == dodoID {
			return s, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeSubscriptions) GetByID(_ context.Context, id string) (*models.Subscription, error) {
	if s, ok := f.byID[id]; ok {
		return s, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (f *fakeSubscriptions) Update(_ context.Context, s *models.Subscription) error {
	f.byID[s.ID] = s
	return nil
}

func TestHostedCheckoutWebhooksLinkDodosSubscriptionOnce(t *testing.T) {
	other := "sub_other"
	repo := &fakeSubscriptions{byID: map[string]*models.Subscription{
		"sub-1": {ID: "sub-1"},
		"sub-2": {ID: "sub-2", DodoSubscriptionID: &other},
	}}
	s := &webhookService{subscriptionRepo: repo}
	ctx := context.Background()

	linked, err := s.linkSubscription(ctx, "sub_dodo", map[string]interface{}{dodo.XemSubscriptionIDKey: "sub-1"})
	if err != nil || linked.ID != "sub-1" || *repo.byID["sub-1"].DodoSubscriptionID != "sub_dodo" {
		t.Fatalf("first event should link sub-1, got %+v, %v", linked, err)
	}
	if again, err := s.linkSubscription(ctx, "sub_dodo", nil); err != nil || again.ID != "sub-1" {
		t.Fatalf("later events should find it by Dodo ID, got %+v, %v", again, err)
	}
	if _, err := s.linkSubscription(ctx, "sub_dodo_2", map[string]interface{}{dodo.XemSubscriptionIDKey: "sub-2"}); err == nil {
		t.Fatal("a subscription linked to another Dodo ID must not be relinked")
	}
	if _, err := s.linkSubscription(ctx, "sub_unknown", nil); err == nil {
		t.Fatal("an unknown subscription without metadata should not be found")
	}
}
