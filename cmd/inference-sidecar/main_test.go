// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadServerTLSConfigRejectsPartialConfiguration(t *testing.T) {
	t.Parallel()

	_, err := loadServerTLSConfig("server.crt", "", "", false)
	if err == nil || !strings.Contains(err.Error(), "tls-cert-file and tls-key-file") {
		t.Fatalf("expected incomplete TLS configuration error, got %v", err)
	}
}

func TestLoadServerTLSConfigRequiresVerifiedClientCertificates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	certPEM, keyPEM := testCertificate(t)
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")
	caPath := filepath.Join(dir, "ca.crt")
	for path, contents := range map[string][]byte{certPath: certPEM, keyPath: keyPEM, caPath: certPEM} {
		if err := os.WriteFile(path, contents, 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	config, err := loadServerTLSConfig(certPath, keyPath, caPath, true)
	if err != nil {
		t.Fatalf("load mTLS configuration: %v", err)
	}
	if config.MinVersion != tls.VersionTLS13 || config.ClientAuth != tls.RequireAndVerifyClientCert || config.ClientCAs == nil {
		t.Fatalf("expected TLS 1.3 verified client certificate configuration, got %#v", config)
	}

	if _, err := loadServerTLSConfig(certPath, keyPath, caPath, false); err == nil || !strings.Contains(err.Error(), "tls-client-ca-file requires") {
		t.Fatalf("expected client CA without mTLS to be rejected, got %v", err)
	}
}

func testCertificate(t *testing.T) ([]byte, []byte) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("generate serial: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "veid-inference-test"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
}
