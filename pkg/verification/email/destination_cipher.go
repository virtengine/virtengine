package email

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

type destinationCipher struct {
	key []byte
}

func newDestinationCipher(encodedKey string) (*destinationCipher, error) {
	key, err := decodeDestinationKey(encodedKey)
	if err != nil {
		return nil, err
	}

	return &destinationCipher{key: key}, nil
}

func decodeDestinationKey(encodedKey string) ([]byte, error) {
	trimmed := strings.TrimSpace(encodedKey)
	if trimmed == "" {
		return nil, fmt.Errorf("destination encryption key is empty")
	}

	key, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("decode destination encryption key: %w", err)
	}

	switch len(key) {
	case 16, 24, 32:
		return key, nil
	default:
		return nil, fmt.Errorf("destination encryption key must decode to 16, 24, or 32 bytes")
	}
}

func (c *destinationCipher) Encrypt(value string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("create destination cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create destination gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate destination nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *destinationCipher) Decrypt(ciphertext string) (string, error) {
	encoded := strings.TrimSpace(ciphertext)
	if encoded == "" {
		return "", fmt.Errorf("encrypted destination is empty")
	}

	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode encrypted destination: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("create destination cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create destination gcm: %w", err)
	}

	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted destination payload is too short")
	}

	nonce := payload[:gcm.NonceSize()]
	data := payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt destination: %w", err)
	}

	return string(plaintext), nil
}
