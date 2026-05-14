package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	supabase "github.com/supabase-community/auth-go"
	"github.com/supabase-community/auth-go/types"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
)

// AuthService validates Supabase JWTs and returns our internal AuthUser.
//
// Caches the token → user lookup for 5 minutes in-process to avoid hammering
// Supabase on every request. The cache key is the raw JWT string.
type AuthService interface {
	GetUserByToken(token string) (*model.AuthUser, error)
	GetUserByID(userID uuid.UUID) (*model.AuthUser, error)
}

type authService struct {
	client supabase.Client
	cache  *gocache.Cache
	cfg    *config.Config
}

// NewAuthService wires up the Supabase auth client and a 5-minute in-process cache.
func NewAuthService(cfg *config.Config) AuthService {
	client := supabase.New(cfg.SupabaseProjectReference, cfg.SupabaseAnonKey)
	return &authService{
		client: client,
		cache:  gocache.New(5*time.Minute, 10*time.Minute),
		cfg:    cfg,
	}
}

func (s *authService) GetUserByToken(token string) (*model.AuthUser, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	cacheKey := "auth:token:" + token
	if hit, ok := s.cache.Get(cacheKey); ok {
		if u, ok := hit.(*model.AuthUser); ok {
			return u, nil
		}
	}

	resp, err := s.client.WithToken(token).GetUser()
	if err != nil {
		return nil, fmt.Errorf("supabase GetUser: %w", err)
	}
	if resp == nil {
		return nil, errors.New("no user returned for token")
	}

	user := mapSupabaseUser(resp.ID, resp.Email, resp.AppMetadata)
	s.cache.SetDefault(cacheKey, user)
	return user, nil
}

func (s *authService) GetUserByID(userID uuid.UUID) (*model.AuthUser, error) {
	cacheKey := "auth:userid:" + userID.String()
	if hit, ok := s.cache.Get(cacheKey); ok {
		if u, ok := hit.(*model.AuthUser); ok {
			return u, nil
		}
	}

	resp, err := s.client.
		WithToken(s.cfg.SupabaseServiceRoleKey).
		AdminGetUser(types.AdminGetUserRequest{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("supabase AdminGetUser: %w", err)
	}
	if resp == nil {
		return nil, errors.New("no user found")
	}

	user := mapSupabaseUser(resp.ID, resp.Email, resp.AppMetadata)
	s.cache.SetDefault(cacheKey, user)
	return user, nil
}

func mapSupabaseUser(id uuid.UUID, email string, appMeta map[string]any) *model.AuthUser {
	if appMeta == nil {
		appMeta = map[string]any{}
	}
	return &model.AuthUser{
		ID:          id,
		Email:       email,
		AppMetadata: appMeta,
	}
}
