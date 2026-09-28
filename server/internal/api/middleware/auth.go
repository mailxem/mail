package middleware

import (
	"fmt"
	"kori/internal/assistant"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils/logger"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

var log = logger.New("auth_middleware")

type AuthMiddleware struct {
	jwtSecret string
	apiKeys   map[string]APIKeyInfo
}

type APIKeyInfo struct {
	TeamID      string
	Permissions []string
	ExpiresAt   time.Time
}

type Claims struct {
	UserID string   `json:"user_id"`
	TeamID string   `json:"team_id"`
	Email  string   `json:"email"`
	Role   string   `json:"role"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		apiKeys:   make(map[string]APIKeyInfo),
	}
}

func (m *AuthMiddleware) RegisterAPIKey(key string, info APIKeyInfo) {
	m.apiKeys[key] = info
}

func (m *AuthMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check API Key first
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey != "" {
				return m.validateAPIKey(c, apiKey, next)
			}

			// Check JWT Token
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization header")
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
			}

			if strings.Contains(c.Request().URL.Path, "/auth/google/callback") {
				return next(c)
			}

			return m.validateJWT(c, tokenParts[1], next, true)
		}
	}
}

func (m *AuthMiddleware) validateAPIKey(c echo.Context, key string, next echo.HandlerFunc) error {

	apiKey := &models.APIKey{}
	if err := db.DB.Where("key = ? AND is_deleted = false", assistant.Lookup(key)).Preload("Permissions").First(apiKey).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
	}

	if !assistant.ValidateActor(db.DB.WithContext(c.Request().Context()), apiKey) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Assistant access has changed")
	}

	// Check expiration
	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return echo.NewHTTPError(http.StatusUnauthorized, "API key has expired")
	}

	// check apikey permissions
	if err := m.validateAPIKeyPermissions(c, apiKey); err != nil {
		return err
	}

	// Set context values
	if apiKey.AssistantUserID != "" {
		c.Set("userID", apiKey.AssistantUserID)
		c.Set("assistantWrite", apiKey.AssistantWrite)
	}
	c.Set("teamID", apiKey.TeamID)
	c.Set("isAPIKey", true)
	c.Set("permissions", apiKey.Permissions)

	return next(c)
}

func (m *AuthMiddleware) validateAPIKeyPermissions(c echo.Context, apiKey *models.APIKey) error {
	method := c.Request().Method
	requiredScope := GetRequiredPermissionForMethod(method)
	if requiredScope == "" {
		return echo.NewHTTPError(http.StatusForbidden, "Invalid request method")
	}

	// Get the request path to determine the resource
	path := c.Request().URL.Path
	resource := m.getResourceFromPath(path)
	requiredPermission := fmt.Sprintf("%s:%s", resource, requiredScope)

	// Check permissions using the common validation function
	if err := ValidateAPIKeyPermissions(c.Request().Context(), db.DB, apiKey.ID, []string{requiredPermission}); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
	}

	// track api key usage
	db.DB.Model(&models.APIKey{}).Where("id = ?", apiKey.ID).Update("last_used_at", time.Now())
	db.DB.Model(&models.APIKeyUsage{}).Create(&models.APIKeyUsage{
		APIKeyID:  apiKey.ID,
		APIKey:    apiKey,
		Endpoint:  path,
		Method:    method,
		Timestamp: time.Now(),
		Success:   true,
	})

	// Set context values
	c.Set("apiKeyID", apiKey.ID)
	c.Set("teamID", apiKey.TeamID)
	c.Set("isAPIKey", true)
	c.Set("permissions", []string{requiredPermission})

	return nil
}

func (m *AuthMiddleware) getResourceFromPath(path string) string {
	// Remove API version prefix if exists
	path = strings.TrimPrefix(path, "/api/v1")

	// Marketing routes use the existing resource permissions rather than an
	// unrelated marketing-wide grant.
	if strings.HasPrefix(path, "/marketing/") {
		resource := strings.Split(strings.TrimPrefix(path, "/marketing/"), "/")[0]
		switch resource {
		case "newsletters", "campaign-drafts":
			return "campaigns"
		case "contact-batch", "contacts":
			return "contacts"
		case "options":
			return "lists"
		}
	}
	if path == "/mailing-lists" || strings.HasPrefix(path, "/mailing-lists/") {
		return "lists"
	}
	// Split path and get the first segment
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (m *AuthMiddleware) validateJWT(c echo.Context, tokenString string, next echo.HandlerFunc, checkMethod bool) error {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Error("Error parsing JWT token: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token has expired")
	}

	// Verify auth transaction
	transaction := &models.AuthTransaction{}
	if err := db.DB.Where("user_id = ? AND team_id = ? AND token = ?",
		claims.UserID, claims.TeamID, tokenString).First(transaction).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Auth transaction not found")
	}

	// Verify user exists
	user := &models.User{}
	if err := db.DB.Where("id = ?", claims.UserID).First(user).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	log.Info("User found: %s", user.Email)

	// Verify team membership
	team := &models.Team{}
	if err := db.DB.Joins("JOIN users ON users.team_id = teams.id").
		Where("teams.id = ? AND users.id = ?", claims.TeamID, claims.UserID).
		First(team).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Team not found")
	}

	requestContentType := strings.Split(c.Request().Header.Get("Content-Type"), ";")[0]

	log.Info("Request content type: %s", requestContentType)

	if (c.Request().Method == "POST" || c.Request().Method == "PUT") && requestContentType != "multipart/form-data" {
		if err := injectRequestTeamID(c.Request(), team.ID); err != nil {
			return err
		}
	}

	// Check method-based permissions
	method := c.Request().Method
	requiredScope := GetRequiredPermissionForMethod(method)
	if requiredScope == "" {
		return echo.NewHTTPError(http.StatusForbidden, "Invalid request method")
	}

	// Admin role has all permissions
	if user.Role == models.UserRoleAdmin || user.Role == models.UserRoleSuperAdmin {
		c.Set("hasAdminAccess", true)
	} else if checkMethod {
		// Check if user has the required scope
		hasPermission := false
		for _, scope := range claims.Scopes {
			if ValidateMethodPermission(method, scope) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}

	// Set context values
	c.Set("userID", claims.UserID)
	c.Set("teamID", claims.TeamID)
	c.Set("email", claims.Email)
	c.Set("role", string(user.Role))
	c.Set("scopes", claims.Scopes)
	c.Set("isAPIKey", false)

	return next(c)
}

// GetUserID Helper functions to get values from context
func GetUserID(c echo.Context) string {
	if id, ok := c.Get("userID").(string); ok {
		return id
	}
	return ""
}

func GetTeamID(c echo.Context) string {
	if id, ok := c.Get("teamID").(string); ok {
		return id
	}
	return ""
}

func GetUserRole(c echo.Context) string {
	if role, ok := c.Get("role").(string); ok {
		return role
	}
	return ""
}

func GetScopes(c echo.Context) []string {
	if scopes, ok := c.Get("scopes").([]string); ok {
		return scopes
	}
	return nil
}

func IsAPIKey(c echo.Context) bool {
	if isAPIKey, ok := c.Get("isAPIKey").(bool); ok {
		return isAPIKey
	}
	return false
}

func HasPermission(c echo.Context, requiredScope string) bool {
	if IsAPIKey(c) {
		if permissions, ok := c.Get("permissions").([]string); ok {
			for _, p := range permissions {
				if p == "ADMIN" || p == requiredScope {
					return true
				}
			}
		}
		return false
	}

	// For JWT tokens, check role and scopes
	role := GetUserRole(c)
	if role == "admin" {
		return true
	}

	scopes := GetScopes(c)
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// SessionMiddleware authenticates a current human membership. The dedicated
// credential handler derives read/write scope itself; API keys cannot mint keys.
func (m *AuthMiddleware) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			parts := strings.Split(c.Request().Header.Get("Authorization"), " ")
			if len(parts) != 2 || parts[0] != "Bearer" || c.Request().Header.Get("X-API-Key") != "" {
				return echo.NewHTTPError(401, "Sign in required")
			}
			return m.validateJWT(c, parts[1], next, false)
		}
	}
}
