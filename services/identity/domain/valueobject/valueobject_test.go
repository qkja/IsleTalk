package valueobject

import "testing"

func TestNewUserId(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"non-empty", "0190a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserId(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewUserId(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got.String() != tt.in {
				t.Fatalf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}

func TestNewPublicName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"valid", "alice", false},
		{"empty", "", true},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true}, // 33 chars
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPublicName(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPublicName(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestNewPhoneHash(t *testing.T) {
	if _, err := NewPhoneHash(""); err == nil {
		t.Fatal("expected error for empty phone hash")
	}
	if h, err := NewPhoneHash("hash123"); err != nil || h.String() != "hash123" {
		t.Fatalf("unexpected result: %v %v", h, err)
	}
}

func TestNewPasswordHash(t *testing.T) {
	if _, err := NewPasswordHash(""); err == nil {
		t.Fatal("expected error for empty password hash")
	}
}

func TestUserIdEqual(t *testing.T) {
	a, _ := NewUserId("id-1")
	b, _ := NewUserId("id-1")
	c, _ := NewUserId("id-2")
	if !a.Equal(b) {
		t.Fatal("same value should be equal")
	}
	if a.Equal(c) {
		t.Fatal("different value should not be equal")
	}
}

func TestNewPublicNameReserved(t *testing.T) {
	for _, w := range []string{"admin", "Admin", "system", "isletalk", "官方", "管理员"} {
		if _, err := NewPublicName(w); err != ErrReservedPublicName {
			t.Fatalf("NewPublicName(%q) err = %v, want ErrReservedPublicName", w, err)
		}
	}
	// 非保留词正常构造
	if _, err := NewPublicName("alice"); err != nil {
		t.Fatalf("NewPublicName(alice) err = %v", err)
	}
}
