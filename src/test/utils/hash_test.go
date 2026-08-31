package utils_test

import (
	"testing"

	"lexi/books/utils"
)

func TestHashPassword_CheckPasswordHash_RoundTrip(t *testing.T) {
	hashed, err := utils.HashPassword("s3cr3t")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hashed == "s3cr3t" {
		t.Fatalf("expected hashed password to differ from plaintext")
	}
	if !utils.CheckPasswordHash("s3cr3t", hashed) {
		t.Fatalf("expected CheckPasswordHash to succeed with the original password")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hashed, err := utils.HashPassword("s3cr3t")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if utils.CheckPasswordHash("wrong-password", hashed) {
		t.Fatalf("expected CheckPasswordHash to fail with the wrong password")
	}
}
