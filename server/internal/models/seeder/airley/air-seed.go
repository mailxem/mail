package airley

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"kori/internal/models"
	"kori/internal/utils/base64"
	console "kori/internal/utils/logger"

	"kori/internal/handlers"
)

var log = console.New("SEEDER")

func LoadAirleyTemplates(db *gorm.DB) error {
	team, err := models.GetTeamByName("Team Airley", db)
	if err != nil {
		team = &models.Team{
			Name: "Team Airley",
		}
		if err := db.Create(team).Error; err != nil {
			return log.Error("Failed to create team", err)
		}

		log.Success("Created airley team %s", team.ID)

		password := "change_this_password"

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return log.Error("Failed to hash password", err)
		}

		user := &models.User{
			Email:     "admin@airley.io",
			Password:  string(hashedPassword),
			TeamID:    team.ID,
			FirstName: "SuperHuman",
			LastName:  "At Airley",
			Role:      models.UserRoleAdmin,
		}

		if err := db.Create(user).Error; err != nil {
			return log.Error("Failed to create user", err)
		}

		if err := models.AssignDefaultPermissions(db, user); err != nil {
			return log.Error("Failed to assign default permissions", err)
		}

		log.Success("Created airley admin user %s", user.ID)
	}

	templates, err := models.LoadJSON[models.InitialTemplates]("internal/models/seeder/airley/templates.json")
	if err != nil {
		return err
	}

	// get api key
	apiKey := &models.APIKey{}
	if err := db.Where("team_id = ? AND name = ?", team.ID, "Airley API Key").First(apiKey).Error; err != nil {
		// create api key
		apiKey := &models.APIKey{
			Name:   "Airley API Key",
			TeamID: team.ID,
		}
		if err := db.Create(apiKey).Error; err != nil {
			return log.Error("Failed to create api key", err)
		}
	}

	log.Info("Created api key %s", apiKey.ID)

	// Create mailing lists
	for _, list := range templates.MailingLists {
		// check if list already exists
		var existingList models.MailingList
		if err := db.Where("name = ? AND team_id = ?", list.Name, team.ID).First(&existingList).Error; err == nil {
			return log.Error("Mailing list already exists", err)
		} else if err != gorm.ErrRecordNotFound {
			return log.Error("Failed to check if mailing list exists", err)
		}

		list.TeamID = team.ID
		if err := db.Create(&list).Error; err != nil {
			return log.Error("Failed to create mailing list", err)
		}
	}

	// upload

	// Create templates
	for _, tmpl := range templates.Templates {
		tmpl.TeamID = team.ID
		category := models.EmailCategory{}
		if err := db.Where("name = ? AND team_id = ?", tmpl.CategoryName, team.ID).First(&category).Error; err != nil {
			return log.Error("Failed to find category", err)
		}

		file, err := CreateHTMLFile(db, team.ID, tmpl.HtmlFileID)
		if err != nil {
			return log.Error("Failed to create html file", err)
		}

		log.Info("Created html file %s", file.ID)

		template := models.Template{
			Name:       tmpl.Name,
			Subject:    tmpl.Subject,
			HtmlFileID: file.ID,
			DesignJSON: base64.EncodeToBase64(string(tmpl.DesignJSON)),
			TeamID:     team.ID,
			CategoryID: category.ID,
			Variables:  tmpl.Variables,
		}
		if err := db.Create(&template).Error; err != nil {
			return log.Error("Failed to create template", err)
		}
	}

	log.Success("Successfully loaded all templates")

	return nil
}

// write a function to create a html file based on string first we need to saniztinze the string remove = and -
func sanitizeString(str string) string {
	re := regexp.MustCompile(`\\`)      // Match all backslashes
	return re.ReplaceAllString(str, "") // Remove them
}

// CreateHTMLFile creates a new HTML file from content string
func CreateHTMLFile(db *gorm.DB, teamId string, content string) (*models.File, error) {
	log.Info("Creating HTML file for team %s", teamId)
	// Sanitize content
	sanitizedContent := sanitizeString(content)

	// Create file in memory
	fileBytes := []byte(sanitizedContent)

	fileName := fmt.Sprintf("tmp/%s.html", uuid.New().String())

	// Get storage handler
	storage := handlers.GetStorageHandler()
	if storage == nil {
		return nil, log.Error("storage handler not configured", nil)
	}

	// Upload file
	url, err := storage.UploadFile(context.Background(), fileBytes, fileName, types.ObjectCannedACLAuthenticatedRead, "text/html")
	if err != nil {
		return nil, log.Error("failed to upload file", err)
	}

	log.Info("Uploaded file to %s", url)

	// Create file record
	file := &models.File{
		TeamID: teamId,
		Path:   url[strings.LastIndex(url, "/")+1:],
		Name:   fileName,
		Size:   int64(len(fileBytes)),
		Type:   "text/html",
	}

	if err := db.Create(file).Error; err != nil {
		return nil, log.Error("failed to create file record", err)
	}

	log.Info("Created file record %s", file.ID)

	return file, nil
}
