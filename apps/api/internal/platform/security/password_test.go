package security

import "testing"

func TestPasswordHasher(t *testing.T) {
	hasher := NewPasswordHasher()
	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !hasher.Compare(hash, "correct horse battery staple") {
		t.Fatal("expected password to match")
	}
	if hasher.Compare(hash, "wrong password") {
		t.Fatal("expected wrong password not to match")
	}
}
