package marketing

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Starter struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	HTMLBody    string `json:"htmlBody"`
	Color       string `json:"color"`
	DesignJSON  string `json:"-"`
}

func Starters() []Starter {
	specs := [][5]string{{"editorial", "The weekly edit", "A considered collection of stories, ideas and links.", "Your weekly dose of inspiration", "#6250e8"}, {"product", "Product notes", "Keep your community close to what you are building.", "A little update. A big difference.", "#157575"}, {"digest", "Community digest", "Member stories, upcoming events and everything in between.", "Good things are happening here", "#b86b3e"}, {"launch", "Something new", "Make your next announcement one worth opening.", "Meet your new favorite thing", "#34363e"}}
	out := []Starter{}
	for _, x := range specs {
		out = append(out, Starter{x[0], x[1], x[2], x[3], fmt.Sprintf(`<!doctype html><html><body style="margin:0;background:#f5f5f8;font-family:Arial,sans-serif;color:#25252e"><table role="presentation" width="100%%" cellpadding="0" cellspacing="0"><tr><td align="center" style="padding:32px 16px"><table role="presentation" width="600" style="max-width:100%%;background:#fff;border-radius:16px" cellpadding="32"><tr><td><p style="font-size:12px;letter-spacing:3px;color:%s">YOUR NEWSLETTER</p><h1 style="font-size:36px;line-height:1.2">%s</h1><p style="line-height:1.8;color:#6b6b76">Hi {{first_name}},</p><p style="line-height:1.8;color:#6b6b76">%s</p><hr style="border:0;border-top:1px solid #eee;margin:28px 0"><h2>A story worth sharing</h2><p style="line-height:1.8;color:#6b6b76">Replace this with your latest news. Keep it personal, useful, and true to your voice.</p><a href="https://example.com" style="display:inline-block;background:%s;color:white;padding:14px 22px;border-radius:8px;text-decoration:none">Read the story →</a><p style="margin-top:36px;color:#8b8b94;font-size:13px">Thanks for being here. See you in the next edition.</p></td></tr></table></td></tr></table></body></html>`, x[4], x[3], x[2], x[4]), x[4], starterDesign(x[3], x[2], x[4])})
	}
	return out
}

// Native Unlayer design data keeps starters editable in the existing template editor.
func starterDesign(subject, description, color string) string {
	texts := []string{`<p style="letter-spacing:3px">YOUR NEWSLETTER</p>`, "<h1>" + subject + "</h1>", "<p>Hi {{first_name}},</p><p>" + description + "</p>", "<h2>A story worth sharing</h2><p>Replace this with your latest news. Keep it personal, useful, and true to your voice.</p>", `<p><a href="https://example.com">Read the story →</a></p>`, "<p>Thanks for being here. See you in the next edition.</p>"}
	contents := []interface{}{}
	for i, text := range texts {
		contents = append(contents, map[string]interface{}{"id": fmt.Sprintf("starter-text-%d", i), "type": "text", "values": map[string]interface{}{"text": text, "containerPadding": "12px 32px", "fontFamily": map[string]string{"label": "Arial", "value": "arial,helvetica,sans-serif"}, "fontSize": "16px", "lineHeight": "160%", "color": color}})
	}
	design := map[string]interface{}{"schemaVersion": 16, "counters": map[string]int{"u_row": 1, "u_column": 1, "u_content_text": len(texts)}, "body": map[string]interface{}{"id": "starter-body", "rows": []interface{}{map[string]interface{}{"id": "starter-row", "cells": []int{1}, "columns": []interface{}{map[string]interface{}{"id": "starter-column", "contents": contents, "values": map[string]interface{}{"backgroundColor": "#ffffff"}}}, "values": map[string]interface{}{"backgroundColor": "#ffffff", "padding": "20px 0px"}}}, "values": map[string]interface{}{"backgroundColor": "#f5f5f8", "contentWidth": "600px", "fontFamily": map[string]string{"label": "Arial", "value": "arial,helvetica,sans-serif"}}}}
	data, _ := json.Marshal(design)
	return base64.StdEncoding.EncodeToString(data)
}
