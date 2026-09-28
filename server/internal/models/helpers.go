package models

import (
	"errors"

	"gorm.io/gorm"
)

// ways to get a smtp config
// 1. if they pass in a smtp config id, we get that one
// 2. if they don't pass in a smtp config id, we get the default one
// 3. if they pass in a smtp config id, but it's not found, we return an error
// 4. if they don't pass in a smtp config id, but the default one is not found, we return an error
// 5. they can also pass the provider name, and we get that provider's config

func GetSMTPConfig(teamID string, smtpConfigID string, providerName string, db *gorm.DB) (*SMTPConfig, error) {

	if smtpConfigID == "" && providerName == "" {

		// get the default smtp config for the team
		smtpConfig := &SMTPConfig{}
		if err := db.Where("team_id = ? AND is_default = true AND is_deleted = false", teamID).First(smtpConfig).Error; err != nil {
			return nil, err
		}
		return smtpConfig, nil
	}

	if smtpConfigID != "" {
		smtpConfig := &SMTPConfig{}
		if err := db.Where("id = ? AND team_id = ? AND is_deleted = false", smtpConfigID, teamID).First(smtpConfig).Error; err != nil {
			return nil, err
		}
		return smtpConfig, nil
	}

	if providerName != "" {
		smtpConfig := &SMTPConfig{}
		if err := db.Where("provider = ? AND team_id = ? AND is_deleted = false", providerName, teamID).First(smtpConfig).Error; err != nil {
			return nil, err
		}
		return smtpConfig, nil
	}

	return nil, errors.New("no smtp config found")
}

func GetIMAPConfig(teamID string, imapConfigID string, db *gorm.DB) (*IMAPConfig, error) {

	if imapConfigID == "" {
		imapConfig := &IMAPConfig{}
		if err := db.Where("team_id = ? AND is_deleted = false", teamID).First(imapConfig).Error; err != nil {
			return nil, err
		}
		return imapConfig, nil
	}

	if imapConfigID != "" {
		imapConfig := &IMAPConfig{}
		if err := db.Where("id = ? AND team_id = ? AND is_deleted = false", imapConfigID, teamID).First(imapConfig).Error; err != nil {
			return nil, err
		}
		return imapConfig, nil
	}

	return nil, errors.New("no imap config found")
}

// GetTeamByName retrieves a team from the database by its name
func GetTeamByName(name string, db *gorm.DB) (*Team, error) {
	team := &Team{}
	if err := db.Where("name = ? AND is_deleted = false", name).First(team).Error; err != nil {
		return nil, err
	}
	return team, nil
}

func GetCampaignByID(id string, db *gorm.DB) (*Campaign, error) {
	campaign := &Campaign{}
	if err := db.Where("id = ? AND is_deleted = false", id).Preload("Template.HtmlFile").First(campaign).Error; err != nil {
		return nil, err
	}
	return campaign, nil
}

func GetEmailListByID(id string, db *gorm.DB) (*MailingList, int, error) {
	emailList := &MailingList{}
	if err := db.Where("id = ? AND is_deleted = false", id).First(emailList).Error; err != nil {
		return nil, 0, err
	}
	var count int64
	if err := db.Model(&Contact{}).Where("list_id = ?", id).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	return emailList, int(count), nil
}

func GetEmailsByCampaignID(campaignID string, db *gorm.DB) ([]*Email, error) {
	emails := []*Email{}
	if err := db.Where("campaign_id = ? AND is_deleted = false", campaignID).Find(&emails).Error; err != nil {
		return nil, err
	}
	return emails, nil
}

func GetUnsubscribedContactsByListID(listID string, db *gorm.DB) ([]*Contact, error) {
	contacts := []*Contact{}
	if err := db.Where("list_id = ? AND status = ? AND is_deleted = false", listID, SubscriberStatusUnsubscribed).Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

func GetContactImportByID(id string, db *gorm.DB) (*ContactImport, error) {
	contactImport := &ContactImport{}
	if err := db.Where("id = ? AND is_deleted = false", id).First(contactImport).Error; err != nil {
		return nil, err
	}
	return contactImport, nil
}

func GetContactImportByFileID(fileID string, db *gorm.DB) (*ContactImport, error) {
	contactImport := &ContactImport{}
	if err := db.Where("file_id = ? AND is_deleted = false", fileID).First(contactImport).Error; err != nil {
		return nil, err
	}
	return contactImport, nil
}

func GetFileByID(id string, db *gorm.DB) (*File, error) {
	file := &File{}
	if err := db.Where("id = ? AND is_deleted = false", id).First(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}
