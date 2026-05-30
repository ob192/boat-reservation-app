package model

import "github.com/google/uuid"

// AuthUser is our internal projection of a Supabase user.
type AuthUser struct {
	ID          uuid.UUID
	Email       string
	AppMetadata map[string]any
}
