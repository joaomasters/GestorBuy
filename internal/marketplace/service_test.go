package marketplace

import (
	"net/url"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	svc, err := New(nil, nil, "test-jwt-secret", "client-id", "client-secret", "https://example.com/callback", "wCldyXFi8ZT57aSflXlutSrDydwAo/gGsfAvNFb+vqs=")
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}
	return svc
}

func TestConnectURL_StateRoundTrip(t *testing.T) {
	svc := newTestService(t)

	authURL, err := svc.ConnectURL("tenant-123")
	if err != nil {
		t.Fatalf("ConnectURL falhou: %v", err)
	}

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("URL de autorização inválida: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatal("URL de autorização não tem state")
	}

	tenantID, err := svc.parseState(state)
	if err != nil {
		t.Fatalf("parseState falhou: %v", err)
	}
	if tenantID != "tenant-123" {
		t.Fatalf("tenantID = %q, esperado %q", tenantID, "tenant-123")
	}
}

func TestParseState_RejectsGarbage(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.parseState("isso-não-é-um-jwt"); err == nil {
		t.Fatal("esperava erro pra state inválido, mas passou")
	}
}

func TestParseState_RejectsStateSignedWithDifferentSecret(t *testing.T) {
	svc := newTestService(t)
	other, err := New(nil, nil, "outro-secret-completamente-diferente", "client-id", "client-secret", "https://example.com/callback", "wCldyXFi8ZT57aSflXlutSrDydwAo/gGsfAvNFb+vqs=")
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}

	authURL, err := other.ConnectURL("tenant-123")
	if err != nil {
		t.Fatalf("ConnectURL falhou: %v", err)
	}
	parsed, _ := url.Parse(authURL)
	state := parsed.Query().Get("state")

	if _, err := svc.parseState(state); err == nil {
		t.Fatal("esperava erro ao validar state assinado com outro secret, mas passou")
	}
}

func TestNew_NotConfiguredWithoutMLCredentials(t *testing.T) {
	svc, err := New(nil, nil, "test-jwt-secret", "", "", "", "")
	if err != nil {
		t.Fatalf("New não deveria falhar só por faltar credencial do ML: %v", err)
	}

	if _, err := svc.ConnectURL("tenant-123"); err != ErrNotConfigured {
		t.Fatalf("esperava ErrNotConfigured, veio %v", err)
	}
}
