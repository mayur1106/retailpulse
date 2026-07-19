package security

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"retailpulse/apps/api/internal/domain"
)

func TestTokenManagerCreatesAndParsesAccessToken(t *testing.T) {
	manager := NewTokenManager("access-secret-with-at-least-32-bytes", "refresh-secret-with-at-least-32-bytes")
	principal := domain.Principal{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Role:           domain.RoleOwner,
		Email:          "owner@example.com",
	}
	token, _, err := manager.CreateAccessToken(principal, time.Minute)
	if err != nil {
		t.Fatalf("create access token: %v", err)
	}
	parsed, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if parsed.UserID != principal.UserID || parsed.OrganizationID != principal.OrganizationID || parsed.Role != principal.Role || parsed.Email != principal.Email {
		t.Fatalf("principal mismatch: %#v", parsed)
	}
}

func TestTokenManagerRefreshHashIsStable(t *testing.T) {
	manager := NewTokenManager("access-secret-with-at-least-32-bytes", "refresh-secret-with-at-least-32-bytes")
	token, hash, err := manager.CreateRefreshToken()
	if err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	if token == "" || hash == "" {
		t.Fatal("expected token and hash")
	}
	if manager.HashRefreshToken(token) != hash {
		t.Fatal("expected stable refresh hash")
	}
}
