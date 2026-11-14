package onboardingemails

import "strings"

// SummaryEmail is the daily or weekly sending summary. The notifications
// worker computes every line; this only lays it out in the shared brand frame,
// which escapes each value.
type SummaryEmail struct {
	Frequency, Workspace, Headline string
	Details                        [][2]string
}

func (s SummaryEmail) Render(link string) (subject, html, text string) {
	unit := "week"
	if s.Frequency == "daily" {
		unit = "day"
	}
	t := Template{Key: "summary", Subject: "Your " + s.Frequency + " summary for " + strings.Join(strings.Fields(s.Workspace), " "), Heading: "Your " + unit + " in email.", Description: s.Headline, Next: "Open the overview for charts, every campaign and the full numbers."}
	p := presentation{Label: strings.ToUpper(s.Frequency) + " SUMMARY", Action: "Open your overview", Footer: "You get this summary through your workspace’s notification settings, where you can change how often it arrives or turn it off."}
	for _, d := range s.Details {
		p.Details = append(p.Details, detail{d[0], d[1]})
	}
	html, text = t.renderWith(p, link)
	return t.Subject, html, text
}
