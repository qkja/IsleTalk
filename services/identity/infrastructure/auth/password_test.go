package auth

import "testing"

func TestPasswordHashVerify(t *testing.T) {
	h := NewPasswordHasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash err: %v", err)
	}

	if !h.Verify("correct horse battery staple", hash) {
		t.Fatal("Verify should return true for correct password")
	}
	if h.Verify("wrong password", hash) {
		t.Fatal("Verify should return false for wrong password")
	}
}

func TestPasswordHashUniqueSalt(t *testing.T) {
	h := NewPasswordHasher()
	a, _ := h.Hash("same-password")
	b, _ := h.Hash("same-password")
	if a.String() == b.String() {
		t.Fatal("same password should produce different hashes (unique salt)")
	}
}
