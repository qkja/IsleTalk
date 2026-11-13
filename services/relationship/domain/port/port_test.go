package port

import "testing"

func TestErrNotFound(t *testing.T) {
	if ErrNotFound.Error() != "entity not found" {
		t.Fatalf("ErrNotFound.Error() = %q", ErrNotFound.Error())
	}
	if !IsNotFound(ErrNotFound) {
		t.Fatal("IsNotFound(ErrNotFound) should be true")
	}
}
