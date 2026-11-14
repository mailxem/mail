package marketing

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"kori/internal/models"
	"kori/internal/utils"
)

// ReachedSection is a section and how many readers get it in their email.
type ReachedSection struct {
	utils.Section
	Readers int `json:"readers"`
}

// SectionReach is how the readers of a send get its sections.
type SectionReach struct {
	Sections []ReachedSection `json:"sections"`
	Readers  int              `json:"readers"`
	Empty    int              `json:"empty"` // readers whose email would show nothing
}

// Reach works out how the contacts the queries match get body's sections.
func Reach(db *gorm.DB, team, body string, contacts ...*gorm.DB) (SectionReach, error) {
	var ids []string
	for _, q := range contacts {
		var page []string
		if err := db.Session(&gorm.Session{NewDB: true}).Table("(?) AS r", q.Select("id")).Pluck("id", &page).Error; err != nil {
			return SectionReach{}, err
		}
		ids = append(ids, page...)
	}
	picker, err := NewSectionPicker(db, team, body)
	if err != nil {
		return SectionReach{}, err
	}
	out := SectionReach{Sections: make([]ReachedSection, len(picker.sections)), Readers: len(ids)}
	for i, s := range picker.sections {
		out.Sections[i].Section = s
	}
	for _, id := range ids {
		data := map[string]string{}
		if !picker.Pick(id, data) {
			out.Empty++
		}
		for f := range strings.SplitSeq(data[utils.SectionsKey], ",") {
			if n, err := strconv.Atoi(f); err == nil {
				out.Sections[n].Readers++
			}
		}
	}
	return out, nil
}

// ListReach is Reach for the active readers of a list.
func ListReach(db *gorm.DB, team, list, body string) (SectionReach, error) {
	return Reach(db, team, body, Scope(db.Model(&models.Contact{}), team).Where("list_id = ? AND status = ?", list, models.SubscriberStatusActive))
}

// EmptyEmailReason is why a send was paused before anyone got an empty email.
func EmptyEmailReason(readers int) string {
	who := fmt.Sprintf("%d readers", readers)
	if readers == 1 {
		who = "1 reader"
	}
	return who + " would get this email with nothing in it, because none of its sections are for them. Change who the sections are for, then resume."
}

// SectionPicker picks each reader's sections for a send of body, and spots
// readers whose email would show nothing.
type SectionPicker struct {
	body     string
	sections []utils.Section
	audience *utils.SectionAudience
	empty    map[string]bool // by plan, so each plan is checked once
}

// NewSectionPicker loads who body's sections are for.
func NewSectionPicker(db *gorm.DB, team, body string) (*SectionPicker, error) {
	sections := utils.ParseSections(body)
	audience, err := utils.LoadSectionAudience(db, team, sections)
	if err != nil {
		return nil, err
	}
	return &SectionPicker{body: body, sections: sections, audience: audience, empty: map[string]bool{}}, nil
}

// Pick adds contact's sections to their email data and reports whether the
// email would show them anything.
func (p *SectionPicker) Pick(contact string, data map[string]string) bool {
	plan := ""
	if len(p.sections) > 0 {
		plan = utils.SectionPlan(p.sections, p.audience.Reader(contact))
		data[utils.SectionsKey] = plan
	}
	empty, seen := p.empty[plan]
	if !seen {
		empty = !utils.HasContent(utils.EmailSections(p.body, map[string]string{utils.SectionsKey: plan}))
		p.empty[plan] = empty
	}
	return !empty
}
