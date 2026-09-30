// Package onboardingemails owns the transactional copy used by the sender and
// the editable starter library. It has no delivery or tenant-data dependencies.
package onboardingemails

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type Template struct{ Key, Subject, Heading, Description, Next string }

var Templates = []Template{
	{"domain_added", "Let’s connect your sending domain", "Your domain. Your next chapter.", "Your sending domain has been added to Xem. There are a few DNS records to put in place before it can send email.", "Open managed sending and add the ownership record to your DNS provider."},
	{"ownership_verified", "Your domain ownership is verified", "That domain is yours.", "Xem has verified your domain ownership. Next, we’ll check the records that help receiving servers trust your email.", "Add the DKIM, MAIL FROM and DMARC records shown in your workspace."},
	{"domain_ready", "Your sending domain passed its checks", "Looking good, DNS.", "Your domain passed the required authentication checks. Sending still depends on your workspace approval and current account status.", "Review your approval status and choose the address you want to send from."},
	{"approved", "Your workspace is approved for managed sending", "You’re approved.", "Your workspace has been approved for Xem managed sending. Your current sending allowance is available in the dashboard.", "Connect a sender on a verified domain, then send a test to your own inbox."},
	{"sender_connected", "Your managed sender is connected", "Say hello from your own domain.", "Your approved workspace now has a connected managed sender. It is available when composing emails and setting up campaigns.", "Send a test to your own inbox and check its delivery status before your first campaign."},
	{"test_queued", "Your sending test is queued", "One small send. A useful first step.", "Xem has queued your test email. Queued means it is waiting to be submitted to the email provider.", "Follow the test in managed sending. This notification does not confirm delivery."},
	{"test_accepted", "The provider accepted your sending test", "Your test is on its way.", "The email provider accepted your test for sending. This is not yet confirmation that the receiving server accepted it.", "Check delivery events in your workspace and look in your inbox and spam folder."},
	{"test_delivered", "Your test reached the receiving mail server", "A successful first connection.", "The provider reported that the receiving mail server accepted your test email. Inbox placement can still vary.", "Review the email in your mailbox. You’re ready to prepare your first campaign."},
	{"test_attention", "Your sending test needs attention", "Let’s check that test.", "Your test did not reach a confirmed successful outcome. Its latest status and next steps are available in your workspace.", "Review the message status before sending another test, especially if its outcome is unknown."},
}

func Find(key string) (Template, bool) {
	for _, t := range Templates {
		if t.Key == key {
			return t, true
		}
	}
	return Template{}, false
}

func (t Template) fragments(workspace, domain, link string) []string {
	return []string{
		`<p style="margin:0;font-size:28px;font-weight:800;letter-spacing:-1.5px;color:#181a16">xem<span style="color:#728820">.</span></p>`,
		`<p style="margin:0;color:#6b7165;font-size:11px;letter-spacing:2px">YOUR MANAGED SENDING JOURNEY</p>`,
		`<h1 style="margin:0;font-size:34px;line-height:1.15;letter-spacing:-1px;color:#181a16">` + html.EscapeString(t.Heading) + `</h1>`,
		`<p style="margin:0;color:#50574b;line-height:1.75">` + html.EscapeString(t.Description) + `</p>`,
		`<p style="margin:0;padding:16px;background:#f4f6ef;border-radius:8px;color:#42483d;font-size:13px">Workspace: <strong>` + html.EscapeString(workspace) + `</strong><br>Domain: <strong>` + html.EscapeString(domain) + `</strong></p>`,
		`<p style="margin:0;color:#50574b;line-height:1.75"><strong style="color:#181a16">Up next</strong><br>` + html.EscapeString(t.Next) + `</p>`,
		`<p style="margin:8px 0"><a href="` + html.EscapeString(link) + `" style="display:inline-block;padding:15px 24px;background:#d2f159;color:#1c2413;font-weight:600;text-decoration:none;border-radius:8px">Open managed sending &rarr;</a></p>`,
		`<p style="margin:0;font-size:12px;line-height:1.7;color:#858b7e">A little less setup. A lot more possibility.<br>You received this service update as the workspace owner.<br>Built with Xem · Email that feels like you.</p>`,
	}
}
func (t Template) Render(workspace, domain, link string) (string, string) {
	parts := t.fragments(workspace, domain, link)
	var rows strings.Builder
	for _, p := range parts {
		rows.WriteString(`<tr><td style="padding:12px 32px">` + p + `</td></tr>`)
	}
	body := `<!doctype html><html><head><meta name="viewport" content="width=device-width,initial-scale=1"><meta charset="utf-8"><title>` + html.EscapeString(t.Subject) + `</title></head><body style="margin:0;background:#edf0e8;font-family:Arial,Helvetica,sans-serif"><div style="display:none;max-height:0;overflow:hidden">` + html.EscapeString(t.Description) + `</div><table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr><td align="center" style="padding:32px 12px"><table role="presentation" width="560" cellpadding="0" cellspacing="0" style="width:100%;max-width:560px;background:#fff;border:1px solid #dde2d5;border-radius:16px;padding:20px 0">` + rows.String() + `</table></td></tr></table></body></html>`
	text := fmt.Sprintf("%s\n\n%s\n\nWorkspace: %s\nDomain: %s\n\nUp next: %s\n\n%s\n\nXem workspace service update.", t.Heading, t.Description, workspace, domain, t.Next, link)
	return body, text
}
func (t Template) Design() []byte {
	contents := []any{}
	for i, p := range t.fragments("{{workspace_name}}", "{{domain}}", "https://app.xem.email/settings/sending") {
		contents = append(contents, map[string]any{"id": fmt.Sprintf("onboarding-text-%d", i), "type": "text", "values": map[string]any{"_meta": map[string]string{"htmlID": fmt.Sprintf("u_content_text_%d", i+1)}, "text": p, "containerPadding": "12px 32px", "fontSize": "16px", "lineHeight": "175%", "color": "#50574b"}})
	}
	design := map[string]any{"schemaVersion": 18, "counters": map[string]int{"u_row": 1, "u_column": 1, "u_content_text": len(contents)}, "body": map[string]any{"id": "onboarding-body", "rows": []any{map[string]any{"id": "onboarding-row", "cells": []int{1}, "columns": []any{map[string]any{"id": "onboarding-column", "contents": contents, "values": map[string]any{"_meta": map[string]string{"htmlID": "u_column_1"}, "backgroundColor": "#ffffff"}}}, "values": map[string]any{"_meta": map[string]string{"htmlID": "u_row_1"}, "padding": "20px 0px", "backgroundColor": "#ffffff"}}}, "values": map[string]any{"backgroundColor": "#edf0e8", "contentWidth": "560px", "fontFamily": map[string]string{"label": "Arial", "value": "arial,helvetica,sans-serif"}}}}
	b, _ := json.MarshalIndent(design, "", "  ")
	return b
}
