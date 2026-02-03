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
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

// Version info (set at build time)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Command line flags
var (
	grpcAddr      = flag.String("grpc-addr", ":50051", "gRPC server address")
	metricsAddr   = flag.String("metrics-addr", ":9090", "Prometheus metrics address")
	modelPath     = flag.String("model-path", "models/trust_score", "Path to TensorFlow SavedModel")
	modelVersion  = flag.String("model-version", "v1.0.0", "Expected model version")
	expectedHash  = flag.String("expected-hash", "", "Expected SHA256 hash of model weights")
	randomSeed    = flag.Int64("random-seed", 42, "Random seed for deterministic execution")
	forceCPU      = flag.Bool("force-cpu", true, "Force CPU-only execution")
	maxMemoryMB   = flag.Int("max-memory-mb", 512, "Maximum memory usage in MB")
	timeout       = flag.Duration("timeout", 2*time.Second, "Inference timeout")
	logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	enableReflect = flag.Bool("enable-reflection", false, "Enable gRPC reflection for debugging")
	servingURL    = flag.String("serving-url", "http://localhost:8501", "TensorFlow Serving base URL")
	servingModel  = flag.String("serving-model", "trust_score", "TensorFlow Serving model name")
	servingSig    = flag.String("serving-signature", "", "TensorFlow Serving signature name")
	servingTO     = flag.Duration("serving-timeout", 5*time.Second, "TensorFlow Serving request timeout")
	servingHealth = flag.String("serving-health-path", "", "Optional TensorFlow Serving health path override")
	servingFail   = flag.String("serving-fallback-url", "", "Fallback TensorFlow Serving base URL")
	allowStub     = flag.Bool("allow-fallback-to-stub", false, "Allow fallback to local stub inference on serving failure")
)

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()

	// Setup logging
	log := setupLogger(*logLevel)

	log.Info("Starting inference sidecar",
		"version", Version,
		"git_commit", GitCommit,
		"build_time", BuildTime,
	)

	// Build inference configuration
	config := inference.InferenceConfig{
		ModelPath:           *modelPath,
		ModelVersion:        *modelVersion,
		ExpectedHash:        *expectedHash,
		Timeout:             *timeout,
		MaxMemoryMB:         *maxMemoryMB,
		UseSidecar:          false, // We ARE the sidecar
		Deterministic:       true,
		ForceCPU:            *forceCPU,
		RandomSeed:          *randomSeed,
		ExpectedInputDim:    inference.TotalFeatureDim,
		UseFallbackOnError:  false, // Sidecar should report errors
		AllowFallbackToStub: *allowStub,
	}

	// Create the inference server
	servingConfig := inference.TFServingConfig{
		BaseURL:       *servingURL,
		FallbackURL:   *servingFail,
		ModelName:     *servingModel,
		SignatureName: *servingSig,
		Timeout:       *servingTO,
		HealthPath:    *servingHealth,
	}

	server, err := NewInferenceSidecarServer(config, servingConfig, log)
	if err != nil {
		log.Error("Failed to create inference server", "error", err)
		return 1
	}
	defer server.Close()

	// Start metrics server
	go startMetricsServer(*metricsAddr, log)

	// Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(16*1024*1024), // 16MB max message size
		grpc.MaxSendMsgSize(16*1024*1024),
	)

	// Register services
	inferencepb.RegisterInferenceServiceServer(grpcServer, server)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus(inferencepb.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for debugging if requested
	if *enableReflect {
		reflection.Register(grpcServer)
		log.Info("gRPC reflection enabled")
	}

	// Start listening
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Error("Failed to listen", "error", err, "addr", *grpcAddr)
		return 1
	}

	log.Info("gRPC server listening", "addr", *grpcAddr)

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

func startMetricsServer(addr string, log Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Info("Metrics server listening", "addr", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Error("Metrics server error", "error", err)
	}
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
