// Package workflowconfig shares configuration contracts between publishing and execution.
package workflowconfig

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Rule struct {
	Variable string `json:"variable"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
type Condition struct {
	Conditions []Rule            `json:"conditions"`
	Operator   string            `json:"operator"`
	Branches   map[string]string `json:"branches"`
}
type Wait struct {
	Duration string `json:"duration"`
	Until    string `json:"until"`
}

var ContactFields = map[string]bool{"first_name": true, "last_name": true, "company": true, "country": true, "city": true, "state": true, "zip": true, "address": true, "phone": true, "linkedin": true, "twitter": true, "facebook": true, "instagram": true}
var variableName = regexp.MustCompile(`^workflow_[a-zA-Z][a-zA-Z0-9_]{0,63}$`)

func Validate(kind string, raw []byte) error {
	switch kind {
	case "CONDITION":
		var c Condition
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("invalid condition configuration")
		}
		if len(c.Conditions) < 1 || len(c.Conditions) > 20 {
			return fmt.Errorf("use between 1 and 20 rules")
		}
		if c.Operator != "" && c.Operator != "AND" && c.Operator != "OR" {
			return fmt.Errorf("choose all or any rules")
		}
		if len(c.Branches) > 0 {
			return fmt.Errorf("use graph edges for condition paths")
		}
		for _, r := range c.Conditions {
			if strings.TrimSpace(r.Variable) == "" || len(r.Variable) > 128 || len(r.Value) > 1000 {
				return fmt.Errorf("invalid rule variable or value")
			}
			switch r.Operator {
			case "==", "!=", ">", "<", ">=", "<=", "contains", "not_contains", "starts_with", "ends_with", "exists", "not_exists", "empty", "not_empty":
			default:
				return fmt.Errorf("invalid condition operator")
			}
		}
	case "WAIT":
		var w Wait
		if json.Unmarshal(raw, &w) != nil {
			return fmt.Errorf("invalid wait configuration")
		}
		if w.Until != "" {
			if w.Duration != "" {
				return fmt.Errorf("choose a duration or a date, not both")
			}
			if _, err := time.Parse(time.RFC3339, w.Until); err != nil {
				return fmt.Errorf("wait date must include a timezone")
			}
		} else {
			d, err := time.ParseDuration(w.Duration)
			if err != nil || d < time.Second || d > 365*24*time.Hour {
				return fmt.Errorf("delay must be between one second and 365 days")
			}
		}
	case "TAG":
		var c struct {
			Action string   `json:"action"`
			Tags   []string `json:"tags"`
		}
		if json.Unmarshal(raw, &c) != nil || (c.Action != "add" && c.Action != "remove") || len(c.Tags) < 1 || len(c.Tags) > 20 {
			return fmt.Errorf("choose add or remove and 1–20 tags")
		}
		for _, tag := range c.Tags {
			if strings.TrimSpace(tag) == "" || len(tag) > 80 {
				return fmt.Errorf("tag names must be 1–80 characters")
			}
		}
	case "UPDATE_SUBSCRIBER":
		var c struct {
			Fields map[string]interface{} `json:"fields"`
		}
		if json.Unmarshal(raw, &c) != nil || len(c.Fields) < 1 {
			return fmt.Errorf("choose contact fields and text values")
		}
		for k, v := range c.Fields {
			text, ok := v.(string)
			if !ContactFields[k] || !ok || len(text) > 1000 {
				return fmt.Errorf("invalid contact field or value: %s", k)
			}
		}
	case "PERCENTAGE_SPLIT":
		var c struct {
			Percentage *int `json:"percentage"`
		}
		if json.Unmarshal(raw, &c) != nil || c.Percentage == nil || *c.Percentage < 1 || *c.Percentage > 99 {
			return fmt.Errorf("path A must receive between 1 and 99 percent")
		}
	case "SET_VARIABLE":
		var c struct {
			Variable string `json:"variable"`
			Value    string `json:"value"`
		}
		if json.Unmarshal(raw, &c) != nil || !variableName.MatchString(c.Variable) || len(c.Value) > 1000 {
			return fmt.Errorf("use a workflow_ variable name and a value up to 1000 characters")
		}
	}
	return nil
}

func WaitDuration(w Wait, now time.Time) (time.Duration, error) {
	if w.Until != "" {
		target, err := time.Parse(time.RFC3339, w.Until)
		if err != nil {
			return 0, err
		}
		d := target.Sub(now)
		if d < 0 {
			d = 0
		}
		return d, nil
	}
	return time.ParseDuration(w.Duration)
}
