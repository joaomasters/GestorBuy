package auth

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Testes puros (sem MongoDB) para a lógica de token — a parte de
// hash/verificação de senha e emissão/parse de JWT não depende de I/O e deve
// ser coberta sem precisar de um banco no ar.

func TestIssueAndParseToken_RoundTrip(t *testing.T) {
	svc := &Service{jwtKey: []byte("test-secret"), jwtTTL: time.Hour}
	u := User{ID: "user-1", TenantID: "tenant-1", Role: RoleAdmin}

	token, err := svc.issueToken(u)
	if err != nil {
		t.Fatalf("issueToken falhou: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken falhou: %v", err)
	}
	if claims.TenantID != u.TenantID {
		t.Errorf("tenant_id = %q, esperado %q", claims.TenantID, u.TenantID)
	}
	if claims.UserID != u.ID {
		t.Errorf("user_id = %q, esperado %q", claims.UserID, u.ID)
	}
	if claims.Role != RoleAdmin {
		t.Errorf("role = %q, esperado %q", claims.Role, RoleAdmin)
	}
}

func TestParseToken_RejectsWrongSecret(t *testing.T) {
	issuer := &Service{jwtKey: []byte("secret-a"), jwtTTL: time.Hour}
	verifier := &Service{jwtKey: []byte("secret-b"), jwtTTL: time.Hour}

	token, err := issuer.issueToken(User{ID: "u1", TenantID: "t1"})
	if err != nil {
		t.Fatalf("issueToken falhou: %v", err)
	}

	if _, err := verifier.ParseToken(token); err == nil {
		t.Fatal("esperava erro ao validar token assinado com outra chave, mas passou")
	}
}

func TestParseToken_RejectsExpired(t *testing.T) {
	svc := &Service{jwtKey: []byte("test-secret"), jwtTTL: -time.Hour} // já expira ao emitir

	token, err := svc.issueToken(User{ID: "u1", TenantID: "t1"})
	if err != nil {
		t.Fatalf("issueToken falhou: %v", err)
	}

	if _, err := svc.ParseToken(token); err == nil {
		t.Fatal("esperava erro de token expirado, mas passou")
	}
}

func TestPasswordHash_RoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("senha-forte-123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword falhou: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("senha-forte-123")); err != nil {
		t.Errorf("senha correta deveria validar, mas falhou: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("senha-errada")); err == nil {
		t.Error("senha incorreta deveria falhar na validação, mas passou")
	}
}
