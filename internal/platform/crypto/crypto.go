// Package crypto cifra campos sensíveis (hoje: tokens OAuth de marketplace)
// antes de persistir no MongoDB.
//
// Isso é uma simplificação deliberada em relação à seção 3.1 do doc de
// arquitetura, que previa Client-Side Field Level Encryption do Atlas com
// um KMS externo (AWS/GCP KMS). Provisionar um KMS está fora do escopo
// desta entrega — em vez disso, ciframos com uma chave mestra única lida de
// variável de ambiente (AES-256-GCM). Resolve o requisito real ("nunca
// token em texto puro no banco") sem infraestrutura nova. Migrar pra
// CSFLE/KMS é um item de Fase 2, não uma reescrita: só troca a
// implementação por trás de Encrypt/Decrypt.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

var ErrInvalidKey = errors.New("crypto: chave precisa ter 32 bytes (AES-256) após decodificar de base64")

// Cipher encapsula a chave mestra — instanciado uma vez no boot da aplicação
// a partir de TOKEN_ENCRYPTION_KEY e passado por injeção pros serviços que
// precisam cifrar/decifrar (ex.: internal/marketplace).
type Cipher struct {
	gcm cipher.AEAD
}

// New recebe a chave em base64 (32 bytes decodificados = AES-256).
func New(base64Key string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("crypto: chave inválida (base64): %w", err)
	}
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: erro ao criar cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: erro ao criar GCM: %w", err)
	}

	return &Cipher{gcm: gcm}, nil
}

// Encrypt cifra plaintext e devolve nonce+ciphertext codificados em base64
// — uma string só, pronta pra gravar num campo bson.
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("crypto: falha ao gerar nonce: %w", err)
	}

	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverte Encrypt. Retorna erro se o ciphertext foi adulterado
// (GCM autentica o conteúdo) ou está malformado.
func (c *Cipher) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto: ciphertext inválido (base64): %w", err)
	}

	nonceSize := c.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("crypto: ciphertext menor que o nonce esperado")
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: falha ao decifrar (adulterado ou chave errada): %w", err)
	}

	return string(plaintext), nil
}
