package handlers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"kori/internal/models"
	"kori/internal/utils"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-sasl"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type IMAPHandler struct {
	db *gorm.DB
}

func NewIMAPHandler(db *gorm.DB) *IMAPHandler {
	return &IMAPHandler{db: db}
}

func (h *IMAPHandler) TestConnection(c echo.Context) error {
	// get the body of the request
	body := c.Request().Body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to read request body: %v", err))
	}

	var credentials IMAPCredentials
	err = json.Unmarshal(bodyBytes, &credentials)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if credentials.Port == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid port")
	}

	log.Info("Testing connection to IMAP server %s:%d", credentials.Server, credentials.Port)

	// Connect to IMAP server
	im, err := client.DialTLS(fmt.Sprintf("%s:%d", credentials.Server, credentials.Port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to connect to IMAP server: %v", err))
	}
	defer im.Close()

	// login to the server
	if err := im.Authenticate(sasl.NewPlainClient("", credentials.Username, credentials.Password)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to authenticate: %v", err))
	}

	return c.JSON(http.StatusOK, "Connection successful")
}

// GetFolders handles fetching IMAP folders
func (h *IMAPHandler) GetFolders(c echo.Context) error {

	// fetch smtp details
	log.Info("Fetching IMAP folders")

	teamID := c.Get("teamID").(string)
	configId := c.QueryParam("config_id") // this is optional, if not provided, it will fetch all imap configs for the team

	var imapConfig *models.IMAPConfig
	var err error
	imapConfig, err = models.GetIMAPConfig(teamID, configId, h.db)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get IMAP config: %v", err))
	}

	// Connect to IMAP server
	im, err := client.DialTLS(fmt.Sprintf("%s:%d", imapConfig.Host, imapConfig.Port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to connect to IMAP server: %v", err))
	}
	defer im.Close()

	// login to the server
	if err := im.Authenticate(sasl.NewPlainClient("", imapConfig.Username, imapConfig.Password)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to authenticate: %v", err))
	}

	// Fetch folders
	folders := make(chan *imap.MailboxInfo, 20)
	err = im.List("", "*", folders)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch folders: %v", err))
	}

	folderList := make([]imap.MailboxInfo, 0, len(folders))
	for folder := range folders {
		folderList = append(folderList, *folder)
	}

	return c.JSON(http.StatusOK, folderList)
}

type pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetEmails handles fetching emails from a selected folder
func (h *IMAPHandler) GetEmails(c echo.Context) error {
	// this would be a get request with folder as a query param, limit as a query param, and offset as a query param
	folder := c.QueryParam("folder")
	since := c.QueryParam("since")
	before := c.QueryParam("before")
	sinceTime := time.Time{}
	beforeTime := time.Time{}
	var err error

	var pagination pagination = pagination{
		Limit:  10,
		Offset: 0,
	}

	if c.QueryParam("limit") != "" {
		pagination.Limit, err = strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit")
		}
	}

	if c.QueryParam("offset") != "" {
		pagination.Offset, err = strconv.Atoi(c.QueryParam("offset"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset")
		}
	}

	if folder == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Folder is required")
	}

	if since != "" {
		sinceTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid since time")
		}
	}

	if before != "" {
		beforeTime, err = time.Parse(time.RFC3339, before)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid before time")
		}
	}

	teamID := c.Get("teamID").(string)

	configId := c.QueryParam("config_id")

	imapConfig, err := models.GetIMAPConfig(teamID, configId, h.db)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get IMAP config: %v", err))
	}

	// Connect to IMAP server
	im, err := client.DialTLS(fmt.Sprintf("%s:%d", imapConfig.Host, imapConfig.Port), &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to connect to IMAP server: %v", err))
	}
	defer im.Close()

	// login to the server
	if err := im.Authenticate(sasl.NewPlainClient("", imapConfig.Username, imapConfig.Password)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to authenticate: %v", err))
	}

	// Select the folder
	selectStatus, err := im.Select(folder, true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to select folder: %v", err))
	}

	if selectStatus.Messages == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No emails found in the folder")
	}

	// fetch the total number of emails in the folder
	mailboxStatus, err := im.Status(folder, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get mailbox status: %v", err))
	}

	totalEmails := mailboxStatus.Messages

	// Fetch emails
	criteria := imap.NewSearchCriteria()
	if since != "" {
		criteria.Since = sinceTime
	}

	if before != "" {
		criteria.Before = beforeTime
	}

	// allow search by subject
	if c.QueryParam("subject") != "" {
		criteria.Header.Add("SUBJECT", c.QueryParam("subject"))
	}

	// allow search by body
	if c.QueryParam("body") != "" {
		criteria.Body = []string{c.QueryParam("body")}
	}

	// allow search by from
	if c.QueryParam("from") != "" {
		criteria.Header.Add("FROM", c.QueryParam("from"))
	}

	// allow search by to
	if c.QueryParam("to") != "" {
		criteria.Header.Add("TO", c.QueryParam("to"))
	}

	// allow search by cc
	if c.QueryParam("cc") != "" {
		criteria.Header.Add("CC", c.QueryParam("cc"))
	}

	// allow search by bcc
	if c.QueryParam("bcc") != "" {
		criteria.Header.Add("BCC", c.QueryParam("bcc"))
	}

	uids, err := im.Search(criteria)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch email UIDs: %v", err))
	}

	sort.Slice(uids, func(i, j int) bool {
		return uids[i] > uids[j]
	})

	// Validate pagination bounds
	if pagination.Offset >= len(uids) {
		return echo.NewHTTPError(http.StatusBadRequest, "Offset exceeds total number of emails")
	}

	// Adjust limit if it would exceed available emails
	if pagination.Offset+pagination.Limit > len(uids) {
		pagination.Limit = len(uids) - pagination.Offset
	}

	// Get slice of UIDs for this page
	uids = uids[pagination.Offset : pagination.Offset+pagination.Limit]

	// Create sequence set and add UIDs
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Create buffered channel sized to page limit
	emails := make(chan *imap.Message, pagination.Limit)

	// Fetch emails with RFC822 (full message content)

	fetchItems := []imap.FetchItem{imap.FetchRFC822Text, imap.FetchRFC822Header}

	if imapConfig.Host == "imap.gmail.com" {
		fetchItems = []imap.FetchItem{imap.FetchRFC822}
	}

	err = im.Fetch(seqset, fetchItems, emails)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch emails: %v", err))
	}

	// Convert map[int]*imap.Email to []imap.Email and process in parallel
	var wg sync.WaitGroup
	messages := make([]EmailMessage, len(uids))
	emailChan := make(chan *imap.Message, len(uids))

	// Start worker pool
	numWorkers := 10 // Adjust based on your needs
	errChan := make(chan error, numWorkers)

	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for email := range emailChan {
				idx := -1
				// Find corresponding index for this email
				for i, uid := range uids {
					if email.SeqNum == uid {
						idx = i
						break
					}
				}
				if idx == -1 {
					errChan <- fmt.Errorf("could not find index for email sequence number %d", email.SeqNum)
					return
				}

				msg := EmailMessage{
					Body:  "",
					Flags: email.Flags,
				}

				for _, literal := range email.Body {
					if literal == nil {
						continue
					}

					b := make([]byte, literal.Len())

					if _, err := io.ReadFull(literal, b); err != nil {
						errChan <- fmt.Errorf("failed to read message body: %v", err)
						return
					}

					emailReader := strings.NewReader(string(b))

					parsedMail, err := utils.ParseEmail(emailReader)

					if err != nil {
						continue
					}

					selectedBody := parsedMail.BodyText
					if parsedMail.BodyHTML != "" {
						selectedBody = parsedMail.BodyHTML
					}

					msg.Body = selectedBody
					msg.Attachments = parsedMail.Attachments
					msg.From = utils.FormatAddresses(parsedMail.From)
					msg.Subject = parsedMail.Subject
					msg.Date = parsedMail.Date.Format(time.RFC3339)
					msg.MessageID = parsedMail.MessageID
					msg.To = utils.FormatAddresses(parsedMail.To)
					msg.Cc = utils.FormatAddresses(parsedMail.Cc)
					msg.Bcc = utils.FormatAddresses(parsedMail.Bcc)
					msg.ReplyTo = utils.FormatAddresses(parsedMail.ReplyTo)
				}
				messages[idx] = msg
			}
		}(i)
	}

	// Feed emails to workers
	for email := range emails {
		emailChan <- email
	}
	close(emailChan)

	// Wait for all workers to complete
	wg.Wait()
	close(errChan)

	// Check for any errors from workers
	for err := range errChan {
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	response := FolderData{
		FolderName:  folder,
		TotalEmails: int(totalEmails),
		Limit:       pagination.Limit,
		Offset:      pagination.Offset,
		Emails:      messages,
	}

	return c.JSON(http.StatusOK, response)
}

// IMAPCredentials holds IMAP connection details
type IMAPCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Server   string `json:"host"`
	Port     int    `json:"port"`
}

// FolderData represents folder and email data
type FolderData struct {
	FolderName  string         `json:"folder_name"`
	TotalEmails int            `json:"total_emails"`
	Limit       int            `json:"limit"`
	Offset      int            `json:"offset"`
	Emails      []EmailMessage `json:"emails"`
}

type EmailMessage struct {
	Body        string                  `json:"body"`
	Flags       []string                `json:"flags"`
	To          string                  `json:"to"`
	Cc          string                  `json:"cc"`
	Bcc         string                  `json:"bcc"`
	From        string                  `json:"from"`
	Subject     string                  `json:"subject"`
	Date        string                  `json:"date"`
	MessageID   string                  `json:"message_id"`
	Attachments []utils.EmailAttachment `json:"attachments"`
	ReplyTo     string                  `json:"reply_to"`
}
