package model

import "github.com/google/uuid"

// AuthUser is our internal projection of a Supabase user.
type AuthUser struct {
	ID          uuid.UUID
	Email       string
	AppMetadata map[string]any
}

// adminUserIDs is the allowlist of users granted admin rights.
var adminUserIDs = map[uuid.UUID]bool{
	uuid.MustParse("b16773d5-a29c-489f-b4a8-380294ac7dc7"): true,
}

// IsAdmin reports whether the user is in the admin allowlist.
func (u *AuthUser) IsAdmin() bool {
	if u == nil {
		return false
	}
	return adminUserIDs[u.ID]
}
