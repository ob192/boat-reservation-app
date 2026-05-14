package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

// ContextUserDetails is the gin.Context key that stores the *model.AuthUser
// after AuthMiddleware has validated the JWT.
const ContextUserDetails = "UserDetails"

// ExtractBearerToken pulls the bearer token from the Authorization header.
// Spec-compliant: requires a literal "Bearer " prefix and a non-empty token.
func ExtractBearerToken(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", errors.New("missing Authorization header")
	}
	const prefix = "Bearer "
	trimmed := strings.TrimSpace(header)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", errors.New(`invalid auth scheme, expected "Bearer"`)
	}
	token := strings.TrimSpace(trimmed[len(prefix):])
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}

// AuthMiddleware validates the JWT and attaches the resolved *model.AuthUser to context.
// On any failure it aborts with 401 + a spec-defined message.
func AuthMiddleware(authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := ExtractBearerToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": httpx.CodeNotAuthenticated})
			return
		}
		user, err := authSvc.GetUserByToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": httpx.CodeSessionExpired})
			return
		}
		c.Set(ContextUserDetails, user)
		c.Next()
	}
}

// AdminMiddleware requires that the already-authenticated user has app_metadata.role == "admin".
// Must be installed *after* AuthMiddleware in the route group.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetUserFromContext(c)
		if err != nil || !user.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": httpx.CodeForbidden})
			return
		}
		c.Next()
	}
}

// GetUserFromContext returns the *model.AuthUser attached by AuthMiddleware.
func GetUserFromContext(c *gin.Context) (*model.AuthUser, error) {
	raw, ok := c.Get(ContextUserDetails)
	if !ok {
		return nil, fmt.Errorf("context key %q not found", ContextUserDetails)
	}
	switch v := raw.(type) {
	case *model.AuthUser:
		if v == nil {
			return nil, fmt.Errorf("nil *model.AuthUser in context")
		}
		return v, nil
	case model.AuthUser:
		return &v, nil
	default:
		return nil, fmt.Errorf("unexpected type %T for context key %q", raw, ContextUserDetails)
	}
}

// GetUserIDFromContext is a convenience wrapper around GetUserFromContext.
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	user, err := GetUserFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}
