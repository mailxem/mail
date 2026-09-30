// Package onboardingemails owns the transactional copy used by the sender and
// the editable starter library. It has no delivery or tenant-data dependencies.
package onboardingemails

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
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

// Mirrors website/tailwind.config.ts, website/lib/brand.ts and the website's
// Brand component. The generator copies the original mark and licensed fonts.
const (
	cream       = "#ffffef"
	iris        = "#5b3cc4"
	ink         = "#22251f"
	lemon       = "#edf09b"
	forest      = "#164b3f"
	lavender    = "#e7d8fa"
	muted       = "#66695e"
	footer      = "#222d25"
	bodyFont    = "'DM Sans',Arial,Helvetica,sans-serif"
	headingFont = "'EB Garamond',Georgia,'Times New Roman',serif"
)

func brandAsset(link, name string) string {
	u, err := url.Parse(link)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" {
		return ""
	}
	u.Path = "/assets/template-starters/brand/" + name
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	u.User = nil
	return u.String()
}
func (t Template) fragments(workspace, domain, link string) []string {
	return []string{
		`<table role="presentation" cellpadding="0" cellspacing="0"><tr><td width="46"><img src="` + html.EscapeString(brandAsset(link, "xem-mark.png")) + `" width="36" height="36" alt="" style="display:block;border:0;border-radius:6px"></td><td style="font-family:` + bodyFont + `;font-size:28px;font-weight:600;letter-spacing:-1.5px;color:` + ink + `">Xem</td></tr></table>`,
		`<p style="margin:0;font-family:` + bodyFont + `;font-size:10px;font-weight:600;letter-spacing:1.7px;color:` + forest + `"><span style="display:inline-block;background:` + lemon + `;padding:8px 11px;border-radius:6px">YOUR MANAGED SENDING JOURNEY</span></p>`,
		`<h1 style="margin:0;font-family:` + headingFont + `;font-size:44px;font-weight:400;line-height:1.03;letter-spacing:-1.5px;color:` + ink + `">` + html.EscapeString(t.Heading) + `</h1>`,
		`<p style="margin:0;font-family:` + bodyFont + `;font-size:15px;line-height:1.8;color:` + muted + `">` + html.EscapeString(t.Description) + `</p>`,
		`<p style="margin:0;padding:18px 20px;background:` + lavender + `;border-radius:12px;font-family:` + bodyFont + `;font-size:13px;line-height:1.8;overflow-wrap:anywhere;word-break:break-word;color:` + ink + `">Workspace: <strong>` + html.EscapeString(workspace) + `</strong><br>Domain: <strong>` + html.EscapeString(domain) + `</strong></p>`,
		`<p style="margin:0;font-family:` + bodyFont + `;font-size:15px;line-height:1.8;color:` + muted + `"><strong style="color:` + forest + `">Up next</strong><br>` + html.EscapeString(t.Next) + `</p>`,
		`<p style="margin:6px 0 12px"><a href="` + html.EscapeString(link) + `" style="display:inline-block;padding:14px 22px;background:` + iris + `;border:1px solid #49309e;box-shadow:0 2px 0 #362278;color:#ffffff;font-family:` + bodyFont + `;font-size:14px;font-weight:600;text-decoration:none;border-radius:8px">Open managed sending &rarr;</a></p>`,
		`<div style="padding:24px;background:` + footer + `;border-radius:16px;color:` + cream + `"><p style="margin:0 0 12px;font-family:` + headingFont + `;font-size:26px;line-height:1.1;letter-spacing:-0.5px">Every email,<br>a little more human.</p><p style="margin:0;font-family:` + bodyFont + `;font-size:11px;line-height:1.8;color:` + cream + `">You received this service update as the workspace owner.<br>Built with Xem · Made for the people on the other end.</p></div>`,
	}
}
func (t Template) Render(workspace, domain, link string) (string, string) {
	var rows strings.Builder
	for _, p := range t.fragments(workspace, domain, link) {
		rows.WriteString(`<tr><td class="email-section" style="padding:12px 32px">` + p + `</td></tr>`)
	}
	// Web fonts enhance supported email clients. All essential typography and colors
	// are inline, with safe fallbacks when remote fonts or images are blocked.
	fonts := `@font-face{font-family:'DM Sans';font-style:normal;font-weight:100 1000;src:url('` + html.EscapeString(brandAsset(link, "dm-sans.woff2")) + `') format('woff2')}@font-face{font-family:'EB Garamond';font-style:normal;font-weight:400 800;src:url('` + html.EscapeString(brandAsset(link, "eb-garamond.woff2")) + `') format('woff2')}`
	body := `<!doctype html><html lang="en"><head><meta name="viewport" content="width=device-width,initial-scale=1"><meta charset="utf-8"><meta name="color-scheme" content="light"><title>` + html.EscapeString(t.Subject) + `</title><style>` + fonts + `@media(max-width:480px){.email-section{padding-left:22px!important;padding-right:22px!important}}</style></head><body style="margin:0;background:` + cream + `;font-family:` + bodyFont + `;color:` + ink + `"><div style="display:none;max-height:0;overflow:hidden">` + html.EscapeString(t.Description) + `</div><table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr><td align="center" style="padding:24px 12px"><table role="presentation" width="560" cellpadding="0" cellspacing="0" style="width:100%;max-width:560px;table-layout:fixed;background:` + cream + `;border:1px solid #deded0;border-radius:24px;padding:20px 0">` + rows.String() + `</table></td></tr></table></body></html>`
	text := fmt.Sprintf("Xem\n\n%s\n\n%s\n\nWorkspace: %s\nDomain: %s\n\nUp next: %s\n\n%s\n\nEvery email, a little more human.\nXem workspace service update.", t.Heading, t.Description, workspace, domain, t.Next, link)
	return body, text
}
func (t Template) Design() []byte {
	contents := []any{}
	for i, p := range t.fragments("{{workspace_name}}", "{{domain}}", "https://app.xem.email/settings/sending") {
		// Registry-owned inline image URLs are resolved on import for self-hosted apps.
		p = strings.ReplaceAll(p, "https://app.xem.email/assets/template-starters/", "/assets/template-starters/")
		font := map[string]string{"label": "DM Sans", "value": bodyFont, "url": "/assets/template-starters/brand/fonts.css"}
		if i == 2 || i == 7 {
			font["label"], font["value"] = "EB Garamond", headingFont
		}
		contents = append(contents, map[string]any{"id": fmt.Sprintf("onboarding-text-%d", i), "type": "text", "values": map[string]any{"_meta": map[string]string{"htmlID": fmt.Sprintf("u_content_text_%d", i+1)}, "text": p, "containerPadding": "12px 32px", "fontFamily": font, "fontSize": "15px", "lineHeight": "180%", "color": muted}})
	}
	design := map[string]any{"schemaVersion": 18, "counters": map[string]int{"u_row": 1, "u_column": 1, "u_content_text": len(contents)}, "body": map[string]any{"id": "onboarding-body", "rows": []any{map[string]any{"id": "onboarding-row", "cells": []int{1}, "columns": []any{map[string]any{"id": "onboarding-column", "contents": contents, "values": map[string]any{"_meta": map[string]string{"htmlID": "u_column_1"}, "backgroundColor": cream}}}, "values": map[string]any{"_meta": map[string]string{"htmlID": "u_row_1"}, "padding": "20px 0px", "backgroundColor": cream}}}, "values": map[string]any{"backgroundColor": cream, "contentWidth": "560px", "fontFamily": map[string]string{"label": "DM Sans", "value": bodyFont}}}}
	b, _ := json.MarshalIndent(design, "", "  ")
	return b
}
