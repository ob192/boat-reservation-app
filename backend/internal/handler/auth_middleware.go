package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

const ContextUserDetails = "UserDetails"

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

// AdminMiddleware checks the `admins` DB table for the authenticated user.
// Must be installed *after* AuthMiddleware in the route group.
func AdminMiddleware(adminRepo repository.AdminRepository) gin.HandlerFunc {

	return func(c *gin.Context) {
		user, err := GetUserFromContext(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": httpx.CodeForbidden})
			return
		}
		ok, err := adminRepo.IsAdmin(c.Request.Context(), user.ID)
		if err != nil {
			// Treat a DB error conservatively: deny access and let the request logger
			// surface the underlying error via the 503 path is not applicable here
			// because we're in a middleware — just log and forbid.
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": httpx.CodeForbidden})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": httpx.CodeForbidden})
			return
		}
		c.Next()
	}
}

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

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	user, err := GetUserFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}
