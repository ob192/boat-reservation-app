package model

import "github.com/google/uuid"

// AuthUser is our internal projection of a Supabase user.
// We intentionally don't depend on the supabase-go types in the rest of the code,
// to keep the boundary clean and the rest of the codebase compilable without the SDK.
type AuthUser struct {
	ID          uuid.UUID
	Email       string
	AppMetadata map[string]any
}

// IsAdmin reports whether the user has app_metadata.role == "admin".
func (u *AuthUser) IsAdmin() bool {
	if u == nil || u.AppMetadata == nil {
		return false
	}
	role, _ := u.AppMetadata["role"].(string)
	return role == "admin"
}
