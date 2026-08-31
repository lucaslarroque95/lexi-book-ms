package utils_test

import (
	"testing"
	"time"

	"lexi/books/test/testutil"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyAndExtractClaims_RoundTripsMultipleRoles(t *testing.T) {
	keys, priv := testutil.NewKeys(t)

	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email":  "admin@lexi.com",
		"userId": "7",
		"role":   "admin,default",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})

	if err := keys.VerifyToken(token); err != nil {
		t.Fatalf("VerifyToken returned error: %v", err)
	}

	userId, roles, err := keys.ExtractClaims(token)
	if err != nil {
		t.Fatalf("ExtractClaims returned error: %v", err)
	}
	if userId != "7" {
		t.Fatalf("expected userId 7, got %q", userId)
	}
	if len(roles) != 2 || roles[0] != "admin" || roles[1] != "default" {
		t.Fatalf("expected roles [admin default], got %v", roles)
	}
}

func TestVerifyAndExtractClaims_NoRoles(t *testing.T) {
	keys, priv := testutil.NewKeys(t)

	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email":  "user@lexi.com",
		"userId": "1",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})

	if err := keys.VerifyToken(token); err != nil {
		t.Fatalf("VerifyToken returned error: %v", err)
	}

	_, roles, err := keys.ExtractClaims(token)
	if err != nil {
		t.Fatalf("ExtractClaims returned error: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected no roles, got %v", roles)
	}
}

func TestVerifyToken_MalformedToken(t *testing.T) {
	keys, _ := testutil.NewKeys(t)

	if err := keys.VerifyToken("not-a-real-token"); err == nil {
		t.Fatalf("expected error for a malformed token")
	}
}

func TestExtractClaims_MalformedToken(t *testing.T) {
	keys, _ := testutil.NewKeys(t)

	if _, _, err := keys.ExtractClaims("not-a-real-token"); err == nil {
		t.Fatalf("expected error for a malformed token")
	}
}

func TestVerifyToken_WrongSigningKey(t *testing.T) {
	_, signingPriv := testutil.NewKeys(t)
	verifyingKeys, _ := testutil.NewKeys(t)

	token := testutil.SignToken(t, signingPriv, jwt.MapClaims{
		"userId": "1",
		"role":   "admin",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})

	if err := verifyingKeys.VerifyToken(token); err == nil {
		t.Fatalf("expected error when verifying a token signed by a different key")
	}
}

func TestExtractClaims_IgnoresSignature(t *testing.T) {
	_, signingPriv := testutil.NewKeys(t)
	verifyingKeys, _ := testutil.NewKeys(t)

	token := testutil.SignToken(t, signingPriv, jwt.MapClaims{
		"userId": "1",
		"role":   "admin",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})

	userId, roles, err := verifyingKeys.ExtractClaims(token)
	if err != nil {
		t.Fatalf("ExtractClaims returned error: %v", err)
	}
	if userId != "1" || len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("expected userId 1 and roles [admin], got userId=%q roles=%v", userId, roles)
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	keys, priv := testutil.NewKeys(t)

	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email":  "user@lexi.com",
		"userId": "1",
		"role":   "admin",
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})

	if err := keys.VerifyToken(token); err == nil {
		t.Fatalf("expected error when verifying an expired token")
	}
}

func TestExtractClaims_SucceedsForExpiredToken(t *testing.T) {
	keys, priv := testutil.NewKeys(t)

	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email":  "user@lexi.com",
		"userId": "9",
		"role":   "admin",
		"exp":    time.Now().Add(-time.Hour).Unix(),
	})

	userId, roles, err := keys.ExtractClaims(token)
	if err != nil {
		t.Fatalf("ExtractClaims returned error: %v", err)
	}
	if userId != "9" || len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("expected userId 9 and roles [admin], got userId=%q roles=%v", userId, roles)
	}
}

func TestExtractClaims_MissingUserIdClaim(t *testing.T) {
	keys, priv := testutil.NewKeys(t)

	token := testutil.SignToken(t, priv, jwt.MapClaims{
		"email": "user@lexi.com",
		"role":  "admin",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	if _, _, err := keys.ExtractClaims(token); err == nil {
		t.Fatalf("expected error for a token without a userId claim")
	}
}
