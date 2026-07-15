package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
)

func TestExtractBearerToken(t *testing.T) {
	makeCtx := func(header string) *gin.Context {
		c, _ := gin.CreateTestContext(nil)
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		c.Request = req
		return c
	}

	t.Run("valid", func(t *testing.T) {
		tok, err := ExtractBearerToken(makeCtx("Bearer abc.def.ghi"))
		if err != nil || tok != "abc.def.ghi" {
			t.Errorf("got (%q, %v)", tok, err)
		}
	})

	t.Run("surrounding whitespace tolerated", func(t *testing.T) {
		tok, err := ExtractBearerToken(makeCtx("  Bearer   abc  "))
		if err != nil || tok != "abc" {
			t.Errorf("got (%q, %v)", tok, err)
		}
	})

	t.Run("failures", func(t *testing.T) {
		bad := map[string]string{
			"missing header":   "",
			"wrong scheme":     "Basic dXNlcjpwYXNz",
			"lowercase bearer": "bearer abc",
			"scheme only":      "Bearer ",
		}
		for name, header := range bad {
			t.Run(name, func(t *testing.T) {
				if _, err := ExtractBearerToken(makeCtx(header)); err == nil {
					t.Errorf("header %q should be rejected", header)
				}
			})
		}
	})
}

func authRouter(authSvc service.AuthService) *gin.Engine {
	r := gin.New()
	r.GET("/protected", AuthMiddleware(authSvc), func(c *gin.Context) {
		u, err := GetUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"email": u.Email})
	})
	return r
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("no header is 401 NOT_AUTHENTICATED", func(t *testing.T) {
		w := doRequest(t, authRouter(&mockAuthService{}), http.MethodGet, "/protected", "", nil)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d", w.Code)
		}
		resp := decodeJSON[map[string]string](t, w)
		if resp["message"] != "NOT_AUTHENTICATED" {
			t.Errorf("message = %q", resp["message"])
		}
	})

	t.Run("bad token is 401 SESSION_EXPIRED", func(t *testing.T) {
		svc := &mockAuthService{
			getUserByToken: func(string) (*model.AuthUser, error) { return nil, errors.New("expired") },
		}
		w := doRequest(t, authRouter(svc), http.MethodGet, "/protected", "", map[string]string{
			"Authorization": "Bearer stale-token",
		})
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d", w.Code)
		}
		resp := decodeJSON[map[string]string](t, w)
		if resp["message"] != "SESSION_EXPIRED" {
			t.Errorf("message = %q", resp["message"])
		}
	})

	t.Run("valid token stashes the user for the handler", func(t *testing.T) {
		user := testUser()
		var gotToken string
		svc := &mockAuthService{
			getUserByToken: func(token string) (*model.AuthUser, error) {
				gotToken = token
				return user, nil
			},
		}
		w := doRequest(t, authRouter(svc), http.MethodGet, "/protected", "", map[string]string{
			"Authorization": "Bearer good-token",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotToken != "good-token" {
			t.Errorf("token = %q", gotToken)
		}
		resp := decodeJSON[map[string]string](t, w)
		if resp["email"] != user.Email {
			t.Errorf("handler saw %q", resp["email"])
		}
	})
}

func adminRouter(repo repository.AdminRepository, user *model.AuthUser) *gin.Engine {
	r := gin.New()
	g := r.Group("/")
	if user != nil {
		g.Use(withUser(user))
	}
	g.Use(AdminMiddleware(repo))
	g.GET("/admin/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

func TestAdminMiddleware(t *testing.T) {
	t.Run("no authenticated user is 403", func(t *testing.T) {
		w := doRequest(t, adminRouter(&mockAdminRepo{}, nil), http.MethodGet, "/admin/ping", "", nil)
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("lookup failure denies conservatively", func(t *testing.T) {
		repo := &mockAdminRepo{
			isAdmin: func(context.Context, uuid.UUID) (bool, error) { return false, errors.New("db down") },
		}
		w := doRequest(t, adminRouter(repo, testUser()), http.MethodGet, "/admin/ping", "", nil)
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("non-admin is 403 — JWT alone is not enough", func(t *testing.T) {
		repo := &mockAdminRepo{
			isAdmin: func(context.Context, uuid.UUID) (bool, error) { return false, nil },
		}
		w := doRequest(t, adminRouter(repo, testUser()), http.MethodGet, "/admin/ping", "", nil)
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d", w.Code)
		}
	})

	t.Run("admin passes through with the right user id", func(t *testing.T) {
		user := testUser()
		var checked uuid.UUID
		repo := &mockAdminRepo{
			isAdmin: func(_ context.Context, id uuid.UUID) (bool, error) {
				checked = id
				return true, nil
			},
		}
		w := doRequest(t, adminRouter(repo, user), http.MethodGet, "/admin/ping", "", nil)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
		if checked != user.ID {
			t.Errorf("checked %v, want %v", checked, user.ID)
		}
	})
}

func TestGetUserFromContext(t *testing.T) {
	newCtx := func() *gin.Context {
		c, _ := gin.CreateTestContext(nil)
		return c
	}

	t.Run("missing key", func(t *testing.T) {
		if _, err := GetUserFromContext(newCtx()); err == nil {
			t.Error("want error for missing user")
		}
	})

	t.Run("pointer value", func(t *testing.T) {
		c := newCtx()
		u := testUser()
		c.Set(ContextUserDetails, u)
		got, err := GetUserFromContext(c)
		if err != nil || got != u {
			t.Errorf("got (%v, %v)", got, err)
		}
	})

	t.Run("value type also accepted", func(t *testing.T) {
		c := newCtx()
		u := testUser()
		c.Set(ContextUserDetails, *u)
		got, err := GetUserFromContext(c)
		if err != nil || got.ID != u.ID {
			t.Errorf("got (%v, %v)", got, err)
		}
	})

	t.Run("nil pointer rejected", func(t *testing.T) {
		c := newCtx()
		var u *model.AuthUser
		c.Set(ContextUserDetails, u)
		if _, err := GetUserFromContext(c); err == nil {
			t.Error("want error for nil user pointer")
		}
	})

	t.Run("wrong type rejected", func(t *testing.T) {
		c := newCtx()
		c.Set(ContextUserDetails, "not-a-user")
		if _, err := GetUserFromContext(c); err == nil {
			t.Error("want error for foreign type")
		}
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	if _, err := GetUserIDFromContext(c); err == nil {
		t.Error("want error without a user")
	}

	u := testUser()
	c.Set(ContextUserDetails, u)
	id, err := GetUserIDFromContext(c)
	if err != nil || id != u.ID {
		t.Errorf("got (%v, %v), want (%v, nil)", id, err, u.ID)
	}
}
