package auth

import (
	"testing"

	"isletalk/identity/domain/valueobject"
)

func TestMTLSServerRequiresEnv(t *testing.T) {
	t.Setenv("IDENTITY_TLS_CERT", "")
	t.Setenv("IDENTITY_TLS_KEY", "")
	t.Setenv("IDENTITY_TLS_CA", "")
	if _, err := ServerTLS(); err == nil {
		t.Fatal("ServerTLS should error without cert env")
	}
	if _, err := ClientTLS("identity-svc"); err == nil {
		t.Fatal("ClientTLS should error without cert env")
	}
}

func TestTokenRevoke(t *testing.T) {
	issuer := NewTokenIssuer()
	uid := mustUID(t)
	if err := issuer.Revoke(nil, uid); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}

func mustUID(t *testing.T) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId("u-1")
	if err != nil {
		t.Fatal(err)
	}
	return v
}
