package validation

import "testing"

func TestRegisterRequestValidation(t *testing.T) {
	if err := RegisterRequest("RetailPulse", "Owner", "owner@example.com", "long-secure-password", "seller"); err != nil {
		t.Fatalf("expected valid request: %v", err)
	}
	if err := RegisterRequest("", "Owner", "owner@example.com", "long-secure-password", "owner"); err == nil {
		t.Fatal("expected missing organization error")
	}
	if err := RegisterRequest("RetailPulse", "Owner", "bad-email", "long-secure-password", "owner"); err == nil {
		t.Fatal("expected invalid email error")
	}
	if err := RegisterRequest("RetailPulse", "Owner", "owner@example.com", "short", "owner"); err == nil {
		t.Fatal("expected short password error")
	}
	if err := RegisterRequest("RetailPulse", "Owner", "owner@example.com", "long-secure-password", "partner"); err == nil {
		t.Fatal("expected account type error")
	}
}
