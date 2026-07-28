// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package inference

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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

func TestSidecarClientExplicitMTLSGetModelInfo(t *testing.T) {
	material := generateSidecarClientMTLSTestMaterial(t)
	addr := startInferenceMTLSTestServer(t, material)

	client, err := NewSidecarClient(validMTLSTestInferenceConfig(addr, material))
	if err != nil {
		t.Fatalf("create sidecar client with valid mTLS material: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if got := client.GetModelVersion(); got != "v1.2.3" {
		t.Fatalf("expected model version v1.2.3, got %q", got)
	}
	if got := client.GetModelHash(); got != strings.Repeat("a", 64) {
		t.Fatalf("expected model hash from server, got %q", got)
	}

	assertStandardGRPCHealthOverMTLS(t, addr, material)
}

func TestSidecarClientMTLSMaterialFailsClosed(t *testing.T) {
	material := generateSidecarClientMTLSTestMaterial(t)
	addr := startInferenceMTLSTestServer(t, material)

	dir := t.TempDir()
	malformedCA := filepath.Join(dir, "malformed-ca.pem")
	if err := os.WriteFile(malformedCA, []byte("not a pem bundle"), 0o600); err != nil {
		t.Fatalf("write malformed ca: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*InferenceConfig)
		want   string
	}{
		{
			name: "missing model version",
			mutate: func(config *InferenceConfig) {
				config.ModelVersion = ""
			},
			want: "model_version",
		},
		{
			name: "missing client key",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLSKeyFile = ""
			},
			want: "required",
		},
		{
			name: "relative client cert",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLSCertFile = "client.pem"
			},
			want: "absolute",
		},
		{
			name: "explicit credentials with plaintext transport",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLS = false
			},
			want: "TLS must be enabled",
		},
		{
			name: "malformed server ca",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLSServerCAFile = malformedCA
			},
			want: "server ca",
		},
		{
			name: "wrong client ca",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLSCertFile = material.wrongClientCert
				config.SidecarTLSKeyFile = material.wrongClientKey
			},
			want: "getmodelinfo",
		},
		{
			name: "no explicit client credentials",
			mutate: func(config *InferenceConfig) {
				config.SidecarTLSCertFile = ""
				config.SidecarTLSKeyFile = ""
				config.SidecarTLSServerCAFile = ""
			},
			want: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validMTLSTestInferenceConfig(addr, material)
			tt.mutate(&config)
			client, err := NewSidecarClient(config)
			if err == nil {
				_ = client.Close()
				t.Fatal("expected sidecar client creation to fail")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.want)) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestSidecarClientRejectsResponseModelIdentityDrift(t *testing.T) {
	material := generateSidecarClientMTLSTestMaterial(t)
	tests := []struct {
		name    string
		service sidecarClientModelInfoServer
		want    string
	}{
		{
			name: "model version",
			service: sidecarClientModelInfoServer{
				scoreModelVersion: "v9.9.9",
				scoreModelHash:    strings.Repeat("a", 64),
			},
			want: "model version mismatch",
		},
		{
			name: "model hash",
			service: sidecarClientModelInfoServer{
				scoreModelVersion: "v1.2.3",
				scoreModelHash:    strings.Repeat("b", 64),
			},
			want: "model hash mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := startInferenceMTLSTestServerWithService(t, material, tt.service)
			client, err := NewSidecarClient(validMTLSTestInferenceConfig(addr, material))
			if err != nil {
				t.Fatalf("create sidecar client: %v", err)
			}
			t.Cleanup(func() { _ = client.Close() })

			_, err = client.callSidecar(context.Background(), make([]float32, TotalFeatureDim), &ScoreInputs{})
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want) {
				t.Fatalf("expected response %s, got %v", tt.want, err)
			}
		})
	}
}

func TestParseSidecarServerCAPoolRequiresSigningCA(t *testing.T) {
	material := generateSidecarClientMTLSTestMaterial(t)
	leafPEM, err := os.ReadFile(material.serverCert)
	if err != nil {
		t.Fatalf("read leaf cert: %v", err)
	}

	_, err = parseSidecarServerCAPool(leafPEM)
	if err == nil {
		t.Fatal("expected leaf-only server CA bundle to fail")
	}
	if !strings.Contains(err.Error(), "CA certificate authorized for certificate signing") {
		t.Fatalf("expected CA signing error, got %v", err)
	}
}

func TestParseSidecarServerCAPoolRejectsMalformedTrailingPEM(t *testing.T) {
	material := generateSidecarClientMTLSTestMaterial(t)
	caPEM, err := os.ReadFile(material.serverCA)
	if err != nil {
		t.Fatalf("read server ca: %v", err)
	}

	_, err = parseSidecarServerCAPool(append(caPEM, []byte("\nnot a pem block")...))
	if err == nil {
		t.Fatal("expected malformed trailing PEM data to fail")
	}
	if !strings.Contains(err.Error(), "malformed trailing PEM data") {
		t.Fatalf("expected malformed trailing PEM error, got %v", err)
	}
}

func validMTLSTestInferenceConfig(addr string, material sidecarClientMTLSTestMaterial) InferenceConfig {
	return InferenceConfig{
		ModelVersion:            "v1.2.3",
		ExpectedHash:            strings.Repeat("a", 64),
		Timeout:                 time.Second,
		MaxMemoryMB:             512,
		UseSidecar:              true,
		SidecarAddress:          addr,
		SidecarTimeout:          2 * time.Second,
		SidecarTLS:              true,
		SidecarTLSCertFile:      material.clientCert,
		SidecarTLSKeyFile:       material.clientKey,
		SidecarTLSServerCAFile:  material.serverCA,
		SidecarTLSServerName:    "localhost",
		Deterministic:           true,
		ForceCPU:                true,
		RandomSeed:              42,
		ExpectedInputDim:        TotalFeatureDim,
		RequireHashVerification: true,
		StrictDeterminism:       true,
		UseFallbackOnError:      false,
		AllowFallbackToStub:     false,
		FallbackScore:           0,
		FallbackConfidence:      0,
		LogInferenceDetails:     false,
		LogInputHashes:          false,
		LogOutputHashes:         false,
	}
}

type sidecarClientMTLSTestMaterial struct {
	serverCA        string
	serverCert      string
	serverKey       string
	clientCA        string
	clientCert      string
	clientKey       string
	wrongClientCert string
	wrongClientKey  string
}

func generateSidecarClientMTLSTestMaterial(t *testing.T) sidecarClientMTLSTestMaterial {
	t.Helper()
	dir := t.TempDir()

	serverCA, serverCAKey := newSidecarClientTestCA(t, "server-ca")
	clientCA, clientCAKey := newSidecarClientTestCA(t, "client-ca")
	wrongCA, wrongCAKey := newSidecarClientTestCA(t, "wrong-client-ca")

	material := sidecarClientMTLSTestMaterial{
		serverCA:        filepath.Join(dir, "server-ca.pem"),
		serverCert:      filepath.Join(dir, "server.pem"),
		serverKey:       filepath.Join(dir, "server-key.pem"),
		clientCA:        filepath.Join(dir, "client-ca.pem"),
		clientCert:      filepath.Join(dir, "client.pem"),
		clientKey:       filepath.Join(dir, "client-key.pem"),
		wrongClientCert: filepath.Join(dir, "wrong-client.pem"),
		wrongClientKey:  filepath.Join(dir, "wrong-client-key.pem"),
	}
	writeSidecarClientCertPEM(t, material.serverCA, serverCA.Raw)
	writeSidecarClientCertPEM(t, material.clientCA, clientCA.Raw)
	writeSidecarClientLeafCert(t, material.serverCert, material.serverKey, "localhost", x509.ExtKeyUsageServerAuth, serverCA, serverCAKey)
	writeSidecarClientLeafCert(t, material.clientCert, material.clientKey, "client", x509.ExtKeyUsageClientAuth, clientCA, clientCAKey)
	writeSidecarClientLeafCert(t, material.wrongClientCert, material.wrongClientKey, "wrong-client", x509.ExtKeyUsageClientAuth, wrongCA, wrongCAKey)
	return material
}

func startInferenceMTLSTestServer(t *testing.T, material sidecarClientMTLSTestMaterial) string {
	return startInferenceMTLSTestServerWithService(t, material, sidecarClientModelInfoServer{})
}

func startInferenceMTLSTestServerWithService(t *testing.T, material sidecarClientMTLSTestMaterial, service sidecarClientModelInfoServer) string {
	t.Helper()
	serverCert, err := tls.LoadX509KeyPair(material.serverCert, material.serverKey)
	if err != nil {
		t.Fatalf("load server cert: %v", err)
	}
	clientCAPEM, err := os.ReadFile(material.clientCA)
	if err != nil {
		t.Fatalf("read client ca: %v", err)
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(clientCAPEM) {
		t.Fatal("append client ca")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(&tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAs,
		})),
	)
	inferencepb.RegisterInferenceServiceServer(server, service)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(inferencepb.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		server.GracefulStop()
		_ = listener.Close()
	})
	return listener.Addr().String()
}

func assertStandardGRPCHealthOverMTLS(t *testing.T, addr string, material sidecarClientMTLSTestMaterial) {
	t.Helper()
	clientCert, err := tls.LoadX509KeyPair(material.clientCert, material.clientKey)
	if err != nil {
		t.Fatalf("load client cert for health check: %v", err)
	}
	serverCAPEM, err := os.ReadFile(material.serverCA)
	if err != nil {
		t.Fatalf("read server ca for health check: %v", err)
	}
	rootCAs, err := parseSidecarServerCAPool(serverCAPEM)
	if err != nil {
		t.Fatalf("parse server ca for health check: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      rootCAs,
			ServerName:   "localhost",
		})),
	)
	if err != nil {
		t.Fatalf("dial standard grpc health over mtls: %v", err)
	}
	defer conn.Close()

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: inferencepb.ServiceName})
	if err != nil {
		t.Fatalf("standard grpc health check over mtls failed: %v", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected standard grpc health SERVING, got %s", resp.Status.String())
	}
}

type sidecarClientModelInfoServer struct {
	inferencepb.UnimplementedInferenceServiceServer
	scoreModelVersion string
	scoreModelHash    string
}

func (sidecarClientModelInfoServer) GetModelInfo(context.Context, *inferencepb.GetModelInfoRequest) (*inferencepb.GetModelInfoResponse, error) {
	return &inferencepb.GetModelInfoResponse{
		Version:  "v1.2.3",
		Hash:     strings.Repeat("a", 64),
		InputDim: int32(TotalFeatureDim),
	}, nil
}

func (s sidecarClientModelInfoServer) ComputeScore(context.Context, *inferencepb.ComputeScoreRequest) (*inferencepb.ComputeScoreResponse, error) {
	modelVersion := s.scoreModelVersion
	if modelVersion == "" {
		modelVersion = "v1.2.3"
	}
	modelHash := s.scoreModelHash
	if modelHash == "" {
		modelHash = strings.Repeat("a", 64)
	}
	rawScore := float32(50)
	return &inferencepb.ComputeScoreResponse{
		Score:        50,
		RawScore:     rawScore,
		Confidence:   1,
		OutputHash:   NewDeterminismController(42, true).ComputeOutputHash([]float32{rawScore}),
		ModelVersion: modelVersion,
		ModelHash:    modelHash,
	}, nil
}

func newSidecarClientTestCA(t *testing.T, cn string) (*x509.Certificate, *rsa.PrivateKey) {
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

func writeSidecarClientLeafCert(t *testing.T, certPath, keyPath, cn string, usage x509.ExtKeyUsage, ca *x509.Certificate, caKey *rsa.PrivateKey) {
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
	writeSidecarClientCertPEM(t, certPath, der)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("write key %s: %v", keyPath, err)
	}
}

func writeSidecarClientCertPEM(t *testing.T, path string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatalf("write cert %s: %v", path, err)
	}
}
