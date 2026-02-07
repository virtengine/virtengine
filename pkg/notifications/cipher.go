package notifications

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// TokenCipher encrypts and decrypts device tokens at rest.
type TokenCipher interface {
	Encrypt(plain string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// NoopCipher performs no encryption.
type NoopCipher struct{}

func (NoopCipher) Encrypt(plain string) (string, error)      { return plain, nil }
func (NoopCipher) Decrypt(ciphertext string) (string, error) { return ciphertext, nil }

// AESGCMCipher encrypts tokens using AES-GCM.
type AESGCMCipher struct {
	key []byte
}

// NewAESGCMCipher builds an AES-GCM cipher with a 16, 24, or 32 byte key.
func NewAESGCMCipher(key []byte) (*AESGCMCipher, error) {
	switch len(key) {
	case 16, 24, 32:
		return &AESGCMCipher{key: key}, nil
	default:
		return nil, fmt.Errorf("invalid AES key length: %d", len(key))
	}
}

func (c *AESGCMCipher) Encrypt(plain string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plain), nil)
	payload := make([]byte, 0, len(nonce)+len(ciphertext))
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func (c *AESGCMCipher) Decrypt(ciphertext string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}
	nonce := payload[:nonceSize]
	enc := payload[nonceSize:]
	plain, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
