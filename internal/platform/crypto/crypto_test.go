package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func testKey(t *testing.T) string {
	t.Helper()
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		t.Fatalf("falha ao gerar chave de teste: %v", err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	c, err := New(testKey(t))
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}

	plaintext := "APP_USR-token-secreto-do-mercado-livre"
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt falhou: %v", err)
	}
	if ciphertext == plaintext {
		t.Fatal("ciphertext igual ao plaintext — não cifrou nada")
	}

	got, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt falhou: %v", err)
	}
	if got != plaintext {
		t.Fatalf("Decrypt = %q, esperado %q", got, plaintext)
	}
}

func TestDecrypt_RejectsTamperedCiphertext(t *testing.T) {
	c, err := New(testKey(t))
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}

	ciphertext, err := c.Encrypt("segredo")
	if err != nil {
		t.Fatalf("Encrypt falhou: %v", err)
	}

	raw, _ := base64.StdEncoding.DecodeString(ciphertext)
	raw[len(raw)-1] ^= 0xFF // flip o último byte
	tampered := base64.StdEncoding.EncodeToString(raw)

	if _, err := c.Decrypt(tampered); err == nil {
		t.Fatal("esperava erro ao decifrar ciphertext adulterado, mas passou")
	}
}

func TestDecrypt_RejectsWrongKey(t *testing.T) {
	encryptor, err := New(testKey(t))
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}
	decryptor, err := New(testKey(t))
	if err != nil {
		t.Fatalf("New falhou: %v", err)
	}

	ciphertext, err := encryptor.Encrypt("segredo")
	if err != nil {
		t.Fatalf("Encrypt falhou: %v", err)
	}

	if _, err := decryptor.Decrypt(ciphertext); err == nil {
		t.Fatal("esperava erro ao decifrar com chave errada, mas passou")
	}
}

func TestNew_RejectsWrongKeySize(t *testing.T) {
	shortKey := base64.StdEncoding.EncodeToString([]byte("muito curta"))
	if _, err := New(shortKey); err == nil {
		t.Fatal("esperava erro pra chave de tamanho errado, mas passou")
	}
}
