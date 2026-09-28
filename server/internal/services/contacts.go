package services

import (
	"context"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
)

func init() {
	events.On("contact_import.created", func(data interface{}) {
		contact_import := data.(*models.ContactImport)
		log.Info("Contact import created %v", contact_import)
		if contact_import.Status == models.ContactImportStatusPending {
			task := tasks.ContactImportTask{
				ImportID: contact_import.ID,
			}
			if err := taskClient.EnqueueContactImportTask(context.Background(), task); err != nil {
				log.Error("Failed to enqueue contact import task: %v", err)
			}
		}
	})
}
