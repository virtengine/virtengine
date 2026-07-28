// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestServerMTLSRequiresVerifiedClientCertificate(t *testing.T) {
	material := generateMTLSTestMaterial(t)
	serverTLS, err := buildServerTLSConfig(sidecarTLSOptions{
		RequireMTLS:  true,
		CertFile:     material.serverCert,
		KeyFile:      material.serverKey,
		ClientCAFile: material.clientCA,
	})
	if err != nil {
		t.Fatalf("build server tls config: %v", err)
	}
	if serverTLS.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected TLS 1.3 minimum, got %x", serverTLS.MinVersion)
	}
	if serverTLS.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("expected RequireAndVerifyClientCert, got %v", serverTLS.ClientAuth)
	}

	addr := startMTLSHealthServer(t, serverTLS)

	t.Run("valid client succeeds", func(t *testing.T) {
		conn := dialHealth(t, addr, validClientTLSConfig(t, material))
		defer conn.Close()
		resp, err := grpc_health_v1.NewHealthClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			t.Fatalf("health check with valid client cert failed: %v", err)
		}
		if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Fatalf("expected SERVING, got %s", resp.Status.String())
		}
	})

	t.Run("no client cert fails", func(t *testing.T) {
		cfg := validClientTLSConfig(t, material)
		cfg.Certificates = nil
		assertHealthFails(t, addr, credentials.NewTLS(cfg), "")
	})

	t.Run("wrong client ca fails", func(t *testing.T) {
		cfg := validClientTLSConfig(t, material)
		cert, err := tls.LoadX509KeyPair(material.wrongClientCert, material.wrongClientKey)
		if err != nil {
			t.Fatalf("load wrong client cert: %v", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
		assertHealthFails(t, addr, credentials.NewTLS(cfg), "")
	})

	t.Run("plaintext fails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			defer conn.Close()
			_, err = grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		}
		if err == nil {
			t.Fatal("expected plaintext connection to fail")
		}
	})
}

func TestServerMTLSFileValidationFailsClosed(t *testing.T) {
	material := generateMTLSTestMaterial(t)
	base := sidecarTLSOptions{
		RequireMTLS:  true,
		CertFile:     material.serverCert,
		KeyFile:      material.serverKey,
		ClientCAFile: material.clientCA,
	}

	dir := t.TempDir()
	malformed := filepath.Join(dir, "malformed.pem")
	if err := os.WriteFile(malformed, []byte("not a pem block"), 0o600); err != nil {
		t.Fatalf("write malformed pem: %v", err)
	}
	trailingMalformed := filepath.Join(dir, "trailing-malformed.pem")
	validCA, err := os.ReadFile(material.clientCA)
	if err != nil {
		t.Fatalf("read valid client ca: %v", err)
	}
	if err := os.WriteFile(trailingMalformed, append(validCA, []byte("\nnot pem trailing data")...), 0o600); err != nil {
		t.Fatalf("write trailing malformed pem: %v", err)
	}
	missing := filepath.Join(dir, "missing.pem")

	tests := []struct {
		name   string
		mutate func(*sidecarTLSOptions)
		want   string
	}{
		{
			name: "empty cert",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = ""
			},
			want: "tls-cert-file",
		},
		{
			name: "empty key",
			mutate: func(opts *sidecarTLSOptions) {
				opts.KeyFile = ""
			},
			want: "tls-key-file",
		},
		{
			name: "empty client ca",
			mutate: func(opts *sidecarTLSOptions) {
				opts.ClientCAFile = ""
			},
			want: "tls-client-ca-file",
		},
		{
			name: "relative cert",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = "relative.pem"
			},
			want: "absolute",
		},
		{
			name: "missing cert",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = missing
			},
			want: "read",
		},
		{
			name: "directory cert",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = dir
			},
			want: "regular file",
		},
		{
			name: "malformed cert",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = malformed
			},
			want: "certificate",
		},
		{
			name: "client certificate used as server certificate",
			mutate: func(opts *sidecarTLSOptions) {
				opts.CertFile = material.clientCert
				opts.KeyFile = material.clientKey
			},
			want: "server authentication",
		},
		{
			name: "malformed client ca",
			mutate: func(opts *sidecarTLSOptions) {
				opts.ClientCAFile = malformed
			},
			want: "client ca",
		},
		{
			name: "trailing malformed client ca",
			mutate: func(opts *sidecarTLSOptions) {
				opts.ClientCAFile = trailingMalformed
			},
			want: "trailing",
		},
		{
			name: "leaf certificate as client ca",
			mutate: func(opts *sidecarTLSOptions) {
				opts.ClientCAFile = material.clientCert
			},
			want: "certificate signing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := base
			tt.mutate(&opts)
			_, err := buildServerTLSConfig(opts)
			if err == nil {
				t.Fatal("expected TLS validation to fail")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.want)) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestExplicitDevelopmentPlaintextOptOutServesPlaintext(t *testing.T) {
	tlsConfig, err := buildServerTLSConfig(sidecarTLSOptions{RequireMTLS: false})
	if err != nil {
		t.Fatalf("build development plaintext config: %v", err)
	}
	if tlsConfig != nil {
		t.Fatal("expected no TLS config for explicit development plaintext opt-out")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	t.Cleanup(func() {
		grpcServer.GracefulStop()
		_ = listener.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial explicit development plaintext server: %v", err)
	}
	defer conn.Close()

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("development plaintext health check: %v", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %s", resp.Status.String())
	}

	opts := defaultSidecarOptions()
	opts.RequireMTLS = false
	if config := buildInferenceConfig(opts); config.UseFallbackOnError || config.AllowFallbackToStub {
		t.Fatal("development plaintext opt-out must not enable fallback scoring")
	}
}

func TestStartupOptionsFailClosedDefaultsAndFallbackPolicy(t *testing.T) {
	defaults := defaultSidecarOptions()
	if !defaults.RequireMTLS {
		t.Fatal("expected require-mtls to default true")
	}
	if defaults.MetricsAddr != ":9092" {
		t.Fatalf("expected metrics default :9092, got %q", defaults.MetricsAddr)
	}

	dev := defaults
	dev.RequireMTLS = false
	dev.AllowFallbackToStub = false
	if err := validateStartupOptions(dev); err != nil {
		t.Fatalf("explicit plaintext opt-out should be allowed for development without fallback: %v", err)
	}
	cfg := buildInferenceConfig(dev)
	if cfg.UseFallbackOnError {
		t.Fatal("sidecar inference config must keep UseFallbackOnError disabled")
	}
	if cfg.AllowFallbackToStub {
		t.Fatal("plaintext opt-out must not automatically enable stub fallback")
	}

	material := generateMTLSTestMaterial(t)
	prod := defaults
	prod.TLSCertFile = material.serverCert
	prod.TLSKeyFile = material.serverKey
	prod.TLSClientCAFile = material.clientCA
	prod.AllowFallbackToStub = true
	if err := validateStartupOptions(prod); err == nil || !strings.Contains(err.Error(), "--allow-fallback-to-stub") {
		t.Fatalf("expected mTLS plus stub fallback to be rejected, got %v", err)
	}

	productionCases := []struct {
		name   string
		mutate func(*sidecarOptions)
		want   string
	}{
		{
			name: "missing serving url",
			mutate: func(opts *sidecarOptions) {
				opts.ServingURL = ""
			},
			want: "--serving-url",
		},
		{
			name: "serving fallback url",
			mutate: func(opts *sidecarOptions) {
				opts.ServingFallbackURL = "https://fallback.invalid"
			},
			want: "--serving-fallback-url",
		},
		{
			name: "non cpu",
			mutate: func(opts *sidecarOptions) {
				opts.ForceCPU = false
			},
			want: "--force-cpu=true",
		},
		{
			name: "wrong seed",
			mutate: func(opts *sidecarOptions) {
				opts.RandomSeed = 43
			},
			want: "--random-seed=42",
		},
	}
	for _, tt := range productionCases {
		t.Run(tt.name, func(t *testing.T) {
			opts := prod
			opts.AllowFallbackToStub = false
			tt.mutate(&opts)
			err := validateStartupOptions(opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected startup error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestRuntimeReadinessFailsWhenVerifiedBundleHasNoServingBackend(t *testing.T) {
	server := &InferenceSidecarServer{
		verification: &verificationResult{State: verificationStateVerified},
		log:          noopLogger{},
	}
	statusCode, payload := runtimeReadinessHTTPResponse(context.Background(), server)
	if statusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected runtime readiness 503, got %d: %s", statusCode, payload)
	}
	if !strings.Contains(string(payload), "runtime_unavailable") {
		t.Fatalf("expected runtime_unavailable payload, got %s", payload)
	}
}

type mtlsTestMaterial struct {
	clientCA        string
	serverCert      string
	serverKey       string
	clientCert      string
	clientKey       string
	wrongClientCert string
	wrongClientKey  string
}

func generateMTLSTestMaterial(t *testing.T) mtlsTestMaterial {
	t.Helper()
	dir := t.TempDir()

	clientCA, clientCAKey := newTestCA(t, "client-ca")
	wrongCA, wrongCAKey := newTestCA(t, "wrong-client-ca")
	serverCA, serverCAKey := newTestCA(t, "server-ca")

	material := mtlsTestMaterial{
		clientCA:        filepath.Join(dir, "client-ca.pem"),
		serverCert:      filepath.Join(dir, "server.pem"),
		serverKey:       filepath.Join(dir, "server-key.pem"),
		clientCert:      filepath.Join(dir, "client.pem"),
		clientKey:       filepath.Join(dir, "client-key.pem"),
		wrongClientCert: filepath.Join(dir, "wrong-client.pem"),
		wrongClientKey:  filepath.Join(dir, "wrong-client-key.pem"),
	}
	writeCertPEM(t, material.clientCA, clientCA.Raw)
	writeLeafCert(t, material.serverCert, material.serverKey, "server", x509.ExtKeyUsageServerAuth, serverCA, serverCAKey)
	writeLeafCert(t, material.clientCert, material.clientKey, "client", x509.ExtKeyUsageClientAuth, clientCA, clientCAKey)
	writeLeafCert(t, material.wrongClientCert, material.wrongClientKey, "wrong-client", x509.ExtKeyUsageClientAuth, wrongCA, wrongCAKey)
	return material
}

func startMTLSHealthServer(t *testing.T, cfg *tls.Config) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	t.Cleanup(func() {
		grpcServer.GracefulStop()
		_ = listener.Close()
	})
	return listener.Addr().String()
}

func dialHealth(t *testing.T, addr string, cfg *tls.Config) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(credentials.NewTLS(cfg)))
	if err != nil {
		t.Fatalf("dial health server: %v", err)
	}
	return conn
}

func assertHealthFails(t *testing.T, addr string, creds credentials.TransportCredentials, want string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err == nil {
		defer conn.Close()
		_, err = grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	}
	if err == nil {
		t.Fatalf("expected health check to fail with %s", want)
	}
	if want != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func validClientTLSConfig(t *testing.T, material mtlsTestMaterial) *tls.Config {
	t.Helper()
	cert, err := tls.LoadX509KeyPair(material.clientCert, material.clientKey)
	if err != nil {
		t.Fatalf("load client cert: %v", err)
	}
	caPEM, err := os.ReadFile(material.serverCert)
	if err != nil {
		t.Fatalf("read server cert: %v", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatal("append server cert root")
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		ServerName:   "localhost",
		RootCAs:      roots,
		Certificates: []tls.Certificate{cert},
	}
}

func newTestCA(t *testing.T, cn string) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate ca key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create ca cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse ca cert: %v", err)
	}
	return cert, key
}

func writeLeafCert(t *testing.T, certPath, keyPath, cn string, usage x509.ExtKeyUsage, ca *x509.Certificate, caKey *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate leaf key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create leaf cert: %v", err)
	}
	writeCertPEM(t, certPath, der)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("write key %s: %v", keyPath, err)
	}
}

func writeCertPEM(t *testing.T, path string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatalf("write cert %s: %v", path, err)
	}
}
