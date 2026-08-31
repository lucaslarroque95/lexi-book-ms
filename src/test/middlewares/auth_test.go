package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lexi/books/middlewares"
	"lexi/books/test/testutil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthenticate_MissingHeaderReturns401(t *testing.T) {
	keys, _ := testutil.NewKeys(t)

	router := gin.New()
	router.GET("/protected", middlewares.Authenticate(keys), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_InvalidTokenReturns401(t *testing.T) {
	keys, _ := testutil.NewKeys(t)

	router := gin.New()
	router.GET("/protected", middlewares.Authenticate(keys), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "not-a-real-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_ValidTokenSetsContextAndCallsNext(t *testing.T) {
	keys, priv := testutil.NewKeys(t)
	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email":  "admin@lexi.com",
		"userId": "42",
		"role":   "admin,default",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})

	var gotUserId string
	var gotRoles []string

	router := gin.New()
	router.GET("/protected", middlewares.Authenticate(keys), func(c *gin.Context) {
		gotUserId = c.GetString("userId")
		rolesValue, _ := c.Get("roles")
		gotRoles, _ = rolesValue.([]string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotUserId != "42" {
		t.Fatalf("expected userId 42, got %q", gotUserId)
	}
	if len(gotRoles) != 2 || gotRoles[0] != "admin" || gotRoles[1] != "default" {
		t.Fatalf("expected roles [admin default], got %v", gotRoles)
	}
}

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	router := gin.New()
	router.GET("/admin",
		func(c *gin.Context) { c.Set("roles", []string{"admin"}); c.Next() },
		middlewares.RequireRole("admin"),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_BlocksNonMatchingRole(t *testing.T) {
	router := gin.New()
	router.GET("/admin",
		func(c *gin.Context) { c.Set("roles", []string{"default"}); c.Next() },
		middlewares.RequireRole("admin"),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRole_BlocksWhenRolesMissingFromContext(t *testing.T) {
	router := gin.New()
	router.GET("/admin", middlewares.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
