package tasks

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestStageTargets(t *testing.T) {
	for _, tc := range []struct {
		total, sent, target int
		last                bool
	}{
		{10000, 0, 500, false},     // 5% of 10,000
		{10000, 500, 2500, false},  // then 25%
		{10000, 2500, 10000, true}, // then everyone
		{1000, 0, 100, false},      // at least 100 first
		{1000, 100, 500, false},    // at least 500 next
		{1000, 500, 1000, true},
		{600, 0, 100, false},
		{600, 100, 500, false},
		{600, 500, 600, true},
	} {
		target, last := stageTarget(tc.total, tc.sent)
		if target != tc.target || last != tc.last {
			t.Errorf("stageTarget(%d, %d) = %d, %v; want %d, %v", tc.total, tc.sent, target, last, tc.target, tc.last)
		}
	}
}

func TestStageHealth(t *testing.T) {
	for _, tc := range []struct {
		sent, failed, complained int64
		pause                    bool
	}{
		{100, 4, 0, false},  // 4% failed
		{100, 5, 0, true},   // 5% failed
		{40, 10, 0, false},  // sample too small to judge failures
		{500, 0, 1, false},  // one complaint is not a pattern
		{2000, 0, 2, true},  // 0.1% with at least 2
		{3000, 0, 2, false}, // under 0.1%
	} {
		if got := stageHealth(tc.sent, tc.failed, tc.complained) != ""; got != tc.pause {
			t.Errorf("stageHealth(%d, %d, %d) paused=%v; want %v", tc.sent, tc.failed, tc.complained, got, tc.pause)
		}
	}
}

func TestUseStagesForNeverMailedListsOnly(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Campaign{}); err != nil {
		t.Fatal(err)
	}
	on, off := true, false
	list := "list-1"
	c := &models.Campaign{Base: models.Base{ID: "c-new"}, ListID: list, Schedule: models.CampaignScheduleOneTime}
	check := func(name string, got, want bool) {
		if got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
	check("never-mailed large list", useStages(db, c, 5000), true)
	check("small list", useStages(db, c, 400), false)
	check("explicitly off", useStages(db, &models.Campaign{Base: models.Base{ID: "c2"}, ListID: list, SendInStages: &off}, 5000), false)
	check("recurring", useStages(db, &models.Campaign{Base: models.Base{ID: "c3"}, ListID: list, Schedule: models.CampaignScheduleRecurring}, 5000), false)
	earlier := models.Campaign{Base: models.Base{ID: "c-old"}, ListID: list, Status: models.CampaignStatusCompleted, TemplateID: "t", TeamID: "team"}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&earlier).Error; err != nil {
		t.Fatal(err)
	}
	check("list mailed before", useStages(db, c, 5000), false)
	check("explicitly on", useStages(db, &models.Campaign{Base: models.Base{ID: "c4"}, ListID: list, SendInStages: &on}, 5000), true)
}
