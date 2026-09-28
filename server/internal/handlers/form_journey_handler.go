package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/ai"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/services"
)

func FormDraftHandler(writer *ai.EmailWriter) echo.HandlerFunc {
	return func(c echo.Context) error {
		if team(c) == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Sign in to draft a form")
		}
		var input ai.FormDraftRequest
		if c.Bind(&input) != nil || len(strings.TrimSpace(input.Instruction)) < 3 || len(input.Instruction) > 4000 {
			return bad("Describe your form in 3–4000 characters")
		}
		draft, err := writer.DraftForm(c.Request().Context(), input)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, err.Error())
		}
		return c.JSON(http.StatusOK, draft)
	}
}

// CreateFormJourney saves editable templates and an inactive graph together.
// It never sends email or publishes the form. A stable request ID makes retries
// safe even when the response was lost after committing the transaction.
func (h *MarketingHandler) CreateFormJourney(c echo.Context) error {
	if team(c) == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Sign in to create a form journey")
	}
	if uuid.Validate(c.Param("id")) != nil {
		return missing()
	}
	var input struct {
		RequestID     string                 `json:"requestId"`
		Name          string                 `json:"name"`
		SMTPConfigID  string                 `json:"smtpConfigId"`
		PostalAddress string                 `json:"postalAddress"`
		Emails        []ai.JourneyEmailDraft `json:"emails"`
	}
	if c.Bind(&input) != nil || uuid.Validate(input.RequestID) != nil || uuid.Validate(input.SMTPConfigID) != nil || len(strings.TrimSpace(input.Name)) < 2 || len(input.Name) > 120 || len(strings.TrimSpace(input.PostalAddress)) < 8 || len(input.PostalAddress) > 500 {
		return bad("Provide a journey name, sender, postal address and request identifier")
	}
	if err := ai.ValidateJourneyEmails(input.Emails); err != nil {
		return bad(err.Error())
	}
	var workflow models.Automation
	err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var form models.Form
		if err := marketing.Scope(tx, team(c)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&form, "id = ?", c.Param("id")).Error; err != nil {
			return missing()
		}
		workflowID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("xem:form-journey:"+form.ID+":"+input.RequestID)).String()
		if err := marketing.Scope(tx, team(c)).Preload("Nodes", "is_deleted = ?", false).Preload("Edges").First(&workflow, "id = ?", workflowID).Error; err == nil {
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		// A replay returns the committed draft even if its sender has since
		// been removed. Only new journeys need current sender validation.
		var senderCount int64
		if err := marketing.Scope(tx, team(c)).Model(&models.SMTPConfig{}).Where("id = ?", input.SMTPConfigID).Count(&senderCount).Error; err != nil {
			return err
		}
		if senderCount != 1 {
			return bad("Choose a sender in this workspace")
		}
		var category models.EmailCategory
		if err := marketing.Scope(tx, team(c)).Where("name = ?", "Marketing").FirstOrCreate(&category, models.EmailCategory{Name: "Marketing", TeamID: team(c)}).Error; err != nil {
			return err
		}
		workflow = models.Automation{Base: models.Base{ID: workflowID}, TeamID: team(c), Name: input.Name, Description: "Follow-up for " + form.Name, TriggerEvent: "form.completed", IsActive: false}
		// Reserve the stable ID before SaveGraph so its ownership check succeeds.
		if err := tx.Omit(clause.Associations).Create(&workflow).Error; err != nil {
			return err
		}
		appendNode := func(kind models.NodeType, data map[string]interface{}) {
			data["position"] = map[string]int{"x": 0, "y": len(workflow.Nodes) * 165}
			raw, _ := json.Marshal(data)
			node := models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: kind, Data: datatypes.JSON(raw)}
			if len(workflow.Nodes) > 0 {
				workflow.Edges = append(workflow.Edges, models.AutomationNodeEdge{SourceID: workflow.Nodes[len(workflow.Nodes)-1].ID, TargetID: node.ID})
			}
			workflow.Nodes = append(workflow.Nodes, node)
		}
		appendNode(models.NodeTypeStart, map[string]interface{}{"formId": form.ID})
		for _, email := range input.Emails {
			body, design := journeyEmailContent(email.Body)
			template := models.Template{TeamID: team(c), CategoryID: category.ID, Name: email.Name, Subject: email.Subject, HTMLBody: body, DesignJSON: design}
			if err := tx.Omit(clause.Associations).Create(&template).Error; err != nil {
				return err
			}
			if email.DelayHours > 0 {
				appendNode(models.NodeTypeWait, map[string]interface{}{"duration": fmt.Sprintf("%dh", email.DelayHours)})
			}
			appendNode(models.NodeTypeEmail, map[string]interface{}{"templateId": template.ID, "smtpConfigId": input.SMTPConfigID, "postalAddress": input.PostalAddress})
		}
		appendNode(models.NodeTypeExit, map[string]interface{}{})
		return services.NewAutomationService(tx).SaveGraph(c.Request().Context(), &workflow)
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, workflow)
}

func journeyEmailContent(body string) (string, string) {
	paragraphs := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n")
	var content strings.Builder
	for _, paragraph := range paragraphs {
		content.WriteString("<p>" + strings.ReplaceAll(html.EscapeString(paragraph), "\n", "<br>") + "</p>")
	}
	text := content.String()
	design := map[string]interface{}{"schemaVersion": 16, "body": map[string]interface{}{"id": "journey-body", "rows": []interface{}{map[string]interface{}{"id": "journey-row", "cells": []int{1}, "columns": []interface{}{map[string]interface{}{"id": "journey-column", "contents": []interface{}{map[string]interface{}{"id": "journey-text", "type": "text", "values": map[string]interface{}{"text": text, "containerPadding": "24px", "fontSize": "16px", "lineHeight": "160%"}}}, "values": map[string]interface{}{}}}, "values": map[string]interface{}{}}}, "values": map[string]interface{}{"contentWidth": "600px", "backgroundColor": "#ffffff"}}}
	raw, _ := json.Marshal(design)
	return "<!doctype html><html><body><main style=\"max-width:600px;margin:auto;font:16px/1.6 Arial,sans-serif;padding:24px\">" + text + "</main></body></html>", base64.StdEncoding.EncodeToString(raw)
}
