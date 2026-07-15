package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	gocache "github.com/patrickmn/go-cache"
	supabase "github.com/supabase-community/auth-go"
	"github.com/supabase-community/auth-go/types"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
)

// fakeSupabase embeds the (huge) supabase.Client interface so we only have to
// implement the three methods authService actually touches. Calling anything
// else panics with a nil-interface dereference, which is the desired failure.
type fakeSupabase struct {
	supabase.Client

	lastToken    string
	getUserCalls int
	adminCalls   int

	getUser      func() (*types.UserResponse, error)
	adminGetUser func(req types.AdminGetUserRequest) (*types.AdminGetUserResponse, error)
}

func (f *fakeSupabase) WithToken(token string) supabase.Client {
	f.lastToken = token
	return f
}

func (f *fakeSupabase) GetUser() (*types.UserResponse, error) {
	f.getUserCalls++
	return f.getUser()
}

func (f *fakeSupabase) AdminGetUser(req types.AdminGetUserRequest) (*types.AdminGetUserResponse, error) {
	f.adminCalls++
	return f.adminGetUser(req)
}

func newAuthServiceWith(client supabase.Client) *authService {
	return &authService{
		client: client,
		cache:  gocache.New(5*time.Minute, 10*time.Minute),
		cfg:    &config.Config{SupabaseServiceRoleKey: "service-role-key"},
	}
}

func TestGetUserByToken(t *testing.T) {
	userID := uuid.New()

	t.Run("empty token rejected before any network call", func(t *testing.T) {
		svc := newAuthServiceWith(&fakeSupabase{})
		if _, err := svc.GetUserByToken(""); err == nil {
			t.Fatal("want error for empty token")
		}
	})

	t.Run("supabase error propagates", func(t *testing.T) {
		fake := &fakeSupabase{
			getUser: func() (*types.UserResponse, error) { return nil, errors.New("401") },
		}
		if _, err := newAuthServiceWith(fake).GetUserByToken("jwt-1"); err == nil {
			t.Fatal("want wrapped supabase error")
		}
		if fake.lastToken != "jwt-1" {
			t.Errorf("token %q not forwarded to supabase, got %q", "jwt-1", fake.lastToken)
		}
	})

	t.Run("nil response is an error", func(t *testing.T) {
		fake := &fakeSupabase{
			getUser: func() (*types.UserResponse, error) { return nil, nil },
		}
		if _, err := newAuthServiceWith(fake).GetUserByToken("jwt-1"); err == nil {
			t.Fatal("want error for nil user response")
		}
	})

	t.Run("success maps user and caches by token", func(t *testing.T) {
		fake := &fakeSupabase{
			getUser: func() (*types.UserResponse, error) {
				return &types.UserResponse{User: types.User{
					ID:          userID,
					Email:       "user@example.com",
					AppMetadata: map[string]any{"role": "customer"},
				}}, nil
			},
		}
		svc := newAuthServiceWith(fake)

		u, err := svc.GetUserByToken("jwt-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.ID != userID || u.Email != "user@example.com" || u.AppMetadata["role"] != "customer" {
			t.Errorf("mapped user wrong: %+v", u)
		}

		// Second call must be served from cache — no extra supabase round trip.
		u2, err := svc.GetUserByToken("jwt-1")
		if err != nil {
			t.Fatalf("unexpected error on cached call: %v", err)
		}
		if fake.getUserCalls != 1 {
			t.Errorf("supabase called %d times, want 1 (cache miss only)", fake.getUserCalls)
		}
		if u2.ID != userID {
			t.Errorf("cached user mismatch: %+v", u2)
		}

		// A different token is a different cache key.
		if _, err := svc.GetUserByToken("jwt-2"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fake.getUserCalls != 2 {
			t.Errorf("distinct token should miss the cache, calls = %d", fake.getUserCalls)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	userID := uuid.New()

	t.Run("supabase admin error propagates", func(t *testing.T) {
		fake := &fakeSupabase{
			adminGetUser: func(types.AdminGetUserRequest) (*types.AdminGetUserResponse, error) {
				return nil, errors.New("boom")
			},
		}
		if _, err := newAuthServiceWith(fake).GetUserByID(userID); err == nil {
			t.Fatal("want wrapped supabase error")
		}
	})

	t.Run("nil response is an error", func(t *testing.T) {
		fake := &fakeSupabase{
			adminGetUser: func(types.AdminGetUserRequest) (*types.AdminGetUserResponse, error) {
				return nil, nil
			},
		}
		if _, err := newAuthServiceWith(fake).GetUserByID(userID); err == nil {
			t.Fatal("want error for nil admin response")
		}
	})

	t.Run("success uses service-role key, requested id, and caches", func(t *testing.T) {
		var gotReq types.AdminGetUserRequest
		fake := &fakeSupabase{}
		fake.adminGetUser = func(req types.AdminGetUserRequest) (*types.AdminGetUserResponse, error) {
			gotReq = req
			return &types.AdminGetUserResponse{User: types.User{ID: userID, Email: "admin@example.com"}}, nil
		}
		svc := newAuthServiceWith(fake)

		u, err := svc.GetUserByID(userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fake.lastToken != "service-role-key" {
			t.Errorf("admin lookup must use the service-role key, got token %q", fake.lastToken)
		}
		if gotReq.UserID != userID {
			t.Errorf("requested user id = %v, want %v", gotReq.UserID, userID)
		}
		if u.ID != userID || u.Email != "admin@example.com" {
			t.Errorf("mapped user wrong: %+v", u)
		}

		if _, err := svc.GetUserByID(userID); err != nil {
			t.Fatalf("unexpected error on cached call: %v", err)
		}
		if fake.adminCalls != 1 {
			t.Errorf("supabase called %d times, want 1 (second call cached)", fake.adminCalls)
		}
	})
}

func TestMapSupabaseUser(t *testing.T) {
	id := uuid.New()

	u := mapSupabaseUser(id, "a@b.c", nil)
	if u.AppMetadata == nil {
		t.Error("nil app metadata must be replaced with an empty map")
	}
	if u.ID != id || u.Email != "a@b.c" {
		t.Errorf("mapped user wrong: %+v", u)
	}

	meta := map[string]any{"k": "v"}
	if got := mapSupabaseUser(id, "a@b.c", meta); got.AppMetadata["k"] != "v" {
		t.Errorf("metadata not preserved: %+v", got.AppMetadata)
	}
}

func TestEmailSendConfirmationNilSafety(t *testing.T) {
	svc := NewEmailService(&config.Config{}, testLogger())

	// Must not panic on nil booking or missing email.
	svc.SendConfirmation(nil)
	svc.SendConfirmation(&model.Booking{UserEmail: ""})
	svc.SendConfirmation(mkBooking())
}
