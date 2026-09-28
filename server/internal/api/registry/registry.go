package registry

import (
	"github.com/labstack/echo/v4"

	"kori/internal/api/controllers"
	"kori/internal/api/middleware"
	"kori/internal/handlers"
	"kori/internal/models"
	"kori/internal/services"

	"gorm.io/gorm"
)

// 📝 RegisterCRUDRoutes registers CRUD routes for all models - godoc
// @Summary Register CRUD routes for all models
// @Description Register CRUD routes for all models
// @Accept json
// @Produce json
func RegisterCRUDRoutes(g *echo.Group, db *gorm.DB) {
	// Teams
	teamService := services.NewBaseService(db, models.Team{})
	teamController := controllers.NewBaseController(teamService)
	teamGroup := g.Group("/teams")
	teamGroup.Use(middleware.RequirePermissions(db, "teams:read"))

	// @Summary List teams
	// @Description Get a list of all teams
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Team
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/teams [get]
	teamGroup.GET("", teamController.List)
	// @Summary Get team
	// @Description Get a team by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Team ID"
	// @Success 200 {object} models.Team
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/teams/{id} [get]
	teamGroup.GET("/:id", teamController.Get)

	// Protected team routes
	teamWriteGroup := teamGroup.Group("")
	teamWriteGroup.Use(middleware.RequirePermissions(db, "teams:write"))
	// @Summary Create team
	// @Description Create a new team
	// @Accept json
	// @Produce json
	// @Param team body models.Team true "Team object"
	// @Success 201 {object} models.Team
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/teams [post]
	teamWriteGroup.POST("", teamController.Create)
	// @Summary Update team
	// @Description Update an existing team
	// @Accept json
	// @Produce json
	// @Param id path string true "Team ID"
	// @Param team body models.Team true "Team object"
	// @Success 200 {object} models.Team
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/teams/{id} [put]
	teamWriteGroup.PUT("/:id", teamController.Update)
	// @Summary Delete team
	// @Description Delete a team
	// @Accept json
	// @Produce json
	// @Param id path string true "Team ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/teams/{id} [delete]
	teamWriteGroup.DELETE("/:id", teamController.Delete)

	// Team Invitations with team-specific permissions
	invitationService := services.NewBaseService(db, models.TeamInvite{})
	invitationController := controllers.NewBaseController(invitationService)
	invitationGroup := g.Group("/team-invitations")
	invitationGroup.Use(middleware.RequirePermissions(db, "team_invites:read"))
	// @Summary List team invitations
	// @Description Get a list of all team invitations
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.TeamInvite
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/team-invitations [get]
	invitationGroup.GET("", invitationController.List)

	// Protected invitation routes
	invitationWriteGroup := invitationGroup.Group("")
	invitationWriteGroup.Use(middleware.RequirePermissions(db, "team_invites:write"))
	// @Summary Delete team invitation
	// @Description Delete a team invitation
	// @Accept json
	// @Produce json
	// @Param id path string true "Invitation ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/team-invitations/{id} [delete]
	invitationWriteGroup.DELETE("/:id", invitationController.Delete)

	// file routes
	fileService := services.NewBaseService(db, models.File{})
	fileController := controllers.NewBaseController(fileService)
	fileGroup := g.Group("/files")
	fileGroup.Use(middleware.RequirePermissions(db, "files:read"))
	// @Summary List files
	// @Description Get a list of all files
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.File
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/files [get]
	fileGroup.GET("", fileController.List)
	// @Summary Get file
	// @Description Get a file by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "File ID"
	// @Success 200 {object} models.File
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/files/{id} [get]
	fileGroup.GET("/:id", fileController.Get)

	// Contacts with team-specific permissions
	contactService := services.NewBaseService(db, models.Contact{})
	contactController := controllers.NewBaseController(contactService)
	contactGroup := g.Group("/contacts")
	contactGroup.Use(middleware.RequirePermissions(db, "contacts:read"))
	// @Summary List contacts
	// @Description Get a list of all contacts
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Contact
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/contacts [get]
	contactGroup.GET("", contactController.List)
	// @Summary Get contact
	// @Description Get a contact by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Contact ID"
	// @Success 200 {object} models.Contact
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/contacts/{id} [get]
	contactGroup.GET("/:id", contactController.Get)

	// Protected contact routes
	contactWriteGroup := contactGroup.Group("")
	contactWriteGroup.Use(middleware.RequirePermissions(db, "contacts:write"))
	// @Summary Create contact
	// @Description Create a new contact
	// @Accept json
	// @Produce json
	// @Param contact body models.Contact true "Contact object"
	// @Success 201 {object} models.Contact
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/contacts [post]
	contactWriteGroup.POST("", contactController.Create)
	// @Summary Update contact
	// @Description Update an existing contact
	// @Accept json
	// @Produce json
	// @Param id path string true "Contact ID"
	// @Param contact body models.Contact true "Contact object"
	// @Success 200 {object} models.Contact
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/contacts/{id} [put]
	contactWriteGroup.PUT("/:id", contactController.Update)
	// @Summary Delete contact
	// @Description Delete a contact
	// @Accept json
	// @Produce json
	// @Param id path string true "Contact ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/contacts/{id} [delete]
	contactWriteGroup.DELETE("/:id", contactController.Delete)

	// Email Categories with team-specific permissions
	categoryService := services.NewBaseService(db, models.EmailCategory{})
	categoryController := controllers.NewBaseController(categoryService)
	categoryGroup := g.Group("/categories")
	categoryGroup.Use(middleware.RequirePermissions(db, "categories:read"))
	// @Summary List categories
	// @Description Get a list of all email categories
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.EmailCategory
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/categories [get]
	categoryGroup.GET("", categoryController.List)
	// @Summary Get category
	// @Description Get an email category by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Category ID"
	// @Success 200 {object} models.EmailCategory
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/categories/{id} [get]
	categoryGroup.GET("/:id", categoryController.Get)

	// Protected category routes
	categoryWriteGroup := categoryGroup.Group("")
	categoryWriteGroup.Use(middleware.RequirePermissions(db, "categories:write"))
	// @Summary Create category
	// @Description Create a new email category
	// @Accept json
	// @Produce json
	// @Param category body models.EmailCategory true "Category object"
	// @Success 201 {object} models.EmailCategory
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/categories [post]
	categoryWriteGroup.POST("", categoryController.Create)
	// @Summary Update category
	// @Description Update an existing email category
	// @Accept json
	// @Produce json
	// @Param id path string true "Category ID"
	// @Param category body models.EmailCategory true "Category object"
	// @Success 200 {object} models.EmailCategory
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/categories/{id} [put]
	categoryWriteGroup.PUT("/:id", categoryController.Update)
	// @Summary Delete category
	// @Description Delete an email category
	// @Accept json
	// @Produce json
	// @Param id path string true "Category ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/categories/{id} [delete]
	categoryWriteGroup.DELETE("/:id", categoryController.Delete)

	// Mailing Lists with team-specific permissions
	mailingListService := services.NewBaseService(db, models.MailingList{})
	mailingListController := controllers.NewBaseController(mailingListService)
	listGroup := g.Group("/mailing-lists")
	listGroup.Use(middleware.RequirePermissions(db, "lists:read"))
	// @Summary List mailing lists
	// @Description Get a list of all mailing lists
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.MailingList
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/mailing-lists [get]
	listGroup.GET("", mailingListController.List)
	// @Summary Get mailing list
	// @Description Get a mailing list by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "List ID"
	// @Success 200 {object} models.MailingList
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/mailing-lists/{id} [get]
	listGroup.GET("/:id", mailingListController.Get)

	// Protected mailing list routes
	listWriteGroup := listGroup.Group("")
	listWriteGroup.Use(middleware.RequirePermissions(db, "lists:write"))
	// @Summary Create mailing list
	// @Description Create a new mailing list
	// @Accept json
	// @Produce json
	// @Param list body models.MailingList true "List object"
	// @Success 201 {object} models.MailingList
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/mailing-lists [post]
	listWriteGroup.POST("", mailingListController.Create)
	// @Summary Update mailing list
	// @Description Update an existing mailing list
	// @Accept json
	// @Produce json
	// @Param id path string true "List ID"
	// @Param list body models.MailingList true "List object"
	// @Success 200 {object} models.MailingList
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/mailing-lists/{id} [put]
	listWriteGroup.PUT("/:id", mailingListController.Update)
	// @Summary Delete mailing list
	// @Description Delete a mailing list
	// @Accept json
	// @Produce json
	// @Param id path string true "List ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/mailing-lists/{id} [delete]
	listWriteGroup.DELETE("/:id", mailingListController.Delete)

	// SMTP Configs with team-specific permissions
	smtpConfigService := services.NewBaseService(db, models.SMTPConfig{})
	smtpConfigController := controllers.NewBaseController(smtpConfigService)
	smtpGroup := g.Group("/smtp-configs")
	smtpGroup.Use(middleware.RequirePermissions(db, "smtp_configs:read"))
	// @Summary List SMTP configs
	// @Description Get a list of all SMTP configurations
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.SMTPConfig
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/smtp-configs [get]
	smtpGroup.GET("", smtpConfigController.List)
	// @Summary Get SMTP config
	// @Description Get an SMTP configuration by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Success 200 {object} models.SMTPConfig
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/smtp-configs/{id} [get]
	smtpGroup.GET("/:id", smtpConfigController.Get)

	// Protected SMTP config routes
	smtpWriteGroup := smtpGroup.Group("")
	smtpWriteGroup.Use(middleware.RequirePermissions(db, "smtp_configs:write"))
	// @Summary Create SMTP config
	// @Description Create a new SMTP configuration
	// @Accept json
	// @Produce json
	// @Param config body models.SMTPConfig true "Config object"
	// @Success 201 {object} models.SMTPConfig
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/smtp-configs [post]
	smtpWriteGroup.POST("", smtpConfigController.Create)
	// @Summary Update SMTP config
	// @Description Update an existing SMTP configuration
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Param config body models.SMTPConfig true "Config object"
	// @Success 200 {object} models.SMTPConfig
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/smtp-configs/{id} [put]
	smtpWriteGroup.PUT("/:id", smtpConfigController.Update)
	// @Summary Delete SMTP config
	// @Description Delete an SMTP configuration
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/smtp-configs/{id} [delete]
	smtpWriteGroup.DELETE("/:id", smtpConfigController.Delete)

	// IMAP with team-specific permissions
	imapService := services.NewBaseService(db, models.IMAPConfig{})
	imapController := controllers.NewBaseController(imapService)
	imapGroup := g.Group("/imap")
	imapGroup.Use(middleware.RequirePermissions(db, "imap_configs:read"))
	// @Summary List IMAP configs
	// @Description Get a list of all IMAP configurations
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.IMAPConfig
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imap [get]
	imapGroup.GET("", imapController.List)
	// @Summary Get IMAP config
	// @Description Get an IMAP configuration by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Success 200 {object} models.IMAPConfig
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imap/{id} [get]
	imapGroup.GET("/:id", imapController.Get)

	// Protected IMAP config routes
	imapWriteGroup := imapGroup.Group("")
	imapWriteGroup.Use(middleware.RequirePermissions(db, "imap_configs:write"))
	// @Summary Create IMAP config
	// @Description Create a new IMAP configuration
	// @Accept json
	// @Produce json
	// @Param config body models.IMAPConfig true "Config object"
	// @Success 201 {object} models.IMAPConfig
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imap [post]
	imapWriteGroup.POST("", imapController.Create)
	// @Summary Update IMAP config
	// @Description Update an existing IMAP configuration
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Param config body models.IMAPConfig true "Config object"
	// @Success 200 {object} models.IMAPConfig
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imap/{id} [put]
	imapWriteGroup.PUT("/:id", imapController.Update)
	// @Summary Delete IMAP config
	// @Description Delete an IMAP configuration
	// @Accept json
	// @Produce json
	// @Param id path string true "Config ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/imap/{id} [delete]
	imapWriteGroup.DELETE("/:id", imapController.Delete)

	// Domains with team-specific permissions
	domainService := services.NewBaseService(db, models.Domain{})
	domainController := controllers.NewBaseController(domainService)
	domainGroup := g.Group("/domains")
	domainGroup.Use(middleware.RequirePermissions(db, "domains:read"))
	// @Summary List domains
	// @Description Get a list of all domains
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Domain
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/domains [get]
	domainGroup.GET("", domainController.List)
	// @Summary Get domain
	// @Description Get a domain by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Domain ID"
	// @Success 200 {object} models.Domain
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/domains/{id} [get]
	domainGroup.GET("/:id", domainController.Get)

	// Protected domain routes
	domainWriteGroup := domainGroup.Group("")
	domainWriteGroup.Use(middleware.RequirePermissions(db, "domains:write"))
	// @Summary Create domain
	// @Description Create a new domain
	// @Accept json
	// @Produce json
	// @Param domain body models.Domain true "Domain object"
	// @Success 201 {object} models.Domain
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/domains [post]
	domainWriteGroup.POST("", domainController.Create)
	// @Summary Update domain
	// @Description Update an existing domain
	// @Accept json
	// @Produce json
	// @Param id path string true "Domain ID"
	// @Param domain body models.Domain true "Domain object"
	// @Success 200 {object} models.Domain
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/domains/{id} [put]
	domainWriteGroup.PUT("/:id", domainController.Update)
	// @Summary Delete domain
	// @Description Delete a domain
	// @Accept json
	// @Produce json
	// @Param id path string true "Domain ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/domains/{id} [delete]
	domainWriteGroup.DELETE("/:id", domainController.Delete)

	// Webhooks with team-specific permissions
	webhookService := services.NewBaseService(db, models.Webhook{})
	webhookController := controllers.NewBaseController(webhookService)
	webhookGroup := g.Group("/webhooks")
	webhookGroup.Use(middleware.RequirePermissions(db, "webhooks:read"))
	// @Summary List webhooks
	// @Description Get a list of all webhooks
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Webhook
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/webhooks [get]
	webhookGroup.GET("", webhookController.List)
	// @Summary Get webhook
	// @Description Get a webhook by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Webhook ID"
	// @Success 200 {object} models.Webhook
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/webhooks/{id} [get]
	webhookGroup.GET("/:id", webhookController.Get)

	// Protected webhook routes
	webhookWriteGroup := webhookGroup.Group("")
	webhookWriteGroup.Use(middleware.RequirePermissions(db, "webhooks:write"))
	// @Summary Create webhook
	// @Description Create a new webhook
	// @Accept json
	// @Produce json
	// @Param webhook body models.Webhook true "Webhook object"
	// @Success 201 {object} models.Webhook
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/webhooks [post]
	webhookWriteGroup.POST("", webhookController.Create)
	// @Summary Update webhook
	// @Description Update an existing webhook
	// @Accept json
	// @Produce json
	// @Param id path string true "Webhook ID"
	// @Param webhook body models.Webhook true "Webhook object"
	// @Success 200 {object} models.Webhook
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/webhooks/{id} [put]
	webhookWriteGroup.PUT("/:id", webhookController.Update)
	// @Summary Delete webhook
	// @Description Delete a webhook
	// @Accept json
	webhookWriteGroup.DELETE("/:id", webhookController.Delete)

	// Templates with team-specific permissions
	templateService := services.NewBaseService(db, models.Template{})
	templateController := controllers.NewBaseController(templateService)
	templateGroup := g.Group("/templates")
	templateGroup.Use(middleware.RequirePermissions(db, "templates:read"))
	// @Summary List templates
	// @Description Get a list of all templates
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Template
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/templates [get]
	templateGroup.GET("", templateController.List)
	// @Summary Get template
	// @Description Get a template by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Template ID"
	// @Success 200 {object} models.Template
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/templates/{id} [get]
	templateGroup.GET("/:id", templateController.Get)

	// Protected template routes
	templateWriteGroup := templateGroup.Group("")
	templateWriteGroup.Use(middleware.RequirePermissions(db, "templates:write"))
	// @Summary Create template
	// @Description Create a new template
	// @Accept json
	// @Produce json
	// @Param template body models.Template true "Template object"
	// @Success 201 {object} models.Template
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/templates [post]
	templateWriteGroup.POST("", templateController.Create)
	templateWriteGroup.PUT("/:id", templateController.Update)
	templateWriteGroup.DELETE("/:id", templateController.Delete)

	// API Keys with team-specific permissions
	apiKeyService := services.NewBaseService(db, models.APIKey{})
	apiKeyController := controllers.NewBaseController(apiKeyService)
	apiKeyGroup := g.Group("/api-keys")
	apiKeyGroup.Use(middleware.RequirePermissions(db, "api_keys:read"))
	// @Summary List API keys
	// @Description Get a list of all API keys
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.APIKey
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/api-keys [get]
	apiKeyGroup.GET("", apiKeyController.List)
	// @Summary Get API key
	// @Description Get an API key by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "API Key ID"
	// @Success 200 {object} models.APIKey
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/api-keys/{id} [get]
	apiKeyGroup.GET("/:id", apiKeyController.Get)

	// Protected API key routes
	apiKeyWriteGroup := apiKeyGroup.Group("")
	apiKeyWriteGroup.Use(middleware.RequirePermissions(db, "api_keys:write"))
	// @Summary Create API key
	// @Description Create a new API key
	// @Accept json
	// @Produce json
	// @Param apiKey body models.APIKey true "API Key object"
	// @Success 201 {object} models.APIKey
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/api-keys [post]
	apiKeyWriteGroup.POST("", apiKeyController.Create)
	apiKeyWriteGroup.PUT("/:id", apiKeyController.Update)
	apiKeyWriteGroup.DELETE("/:id", apiKeyController.Delete)

	// API KEY USAGE with team-specific permissions
	apiKeyUsageController := handlers.NewMarketingHandler(db, "", "")
	apiKeyUsageGroup := g.Group("/api-key-usage")
	apiKeyUsageGroup.Use(middleware.RequirePermissions(db, "api_key_usage:read"))
	// @Summary List API key usage
	// @Description Get a list of all API key usage
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.APIKeyUsage
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/api-key-usage [get]
	apiKeyUsageGroup.GET("", apiKeyUsageController.APIKeyUsage)
	// @Summary Get API key usage
	// @Description Get API key usage by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "API Key Usage ID"
	// @Success 200 {object} models.APIKeyUsage
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/api-key-usage/{id} [get]
	apiKeyUsageGroup.GET("/:id", apiKeyUsageController.APIKeyUsage)

	// Campaigns with team-specific permissions
	campaignService := services.NewBaseService(db, models.Campaign{})
	campaignController := controllers.NewBaseController(campaignService)
	campaignGroup := g.Group("/campaigns")
	campaignGroup.Use(middleware.RequirePermissions(db, "campaigns:read"))
	// @Summary List campaigns
	// @Description Get a list of all campaigns
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Campaign
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/campaigns [get]
	campaignGroup.GET("", campaignController.List)
	// @Summary Get campaign
	// @Description Get a campaign by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Campaign ID"
	// @Success 200 {object} models.Campaign
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/campaigns/{id} [get]
	campaignGroup.GET("/:id", campaignController.Get)

	// Protected campaign routes
	campaignWriteGroup := campaignGroup.Group("")
	campaignWriteGroup.Use(middleware.RequirePermissions(db, "campaigns:write"))
	// @Summary Create campaign
	// @Description Create a new campaign
	// @Accept json
	// @Produce json
	// @Param campaign body models.Campaign true "Campaign object"
	// @Success 201 {object} models.Campaign
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/campaigns [post]
	campaignWriteGroup.POST("", campaignController.Create)
	// @Summary Update campaign
	// @Description Update an existing campaign
	// @Accept json
	// @Produce json
	// @Param id path string true "Campaign ID"
	// @Param campaign body models.Campaign true "Campaign object"
	// @Success 200 {object} models.Campaign
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	campaignWriteGroup.PUT("/:id", campaignController.Update)
	// @Summary Delete campaign
	// @Description Delete a campaign
	// @Accept json
	// @Produce json
	// @Param id path string true "Campaign ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/campaigns/{id} [delete]
	campaignWriteGroup.DELETE("/:id", campaignController.Delete)

	// Automation routes with team-specific permissions
	automationService := services.NewBaseService(db, models.Automation{})
	automationController := controllers.NewBaseController(automationService)
	automationGroup := g.Group("/automations")
	automationGroup.Use(middleware.RequirePermissions(db, "automations:read"))
	// @Summary List automations
	// @Description Get a list of all automations
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Automation
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/automations [get]
	automationGroup.GET("", automationController.List)

	// Protected automation routes
	automationWriteGroup := automationGroup.Group("")
	automationWriteGroup.Use(middleware.RequirePermissions(db, "automations:write"))
	// @Summary Create automation
	// @Description Create a new automation
	// @Accept json
	// @Produce json
	// @Param automation body models.Automation true "Automation object"
	// @Success 201 {object} models.Automation
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/automations [post]
	automationWriteGroup.POST("", automationController.Create)
	// @Summary Update automation
	// @Description Update an existing automation
	// @Accept json
	// @Produce json
	// @Param id path string true "Automation ID"
	// @Param automation body models.Automation true "Automation object"
	// @Success 200 {object} models.Automation
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/automations/{id} [put]
	automationWriteGroup.PUT("/:id", automationController.Update)
	// @Summary Delete automation
	// @Description Delete an automation
	// @Accept json
	// @Produce json
	// @Param id path string true "Automation ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/automations/{id} [delete]
	automationWriteGroup.DELETE("/:id", automationController.Delete)

	// Model routes with team-specific permissions
	modelService := services.NewBaseService(db, models.Model{})
	modelController := controllers.NewBaseController(modelService)
	modelGroup := g.Group("/models")
	modelGroup.Use(middleware.RequirePermissions(db, "models:read"))
	// @Summary List models
	// @Description Get a list of all models
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Model
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/models [get]
	modelGroup.GET("", modelController.List)
	// @Summary Get model
	// @Description Get a model by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Model ID"
	// @Success 200 {object} models.Model
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/models/{id} [get]
	modelGroup.GET("/:id", modelController.Get)

	// Protected model routes
	modelWriteGroup := modelGroup.Group("")
	modelWriteGroup.Use(middleware.RequirePermissions(db, "models:write"))
	// @Summary Create model
	// @Description Create a new model
	// @Accept json
	// @Produce json
	// @Param model body models.Model true "Model object"
	// @Success 201 {object} models.Model
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/models [post]
	modelWriteGroup.POST("", modelController.Create)
	// @Summary Update model
	// @Description Update an existing model
	// @Accept json
	// @Produce json
	// @Param id path string true "Model ID"
	// @Param model body models.Model true "Model object"
	// @Success 200 {object} models.Model
	// @Failure 400 {object} map[string]string "Bad request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/models/{id} [put]
	modelWriteGroup.PUT("/:id", modelController.Update)
	// @Summary Delete model
	// @Description Delete a model
	// @Accept json
	// @Produce json
	// @Param id path string true "Model ID"
	// @Success 204 "No content"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/models/{id} [delete]
	modelWriteGroup.DELETE("/:id", modelController.Delete)

	// Emails with team-specific permissions
	emailService := services.NewBaseService(db, models.Email{})
	emailController := controllers.NewBaseController(emailService)
	emailGroup := g.Group("/emails")
	emailGroup.Use(middleware.RequirePermissions(db, "emails:read"))
	// @Summary List emails
	// @Description Get a list of all emails
	// @Accept json
	// @Produce json
	// @Success 200 {array} models.Email
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/emails [get]
	emailGroup.GET("", emailController.List)
	// @Summary Get email
	// @Description Get an email by ID
	// @Accept json
	// @Produce json
	// @Param id path string true "Email ID"
	// @Success 200 {object} models.Email
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Forbidden"
	// @Failure 404 {object} map[string]string "Not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/v1/emails/{id} [get]
	emailGroup.GET("/:id", emailController.Get)
}
