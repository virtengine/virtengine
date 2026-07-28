// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package main provides the inference sidecar server for VEID identity scoring.
// This server loads a TensorFlow SavedModel and provides deterministic inference
// via gRPC for blockchain consensus-critical scoring.
//
// VE-219: Deterministic identity verification runtime
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
	"github.com/virtengine/virtengine/pkg/observability"
)

// Version info (set at build time)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

type sidecarOptions struct {
	GRPCAddr            string
	MetricsAddr         string
	ModelPath           string
	ModelVersion        string
	ExpectedHash        string
	ManifestPath        string
	RandomSeed          int64
	ForceCPU            bool
	MaxMemoryMB         int
	Timeout             time.Duration
	LogLevel            string
	EnableReflection    bool
	ServingURL          string
	ServingModel        string
	ServingSignature    string
	ServingTimeout      time.Duration
	ServingHealthPath   string
	ServingFallbackURL  string
	AllowFallbackToStub bool
	RequireMTLS         bool
	TLSCertFile         string
	TLSKeyFile          string
	TLSClientCAFile     string
}

type sidecarTLSOptions struct {
	RequireMTLS  bool
	CertFile     string
	KeyFile      string
	ClientCAFile string
}

func main() {
	os.Exit(run())
}

func run() int {
	opts, err := parseSidecarOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid inference sidecar options: %v\n", err)
		return 2
	}

	// Setup logging
	log := setupLogger(opts.LogLevel)

	if err := validateStartupOptions(opts); err != nil {
		log.Error("Invalid inference sidecar startup policy", "error", err)
		return 1
	}
	tlsConfig, err := buildServerTLSConfig(sidecarTLSOptions{
		RequireMTLS:  opts.RequireMTLS,
		CertFile:     opts.TLSCertFile,
		KeyFile:      opts.TLSKeyFile,
		ClientCAFile: opts.TLSClientCAFile,
	})
	if err != nil {
		log.Error("Invalid inference sidecar TLS policy", "error", err)
		return 1
	}

	log.Info("Starting inference sidecar",
		"version", Version,
		"git_commit", GitCommit,
		"build_time", BuildTime,
		"require_mtls", opts.RequireMTLS,
	)

	// Build inference configuration
	config := buildInferenceConfig(opts)

	// Create the inference server
	servingConfig := buildServingConfig(opts)

	server, err := NewInferenceSidecarServer(config, servingConfig, opts.ManifestPath, log)
	if err != nil {
		log.Error("Failed to create inference server", "error", err)
		return 1
	}
	defer server.Close()

	if readiness := server.Readiness(); readiness != nil && !readiness.Ready() {
		log.Warn("Inference sidecar started not ready",
			"verification_state", readiness.State,
			"manifest_path", readiness.ManifestPath,
			"path", readiness.FailurePath,
			"error", readiness.FailureReason,
		)
	}

	// Start metrics server
	go startMetricsServer(opts.MetricsAddr, server, log)

	// Start gRPC server
	grpcServerOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(16 * 1024 * 1024), // 16MB max message size
		grpc.MaxSendMsgSize(16 * 1024 * 1024),
		grpc.StatsHandler(observability.GRPCServerStatsHandler()),
	}
	if tlsConfig != nil {
		grpcServerOptions = append(grpcServerOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	grpcServer := grpc.NewServer(grpcServerOptions...)

	// Register services
	inferencepb.RegisterInferenceServiceServer(grpcServer, server)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthStatus := grpc_health_v1.HealthCheckResponse_SERVING
	if readiness := server.Readiness(); readiness == nil || !readiness.Ready() {
		healthStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	healthServer.SetServingStatus(inferencepb.ServiceName, healthStatus)

	// Enable reflection for debugging if requested
	if opts.EnableReflection {
		reflection.Register(grpcServer)
		log.Info("gRPC reflection enabled")
	}

	// Start listening
	lis, err := net.Listen("tcp", opts.GRPCAddr)
	if err != nil {
		log.Error("Failed to listen", "error", err, "addr", opts.GRPCAddr)
		return 1
	}

	log.Info("gRPC server listening", "addr", opts.GRPCAddr, "require_mtls", opts.RequireMTLS)

	// Handle shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run server in background
	errChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		log.Info("Shutdown signal received")
		grpcServer.GracefulStop()
	case err := <-errChan:
		log.Error("Server error", "error", err)
		return 1
	}

	log.Info("Server stopped gracefully")
	return 0
}

func defaultSidecarOptions() sidecarOptions {
	return sidecarOptions{
		GRPCAddr:           ":50051",
		MetricsAddr:        ":9092",
		ModelPath:          "models/trust_score/current/model",
		ModelVersion:       "",
		ExpectedHash:       "",
		ManifestPath:       "",
		RandomSeed:         42,
		ForceCPU:           true,
		MaxMemoryMB:        512,
		Timeout:            2 * time.Second,
		LogLevel:           "info",
		EnableReflection:   false,
		ServingURL:         "http://localhost:8501",
		ServingModel:       "trust_score",
		ServingSignature:   "",
		ServingTimeout:     5 * time.Second,
		ServingHealthPath:  "",
		ServingFallbackURL: "",
		RequireMTLS:        true,
	}
}

func parseSidecarOptions(args []string) (sidecarOptions, error) {
	opts := defaultSidecarOptions()
	fs := flag.NewFlagSet("inference-sidecar", flag.ContinueOnError)
	fs.StringVar(&opts.GRPCAddr, "grpc-addr", opts.GRPCAddr, "gRPC server address")
	fs.StringVar(&opts.MetricsAddr, "metrics-addr", opts.MetricsAddr, "Prometheus metrics and readiness address")
	fs.StringVar(&opts.ModelPath, "model-path", opts.ModelPath, "Path to TensorFlow SavedModel")
	fs.StringVar(&opts.ModelVersion, "model-version", opts.ModelVersion, "Expected model version; defaults to release_manifest.json when unset")
	fs.StringVar(&opts.ExpectedHash, "expected-hash", opts.ExpectedHash, "Expected SHA256 hash of model weights; defaults to release_manifest.json when unset")
	fs.StringVar(&opts.ManifestPath, "manifest-path", opts.ManifestPath, "Path to release_manifest.json; defaults to a sibling of --model-path")
	fs.Int64Var(&opts.RandomSeed, "random-seed", opts.RandomSeed, "Random seed for deterministic execution")
	fs.BoolVar(&opts.ForceCPU, "force-cpu", opts.ForceCPU, "Force CPU-only execution")
	fs.IntVar(&opts.MaxMemoryMB, "max-memory-mb", opts.MaxMemoryMB, "Maximum memory usage in MB")
	fs.DurationVar(&opts.Timeout, "timeout", opts.Timeout, "Inference timeout")
	fs.StringVar(&opts.LogLevel, "log-level", opts.LogLevel, "Log level (debug, info, warn, error)")
	fs.BoolVar(&opts.EnableReflection, "enable-reflection", opts.EnableReflection, "Enable gRPC reflection for debugging")
	fs.StringVar(&opts.ServingURL, "serving-url", opts.ServingURL, "TensorFlow Serving base URL")
	fs.StringVar(&opts.ServingModel, "serving-model", opts.ServingModel, "TensorFlow Serving model name")
	fs.StringVar(&opts.ServingSignature, "serving-signature", opts.ServingSignature, "TensorFlow Serving signature name")
	fs.DurationVar(&opts.ServingTimeout, "serving-timeout", opts.ServingTimeout, "TensorFlow Serving request timeout")
	fs.StringVar(&opts.ServingHealthPath, "serving-health-path", opts.ServingHealthPath, "Optional TensorFlow Serving health path override")
	fs.StringVar(&opts.ServingFallbackURL, "serving-fallback-url", opts.ServingFallbackURL, "Fallback TensorFlow Serving base URL")
	fs.BoolVar(&opts.AllowFallbackToStub, "allow-fallback-to-stub", opts.AllowFallbackToStub, "Allow fallback to local stub inference on serving failure")
	fs.BoolVar(&opts.RequireMTLS, "require-mtls", opts.RequireMTLS, "Require TLS 1.3 and verified client certificates for gRPC")
	fs.StringVar(&opts.TLSCertFile, "tls-cert-file", opts.TLSCertFile, "Absolute path to inference sidecar server certificate")
	fs.StringVar(&opts.TLSKeyFile, "tls-key-file", opts.TLSKeyFile, "Absolute path to inference sidecar server private key")
	fs.StringVar(&opts.TLSClientCAFile, "tls-client-ca-file", opts.TLSClientCAFile, "Absolute path to client CA bundle for mTLS")
	if err := fs.Parse(args); err != nil {
		return sidecarOptions{}, err
	}
	return opts, nil
}

func buildInferenceConfig(opts sidecarOptions) inference.InferenceConfig {
	return inference.InferenceConfig{
		ModelPath:           opts.ModelPath,
		ModelVersion:        opts.ModelVersion,
		ExpectedHash:        opts.ExpectedHash,
		Timeout:             opts.Timeout,
		MaxMemoryMB:         opts.MaxMemoryMB,
		UseSidecar:          false, // We ARE the sidecar
		Deterministic:       true,
		ForceCPU:            opts.ForceCPU,
		RandomSeed:          opts.RandomSeed,
		ExpectedInputDim:    inference.TotalFeatureDim,
		UseFallbackOnError:  false, // Sidecar should report errors
		AllowFallbackToStub: opts.AllowFallbackToStub,
	}
}

func buildServingConfig(opts sidecarOptions) inference.TFServingConfig {
	return inference.TFServingConfig{
		BaseURL:       opts.ServingURL,
		FallbackURL:   opts.ServingFallbackURL,
		ModelName:     opts.ServingModel,
		SignatureName: opts.ServingSignature,
		Timeout:       opts.ServingTimeout,
		HealthPath:    opts.ServingHealthPath,
	}
}

func validateStartupOptions(opts sidecarOptions) error {
	if !opts.RequireMTLS {
		return nil
	}
	if opts.AllowFallbackToStub {
		return fmt.Errorf("--allow-fallback-to-stub is rejected when --require-mtls=true")
	}
	if strings.TrimSpace(opts.ServingURL) == "" {
		return fmt.Errorf("--serving-url is required when --require-mtls=true to prevent local stub execution")
	}
	if strings.TrimSpace(opts.ServingFallbackURL) != "" {
		return fmt.Errorf("--serving-fallback-url is rejected when --require-mtls=true because fallback runtime identity is not pinned")
	}
	if !opts.ForceCPU {
		return fmt.Errorf("--force-cpu=true is required when --require-mtls=true")
	}
	if opts.RandomSeed != 42 {
		return fmt.Errorf("--random-seed=42 is required when --require-mtls=true")
	}
	return nil
}

func buildServerTLSConfig(opts sidecarTLSOptions) (*tls.Config, error) {
	if !opts.RequireMTLS {
		return nil, nil
	}

	certFile, err := validateTLSFile("--tls-cert-file", opts.CertFile)
	if err != nil {
		return nil, err
	}
	keyFile, err := validateTLSFile("--tls-key-file", opts.KeyFile)
	if err != nil {
		return nil, err
	}
	clientCAFile, err := validateTLSFile("--tls-client-ca-file", opts.ClientCAFile)
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load mTLS certificate key pair from --tls-cert-file/--tls-key-file: %w", err)
	}
	if err := validateServerLeafCertificate(cert, time.Now().UTC()); err != nil {
		return nil, err
	}

	clientCAPEM, err := os.ReadFile(clientCAFile)
	if err != nil {
		return nil, fmt.Errorf("read --tls-client-ca-file: %w", err)
	}
	clientCAs, err := parseClientCAPool(clientCAPEM)
	if err != nil {
		return nil, fmt.Errorf("--tls-client-ca-file: %w", err)
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
	}, nil
}

func validateServerLeafCertificate(cert tls.Certificate, now time.Time) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("mTLS server certificate chain is empty")
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("parse mTLS server certificate: %w", err)
	}
	if now.Before(leaf.NotBefore) || !now.Before(leaf.NotAfter) {
		return fmt.Errorf("mTLS server certificate is outside its validity window")
	}
	if leaf.IsCA {
		return fmt.Errorf("mTLS server certificate must be a leaf certificate")
	}
	if leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("mTLS server certificate must allow digital signatures")
	}
	for _, usage := range leaf.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth || usage == x509.ExtKeyUsageAny {
			return nil
		}
	}
	return fmt.Errorf("mTLS server certificate must allow server authentication")
}

func parseClientCAPool(clientCAPEM []byte) (*x509.CertPool, error) {
	remaining := strings.TrimSpace(string(clientCAPEM))
	if remaining == "" {
		return nil, fmt.Errorf("client CA bundle must contain at least one PEM certificate")
	}

	clientCAs := x509.NewCertPool()
	caCount := 0
	for {
		block, rest := pem.Decode([]byte(remaining))
		if block == nil {
			if strings.TrimSpace(remaining) != "" {
				return nil, fmt.Errorf("client CA bundle contains malformed trailing PEM data")
			}
			break
		}
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("client CA bundle contains unsupported PEM block %q", block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse client CA certificate: %w", err)
		}
		if cert.IsCA && cert.BasicConstraintsValid && cert.KeyUsage&x509.KeyUsageCertSign != 0 {
			clientCAs.AddCert(cert)
			caCount++
		}
		remaining = strings.TrimSpace(string(rest))
		if remaining == "" {
			break
		}
	}
	if caCount == 0 {
		return nil, fmt.Errorf("client CA bundle must contain at least one CA certificate authorized for certificate signing")
	}
	return clientCAs, nil
}

func validateTLSFile(flagName, path string) (string, error) {
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return "", fmt.Errorf("%s is required when --require-mtls=true", flagName)
	}
	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("%s must be an absolute path", flagName)
	}
	info, err := os.Stat(cleaned)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", flagName, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s must reference a readable regular file", flagName)
	}
	file, err := os.Open(cleaned)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", flagName, err)
	}
	_ = file.Close()
	return cleaned, nil
}

func startMetricsServer(addr string, server *InferenceSidecarServer, log Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		statusCode, payload := runtimeReadinessHTTPResponse(r.Context(), server)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write(payload)
	})

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Info("Metrics server listening", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Error("Metrics server error", "error", err)
	}
}

func runtimeReadinessHTTPResponse(ctx context.Context, server *InferenceSidecarServer) (int, []byte) {
	if server == nil {
		return marshalReadinessStatus(readinessStatus{
			Ready:   false,
			State:   "runtime_unavailable",
			Message: "inference sidecar server is unavailable",
		})
	}

	statusCode, payload := readinessHTTPResponse(server.Readiness())
	if statusCode != http.StatusOK {
		return statusCode, payload
	}

	health, err := server.HealthCheck(ctx, &inferencepb.HealthCheckRequest{})
	if err == nil && health != nil && health.Status == inferencepb.HealthStatus_HEALTH_STATUS_HEALTHY {
		return statusCode, payload
	}

	message := "inference runtime health check failed"
	if err != nil {
		message = err.Error()
	} else if health != nil && strings.TrimSpace(health.ErrorMessage) != "" {
		message = health.ErrorMessage
	}
	return marshalReadinessStatus(readinessStatus{
		Ready:   false,
		State:   "runtime_unavailable",
		Message: message,
	})
}

func marshalReadinessStatus(status readinessStatus) (int, []byte) {
	payload, err := json.Marshal(status)
	if err != nil {
		return http.StatusServiceUnavailable, []byte(`{"ready":false,"state":"runtime_unavailable","message":"failed to encode readiness payload"}`)
	}
	statusCode := http.StatusServiceUnavailable
	if status.Ready {
		statusCode = http.StatusOK
	}
	return statusCode, payload
}

type readinessStatus struct {
	Ready        bool   `json:"ready"`
	State        string `json:"state"`
	Message      string `json:"message,omitempty"`
	ManifestPath string `json:"manifest_path,omitempty"`
	Path         string `json:"path,omitempty"`
	ModelVersion string `json:"model_version,omitempty"`
	ModelHash    string `json:"model_hash,omitempty"`
}

func readinessHTTPResponse(readiness *verificationResult) (int, []byte) {
	status := readinessStatus{
		Ready: false,
		State: "verification_unavailable",
	}

	if readiness == nil {
		status.Message = "model verification status unavailable"
	} else {
		status.State = string(readiness.State)
		status.ManifestPath = readiness.ManifestPath
		status.Path = readiness.FailurePath
		status.ModelVersion = readiness.Version
		status.ModelHash = readiness.ModelHash
		status.Message = readiness.StatusMessage()
		status.Ready = readiness.Ready()
	}

	statusCode := http.StatusServiceUnavailable
	if status.Ready {
		statusCode = http.StatusOK
	}

	payload, err := json.Marshal(status)
	if err != nil {
		return http.StatusServiceUnavailable, []byte(`{"ready":false,"state":"bad_manifest","message":"failed to encode readiness payload"}`)
	}
	return statusCode, payload
}

// Logger is a simple logging interface
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// simpleLogger implements Logger with basic stdout logging
type simpleLogger struct {
	level string
}

func setupLogger(level string) Logger {
	return &simpleLogger{level: level}
}

func (l *simpleLogger) shouldLog(level string) bool {
	levels := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	return levels[level] >= levels[l.level]
}

func (l *simpleLogger) log(level, msg string, keysAndValues ...interface{}) {
	if !l.shouldLog(level) {
		return
	}
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("%s [%s] %s", timestamp, level, msg)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		fmt.Printf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
	}
	fmt.Println()
}

func (l *simpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log("debug", msg, keysAndValues...)
}

func (l *simpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log("info", msg, keysAndValues...)
}

func (l *simpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log("warn", msg, keysAndValues...)
}

func (l *simpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log("error", msg, keysAndValues...)
}
