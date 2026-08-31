package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"lexi/books/utils"

	"github.com/golang-jwt/jwt/v5"
)

// NewKeys builds a utils.Keys around a freshly generated RSA key pair's
// public half. It also returns the private half so tests can hand-sign
// tokens via SignToken: this service only verifies tokens issued elsewhere,
// it never generates them.
func NewKeys(t *testing.T) (utils.Keys, *rsa.PrivateKey) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	return NewKeysFromRSA(t, priv), priv
}

// NewKeysFromRSA builds a utils.Keys around a caller-supplied RSA key pair's
// public half.
func NewKeysFromRSA(t *testing.T, priv *rsa.PrivateKey) utils.Keys {
	t.Helper()

	dir := t.TempDir()
	pubPath := filepath.Join(dir, "public.pem")

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile(pubPath, pubPEM, 0o600); err != nil {
		t.Fatalf("failed to write public key: %v", err)
	}

	keys := utils.Keys{}
	if err := keys.LoadPublicKey(pubPath); err != nil {
		t.Fatalf("failed to load generated public key: %v", err)
	}

	return keys
}

// SignToken hand-crafts and signs a token with the given claims, standing in
// for the external service that actually issues tokens.
func SignToken(t *testing.T, priv *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}
