package participant

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Identity struct {
	ID         string `json:"id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func LoadOrCreateIdentity(path string, id string) (*Identity, error) {
	if path == "" {
		return nil, fmt.Errorf("identity path required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var identity Identity
		if err := json.Unmarshal(data, &identity); err != nil {
			return nil, err
		}
		return &identity, nil
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	identity := &Identity{
		ID:         id,
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
	}
	if err := identity.Save(path); err != nil {
		return nil, err
	}
	return identity, nil
}

func NewDeterministicIdentity(id, seedMaterial string) *Identity {
	seed := sha256.Sum256([]byte(seedMaterial))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return &Identity{
		ID:         id,
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
	}
}

func (i *Identity) Save(path string) error {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (i *Identity) Sign(data []byte) (string, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(i.PrivateKey)
	if err != nil {
		return "", err
	}
	signature := ed25519.Sign(privateKeyBytes, data)
	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifySignature(publicKey string, data []byte, signature string) error {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(publicKeyBytes, data, signatureBytes) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}
