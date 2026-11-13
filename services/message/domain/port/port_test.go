package port

import "testing"

func TestErrNotFoundAndIsNotFound(t *testing.T) {
	err := ErrNotFound
	if !IsNotFound(err) {
		t.Fatal("IsNotFound should be true for ErrNotFound")
	}
	if IsNotFound(nil) {
		t.Fatal("IsNotFound(nil) should be false")
	}
	if err.Error() == "" {
		t.Fatal("ErrNotFound.Error() should not be empty")
	}
}
