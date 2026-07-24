package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestNewPKCE(t *testing.T) {
	verifier, challenge, err := newPKCE()
	if err != nil {
		t.Fatalf("newPKCE: %v", err)
	}
	if verifier == "" || challenge == "" {
		t.Fatalf("empty verifier/challenge: %q %q", verifier, challenge)
	}
	if _, err := base64.RawURLEncoding.DecodeString(verifier); err != nil {
		t.Fatalf("verifier is not raw base64url: %v", err)
	}
	sum := sha256.Sum256([]byte(verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if challenge != want {
		t.Fatalf("challenge = %q, want S256(verifier) = %q", challenge, want)
	}

	verifier2, _, _ := newPKCE()
	if verifier == verifier2 {
		t.Fatal("verifier should be random, got the same value twice")
	}
}

func TestNewState(t *testing.T) {
	s1, err := newState()
	if err != nil {
		t.Fatalf("newState: %v", err)
	}
	if s1 == "" {
		t.Fatal("state is empty")
	}
	if s2, _ := newState(); s1 == s2 {
		t.Fatal("state should be random, got the same value twice")
	}
}
