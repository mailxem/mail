package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/utils"

	"crypto/rand"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyResetCodeRequest struct {
	Code     string `json:"code" validate:"required"`
	Password string `json:"new_password" validate:"required,min=8"`
}

type GoogleAuthRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// Register handles the registration of a new user by validating input, hashing the password, storing user data, and assigning permissions.
// @Summary Register a new user
// @Description Register a new user with email, password and name details
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]string "User registered successfully"
// @Failure 400 {object} map[string]string "Validation error or email exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var createTeam bool = true
	var team models.Team
	var user models.User

	// check if user already exists
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			createTeam = true
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already exists"})
		}
	}

	// check if user is already invited
	var invite models.TeamInvite
	if err := h.db.Where("email = ? AND status = ? AND expires_at > ?", req.Email, models.InviteStatusPending, time.Now()).First(&invite).Error; err == nil {
		// accept invite
		invite.Status = models.InviteStatusAccepted
		if err := h.db.Save(&invite).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to accept invite"})
		}
		createTeam = false
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	if createTeam {
		// create a team
		team = models.Team{
			Name: req.FirstName + "'s Team", // Example team name based on the user's first name
		}

		if err = tx.Create(&team).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create team"})
		}

	}

	user = models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      models.UserRoleAdmin, // Default role for new users
		TeamID:    team.ID,
	}

	if invite.Status == models.InviteStatusAccepted {
		user.Role = invite.Role
		user.TeamID = invite.TeamID
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email already exists"})
	}

	// Assign default permissions based on role
	if err := models.AssignDefaultPermissions(tx, &user); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to assign permissions"})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	events.Emit("users.created", &user)

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Login handles user login by validating credentials, generating a JWT token, and returning it.
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, err := utils.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	authtransaction := &models.AuthTransaction{
		UserID:    user.ID,
		TeamID:    user.TeamID,
		Token:     token,
		Refresh:   refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}

	if err := h.db.Create(authtransaction).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create auth transaction"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token, "refresh_token": refreshToken})
}

// RequestPasswordReset handles the request to reset a user's password by generating a reset code, storing it, and sending an email.
// @Summary Request password reset
// @Description Request a password reset code to be sent via email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Email for password reset"
// @Success 200 {object} map[string]string "Reset code sent if email exists"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	tx := h.db.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusOK, map[string]string{"message": "If the email exists, a reset code will be sent"})
	}

	code, err := generateResetCode(10)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate reset code"})
	}

	reset := models.PasswordReset{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := tx.Create(&reset).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create reset code"})
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	reset.User = &user

	events.Emit("password.reset", &reset)

	return c.JSON(http.StatusOK, map[string]string{"message": "If the email exists, a reset code will be sent"})
}

// VerifyResetCode handles the verification of a reset code, updating the user's password, and marking the reset code as used.
// @Summary Verify reset code and set new password
// @Description Verify password reset code and update password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyResetCodeRequest true "Reset code verification and new password"
// @Success 200 {object} map[string]string "Password reset successful"
// @Failure 400 {object} map[string]string "Invalid or expired reset code"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/password-reset/verify [post]
func (h *AuthHandler) VerifyResetCode(c echo.Context) error {
	var req VerifyResetCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var reset models.PasswordReset
	if err := h.db.Where("code = ? AND used = ? AND expires_at > ?",
		req.Code, false, time.Now()).First(&reset).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired reset code"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	var user models.User
	if err := h.db.Where("id = ?", reset.UserID).First(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	h.db.Model(&user).Update("password", string(hashedPassword))
	h.db.Model(&reset).Update("used", true)

	return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

// GenerateResetCode generates a cryptographically secure random code
// without special characters, using crypto/rand
func generateResetCode(length int) (string, error) {
	// Generate random bytes (we need more than length because
	// of the base64 encoding and replacement of special chars)
	buffer := make([]byte, length*2)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	// Convert to base64 string
	encoded := base64.StdEncoding.EncodeToString(buffer)

	// Remove non-alphanumeric characters
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1 // Will be removed
	}, encoded)

	// Trim to desired length
	if len(result) > length {
		result = result[:length]
	}

	return result, nil
}

// ListUsers returns a list of all users (admin only)
// @Summary List all users
// @Description Get a list of all users (requires admin permissions)
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users [get]
func (h *AuthHandler) ListUsers(c echo.Context) error {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users"})
	}
	return c.JSON(http.StatusOK, users)
}

// GetUser returns details of a specific user (admin only)
// @Summary Get user details
// @Description Get details of a specific user (requires admin permissions)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [get]
func (h *AuthHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}
	return c.JSON(http.StatusOK, user)
}

// UpdateUser updates a user's details (admin only)
// @Summary Update user details
// @Description Update details of a specific user (requires admin permissions)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.User true "Updated user details"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [put]
func (h *AuthHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	// Only update allowed fields
	var updateData struct {
		FirstName        string          `json:"first_name"`
		LastName         string          `json:"last_name"`
		Role             models.UserRole `json:"role"`
		ProfilePictureID string          `json:"profilePictureId"`
	}

	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Validate role
	if !models.IsValidUserRole(updateData.Role) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid role"})
	}

	user.FirstName = updateData.FirstName
	user.LastName = updateData.LastName
	user.Role = updateData.Role

	if updateData.ProfilePictureID != "" {
		user.ProfilePictureID = updateData.ProfilePictureID
	}

	if err := h.db.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user (admin only)
// @Summary Delete user
// @Description Delete a specific user (requires admin permissions)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/users/{id} [delete]
func (h *AuthHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user"})
	}

	if err := h.db.Delete(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// RefreshToken refreshes a user's access token using their refresh token
// @Summary Refresh access token
// @Description Get a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} map[string]string "New access token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// get refresh token from request
	refreshToken := input.RefreshToken

	// validate refresh token
	_, err := utils.ValidateRefreshToken(refreshToken, os.Getenv("JWT_SECRET"))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
	}

	// check in db if refresh token is valid
	var authTransaction models.AuthTransaction
	if err := h.db.Where("refresh = ? AND expires_at > ?", refreshToken, time.Now()).First(&authTransaction).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
	}

	// get user from claims
	var user models.User
	if err := h.db.First(&user, authTransaction.UserID).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	// generate new access token
	accessToken, err := utils.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	// save new access token to db
	authTransaction.Token = accessToken
	if err := h.db.Save(&authTransaction).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save access token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": accessToken, "exp": "15m"})
}

// GetMe returns the current user
// @Summary Get current user
// @Description Get details of the current authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.User
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c echo.Context) error {
	userId := c.Get("userID").(string)

	var user models.User
	if err := h.db.Where("id = ?", userId).Preload("Team").First(&user).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

// InviteUserRequest is the request body for inviting a user to a team
// @Description Send an invitation email to a user to join a team
type InviteUserRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2"`
	Role  string `json:"role" default:"MEMBER" validate:"required,oneof=MEMBER ADMIN SUPER_ADMIN"`
}

// InviteUser handles sending invitations to new team members
// @Summary Invite a user to join a team
// @Description Send an invitation email to a user to join a team
// @Tags auth
// @Accept json
// @Produce json
// @Param request body InviteUserRequest true "Invitation details"
// @Success 201 {object} map[string]string "Invitation sent successfully"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/invite [post]
func (h *AuthHandler) InviteUser(c echo.Context) error {
	// 🔒 Get current user ID from context
	userID := c.Get("userID").(string)
	teamID := c.Get("teamID").(string)

	log.Info("Inviting user %s to team %s", userID, teamID)

	var request InviteUserRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// 🔍 Validate invite data
	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Generate invite code
	code, err := utils.GenerateRandomString(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate invite code"})
	}

	// 💾 Save invitation
	invite := models.TeamInvite{
		Code:      code,
		ExpiresAt: time.Now().Add(24 * 7 * time.Hour),
		InviterID: userID,
		TeamID:    teamID,
		Status:    models.InviteStatusPending,
		Role:      models.UserRole(request.Role),
		Email:     request.Email,
		Name:      request.Name,
	}

	// delete previous invitation
	if err := h.db.Where("email = ? AND team_id = ? AND status = ?", request.Email, teamID, models.InviteStatusPending).Delete(&models.TeamInvite{}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete previous invitation"})
	}

	// 💾 Save invitation
	if err := h.db.Create(&invite).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create invitation"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Invitation sent successfully"})
}

// ResendInvite handles resending an invitation to a user

// @Summary Resend an invitation to a user
// @Description Resend an invitation to a user
// @Tags auth
// @Accept json
// @Produce json
// @Param code path string true "Invitation code"
// @Success 200 {object} map[string]string "Invitation resent successfully"
// @Failure 400 {object} map[string]string "Invalid invitation"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/invite/resend/{code} [post]
func (h *AuthHandler) ResendInvite(c echo.Context) error {
	code := c.Param("code")

	// 🔍 Find invitation
	var invite models.TeamInvite
	if err := h.db.Where("code = ? AND status = ? AND expires_at > ?",
		code, models.InviteStatusPending, time.Now()).First(&invite).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired invitation"})
	}

	// 🔒 Get current user ID from context
	userID := c.Get("userID").(string)

	// 🔍 Check if user is the inviter
	if invite.InviterID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to resend this invitation"})
	}

	// 🔒 Generate new invite code
	newCode, err := utils.GenerateRandomString(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate new invite code"})
	}

	// 💾 Save new invitation
	invite.Code = newCode
	if err := h.db.Save(&invite).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save new invitation"})
	}

	// 🔔 Send new invitation email
	events.Emit("invite.resend", &invite)

	return c.JSON(http.StatusOK, map[string]string{"message": "Invitation resent successfully"})
}

// AcceptInvite handles accepting team invitations
// @Summary Accept a team invitation
// @Description Accept an invitation to join a team
// @Tags auth
// @Accept json
// @Produce json
// @Param code path string true "Invitation code"
// @Success 200 {object} map[string]string "Invitation accepted successfully"
// @Failure 400 {object} map[string]string "Invalid invitation"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/invite/accept/{code} [post]
type AcceptInviteRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) AcceptInvite(c echo.Context) error {
	code := c.Param("code")

	// 🔒 Get password from request body
	var req AcceptInviteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// 🔍 Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// 🔐 Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	tx := h.db.WithContext(c.Request().Context()).Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(500, "Unable to accept invitation")
	}
	defer tx.Rollback()
	var invite models.TeamInvite
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code = ? AND status = ? AND expires_at > ?", code, models.InviteStatusPending, time.Now()).First(&invite).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired invitation"})
	}

	// 👤 Create new user
	newUser := models.User{
		Email:     invite.Email,
		FirstName: invite.Name,
		LastName:  "",
		Password:  string(hashedPassword),
		TeamID:    invite.TeamID,
		Role:      invite.Role, // Default role for invited users
	}

	if err := tx.Create(&newUser).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// ✅ Update invitation status
	invite.Status = models.InviteStatusAccepted
	if err := tx.Save(&invite).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update invitation"})
	}

	// Assign default permissions based on role
	if err := models.AssignDefaultPermissions(tx, &newUser); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to assign permissions"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Invitation accepted successfully"})
}

// CheckInvite handles checking if an invitation code is valid
// @Summary Check invitation validity
// @Description Check if an invitation code is valid and not expired
// @Tags auth
// @Accept json
// @Produce json
// @Param code path string true "Invitation code"
// @Success 200 {object} map[string]interface{} "Invitation details"
// @Failure 400 {object} map[string]string "Invalid or expired invitation"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/invite/{code} [get]
func (h *AuthHandler) CheckInvite(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invitation code is required"})
	}

	// 🔍 Find invitation by code
	var invite models.TeamInvite
	if err := h.db.Preload("Team").Where("code = ? AND status = ? AND expires_at > ?",
		code, models.InviteStatusPending, time.Now()).First(&invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or expired invitation"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check invitation"})
	}

	// Return invitation details (excluding sensitive information)
	return c.JSON(http.StatusOK, map[string]any{
		"valid":      true,
		"email":      invite.Email,
		"name":       invite.Name,
		"role":       invite.Role,
		"team_name":  invite.Team.Name,
		"expires_at": invite.ExpiresAt,
	})
}

// DeleteInvite handles deleting team invitations
// @Summary Delete a team invitation
// @Description Delete a pending team invitation
// @Tags auth
// @Accept json
// @Produce json
// @Param code path string true "Invitation code"
// @Success 200 {object} map[string]string "Invitation deleted successfully"
// @Failure 400 {object} map[string]string "Invalid invitation"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/invite/{code} [delete]
func (h *AuthHandler) DeleteInvite(c echo.Context) error {
	// 🔒 Get current user ID from context
	userID := c.Get("userID").(string)
	inviteCode := c.Param("code")

	// 🔍 Find and validate invitation
	var invite models.TeamInvite
	if err := h.db.Where("code = ? AND inviter_id = ?", inviteCode, userID).First(&invite).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invitation not found"})
	}

	// ❌ Delete invitation
	if err := h.db.Delete(&invite).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete invitation"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Invitation deleted successfully"})
}

// GoogleAuth handles authentication with Google OAuth
// @Summary Authenticate with Google
// @Description Authenticate user using Google OAuth ID token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleAuthRequest true "Google ID token"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "No access token provided"
// @Failure 400 {object} map[string]string "Failed to parse user data from Google"
// @Failure 401 {object} map[string]string "Failed to get user data from Google"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleAuthCallback(c echo.Context) error {
	accessToken := c.Request().Header.Get("Authorization")

	if accessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No access token provided"})
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	// get user data from google
	userDataBytes, err := utils.GetUserDataFromGoogle(accessToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get user data from Google"})
	}

	// parse user data
	var userData map[string]interface{}
	if err := json.Unmarshal(userDataBytes, &userData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to parse user data from Google"})
	}

	// Start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to start transaction"})
	}

	// Check if user exists with either email or provider ID
	var user models.User
	err = tx.Where("email = ? OR (provider = ? AND provider_id = ?)",
		userData["email"], "google", userData["id"]).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Check for pending team invitation first
			var invite models.TeamInvite
			inviteErr := tx.Where("email = ? AND status = ? AND expires_at > ?",
				userData["email"], models.InviteStatusPending, time.Now()).First(&invite).Error

			var teamID string
			var userRole models.UserRole

			if inviteErr == nil {
				// Use the invited team and role
				teamID = invite.TeamID
				userRole = invite.Role

				// Mark invitation as accepted
				invite.Status = models.InviteStatusAccepted
				if err := tx.Save(&invite).Error; err != nil {
					tx.Rollback()
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update invitation"})
				}
			} else {
				// No invitation found, create new team
				team := models.Team{
					Name: userData["given_name"].(string) + "'s Team",
				}

				if err = tx.Create(&team).Error; err != nil {
					tx.Rollback()
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create team"})
				}

				teamID = team.ID
				userRole = models.UserRoleAdmin
			}
			var fileModel *models.File
			// download the profile picture
			if photoURL, ok := userData["photoUrl"].(string); ok {
				profilePicture, err := http.Get(photoURL)
				if err != nil {
					// Log the error but do not affect account creation
					log.Error("Failed to download profile picture", err)
				} else {
					defer profilePicture.Body.Close()
					// read the profile picture
					profilePictureBytes, err := io.ReadAll(profilePicture.Body)
					if err != nil {
						log.Error("Failed to read profile picture", err)
					} else {
						// Get storage handler
						storage := GetStorageHandler()
						if storage != nil {
							// Create a temporary user ID since we don't have the real one yet
							tempUserID := uuid.New().String()
							// upload the profile picture to s3
							profilePictureURL, err := storage.UploadFile(c.Request().Context(), profilePictureBytes, tempUserID, "public-read", "image/jpeg")
							if err != nil {
								log.Error("Failed to upload profile picture", err)
							} else {
								fileModel = &models.File{
									TeamID: teamID,
									Path:   profilePictureURL[strings.LastIndex(profilePictureURL, "/")+1:],
									Name:   "profile_picture.jpg",
									Size:   int64(len(profilePictureBytes)),
									Type:   "image/jpeg",
								}
								if err := tx.Create(fileModel).Error; err != nil {
									log.Error("Failed to create profile picture", err)
									fileModel = nil
								}
							}
						} else {
							log.Error("Storage handler not configured", nil)
						}
					}
				}
			}

			// Create user with both google and local auth capabilities
			user = models.User{
				Email:        userData["email"].(string),
				FirstName:    userData["given_name"].(string),
				LastName:     userData["family_name"].(string),
				Role:         userRole,
				TeamID:       teamID,
				Provider:     "google",
				ProviderID:   userData["id"].(string),
				Password:     "", // Empty password for google users
				ProviderData: datatypes.JSON{},
			}

			// Only set ProfilePictureID if we successfully created the file
			if fileModel != nil && fileModel.ID != "" {
				user.ProfilePictureID = fileModel.ID
			} else {
				user.ProfilePictureID = "5574fee5-3ce4-49e5-af2e-21361fc433e4"
			}

			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
			}

			// Assign default permissions
			if err := models.AssignDefaultPermissions(tx, &user); err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to assign permissions"})
			}

			// Emit different events based on invitation status
			if inviteErr == nil {
				events.Emit("users.invite_accepted", &user)
			} else {
				events.Emit("users.created", &user)
			}
		} else {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check user existence"})
		}
	} else {
		// If user exists but hasn't used Google auth before, link the accounts
		if user.Provider == "local" {
			user.Provider = "google"
			user.ProviderID = userData["id"].(string)
			if user.ProfilePictureID == "" {
				user.ProfilePictureID = "5574fee5-3ce4-49e5-af2e-21361fc433e4"
			}
			if err := tx.Save(&user).Error; err != nil {
				tx.Rollback()
				fmt.Println("Failed to update user", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	// Create auth transaction
	authtransaction := &models.AuthTransaction{
		UserID:    user.ID,
		TeamID:    user.TeamID,
		Token:     jwtToken,
		Refresh:   refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // 30 days
	}

	if err := h.db.Create(authtransaction).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create auth transaction"})
	}

	events.Emit("users.google_auth", &user)

	return c.JSON(http.StatusOK, map[string]string{
		"token":         jwtToken,
		"refresh_token": refreshToken,
	})
}
