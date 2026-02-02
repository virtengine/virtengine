// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Reload Status
// ============================================================================

// ReloadStatus contains status information about model reload operations
type ReloadStatus struct {
	// LastReloadTime is the timestamp of the last successful reload
	LastReloadTime time.Time

	// LastReloadDuration is the duration of the last reload operation
	LastReloadDuration time.Duration

	// SuccessfulReloads is the count of successful reload operations
	SuccessfulReloads uint64

	// FailedReloads is the count of failed reload operations
	FailedReloads uint64

	// CurrentModelHash is the SHA256 hash of the currently loaded model
	CurrentModelHash string

	// CurrentModelVersion is the version of the currently loaded model
	CurrentModelVersion string

	// IsReloading indicates if a reload operation is currently in progress
	IsReloading bool

	// LastError contains the error message from the last failed reload
	LastError string

	// LastErrorTime is the timestamp of the last failed reload
	LastErrorTime time.Time
}

// ============================================================================
// Model Reload Manager
// ============================================================================

// ModelReloadManager handles atomic model hot-reload with version tracking
// and rollback capabilities. It ensures no requests are dropped during reload
// by using atomic model swapping.
type ModelReloadManager struct {
	// config holds the inference configuration
	config InferenceConfig

	// mu protects all mutable state
	mu sync.RWMutex

	// currentModel is the currently active model
	currentModel *TFModel

	// modelLoader is used to load new models
	modelLoader *ModelLoader

	// status contains reload status information
	status ReloadStatus

	// isReloading is an atomic flag for reload-in-progress
	isReloading atomic.Bool

	// stopWatcher signals the file watcher to stop
	stopWatcher chan struct{}

	// watcherRunning indicates if the file watcher is active
	watcherRunning atomic.Bool

	// closed indicates if the manager has been closed
	closed atomic.Bool

	// logger for reload operations
	logger *log.Logger
}

// NewModelReloadManager creates a new ModelReloadManager with the given configuration
func NewModelReloadManager(config InferenceConfig) *ModelReloadManager {
	return &ModelReloadManager{
		config:      config,
		modelLoader: NewModelLoader(config),
		status: ReloadStatus{
			CurrentModelHash:    "",
			CurrentModelVersion: "",
		},
		stopWatcher: make(chan struct{}),
		logger:      log.New(os.Stdout, "[ModelReloadManager] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// ============================================================================
// Model Loading and Reloading
// ============================================================================

// LoadInitial performs the initial model load
// This should be called once during initialization
func (m *ModelReloadManager) LoadInitial() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed.Load() {
		return fmt.Errorf("reload manager is closed")
	}

	m.logger.Printf("Loading initial model from: %s", m.config.ModelPath)

	model, err := m.modelLoader.Load()
	if err != nil {
		return fmt.Errorf("failed to load initial model: %w", err)
	}

	m.currentModel = model
	m.status.CurrentModelHash = model.GetModelHash()
	m.status.CurrentModelVersion = model.GetVersion()
	m.status.LastReloadTime = time.Now()

	m.logger.Printf("Initial model loaded successfully: version=%s, hash=%s",
		m.status.CurrentModelVersion, m.status.CurrentModelHash)

	return nil
}

// Reload performs an atomic model reload from the specified path
// If the reload fails, the previous model remains active (rollback)
// No requests are dropped during the reload operation
func (m *ModelReloadManager) Reload(newModelPath string) error {
	if m.closed.Load() {
		return fmt.Errorf("reload manager is closed")
	}

	// Prevent concurrent reloads
	if !m.isReloading.CompareAndSwap(false, true) {
		return fmt.Errorf("reload already in progress")
	}
	defer m.isReloading.Store(false)

	startTime := time.Now()
	m.logger.Printf("Starting model reload from: %s", newModelPath)

	// Create a new loader configuration with the new path
	newConfig := m.config.WithModelPath(newModelPath)
	newLoader := NewModelLoader(newConfig)

	// Attempt to load the new model (outside the write lock)
	newModel, err := newLoader.Load()
	if err != nil {
		m.recordFailure(err)
		m.logger.Printf("Model reload failed: %v", err)
		return fmt.Errorf("failed to load new model: %w", err)
	}

	// Compute hash of new model for verification
	newHash := newModel.GetModelHash()
	newVersion := newModel.GetVersion()

	m.logger.Printf("New model loaded: version=%s, hash=%s", newVersion, newHash)

	// Atomic swap: acquire write lock only for the swap operation
	m.mu.Lock()
	oldModel := m.currentModel
	m.currentModel = newModel

	// Update status
	duration := time.Since(startTime)
	m.status.LastReloadTime = time.Now()
	m.status.LastReloadDuration = duration
	m.status.SuccessfulReloads++
	m.status.CurrentModelHash = newHash
	m.status.CurrentModelVersion = newVersion
	m.status.LastError = ""
	m.mu.Unlock()

	// Clean up old model after swap (outside the lock)
	if oldModel != nil {
		if err := oldModel.Close(); err != nil {
			m.logger.Printf("Warning: failed to close old model: %v", err)
		}
	}

	m.logger.Printf("Model reload completed successfully in %v", duration)

	return nil
}

// recordFailure records a failed reload attempt
func (m *ModelReloadManager) recordFailure(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.status.FailedReloads++
	m.status.LastError = err.Error()
	m.status.LastErrorTime = time.Now()
}

// ============================================================================
// Model Access
// ============================================================================

// GetCurrentModel returns the currently active model
// This is thread-safe and non-blocking for read operations
func (m *ModelReloadManager) GetCurrentModel() *TFModel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentModel
}

// GetCurrentVersion returns the version of the currently loaded model
func (m *ModelReloadManager) GetCurrentVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.CurrentModelVersion
}

// GetCurrentHash returns the hash of the currently loaded model
func (m *ModelReloadManager) GetCurrentHash() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status.CurrentModelHash
}

// GetReloadStatus returns the current reload status
func (m *ModelReloadManager) GetReloadStatus() ReloadStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := m.status
	status.IsReloading = m.isReloading.Load()
	return status
}

// ============================================================================
// File Watcher for Auto-Reload
// ============================================================================

// WatchForChanges starts watching the specified path for model file changes
// When changes are detected, the model is automatically reloaded
// This runs in a separate goroutine and can be stopped with Close()
func (m *ModelReloadManager) WatchForChanges(path string) error {
	if m.closed.Load() {
		return fmt.Errorf("reload manager is closed")
	}

	if m.watcherRunning.Load() {
		return fmt.Errorf("watcher already running")
	}

	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("watch path does not exist: %s", path)
	}

	m.watcherRunning.Store(true)
	m.logger.Printf("Starting file watcher for: %s", path)

	go m.watchLoop(path)

	return nil
}

// watchLoop is the main loop for the file watcher
func (m *ModelReloadManager) watchLoop(path string) {
	defer m.watcherRunning.Store(false)

	// Compute initial hash of model directory
	lastHash, err := m.computeDirectoryHash(path)
	if err != nil {
		m.logger.Printf("Warning: failed to compute initial directory hash: %v", err)
		lastHash = ""
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopWatcher:
			m.logger.Printf("File watcher stopped")
			return
		case <-ticker.C:
			// Check for changes
			currentHash, err := m.computeDirectoryHash(path)
			if err != nil {
				m.logger.Printf("Warning: failed to compute directory hash: %v", err)
				continue
			}

			if currentHash != lastHash && lastHash != "" {
				m.logger.Printf("Detected model change, triggering reload")
				if err := m.Reload(path); err != nil {
					m.logger.Printf("Auto-reload failed: %v", err)
				} else {
					lastHash = currentHash
				}
			} else if lastHash == "" {
				lastHash = currentHash
			}
		}
	}
}

// computeDirectoryHash computes a hash of all files in a directory
// for change detection purposes
func (m *ModelReloadManager) computeDirectoryHash(dirPath string) (string, error) {
	h := sha256.New()

	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Sort for deterministic hash
	sort.Strings(files)

	for _, path := range files {
		// Include file path and modification time in hash
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write([]byte(info.ModTime().String()))
		_, _ = io.WriteString(h, fmt.Sprintf("%d", info.Size()))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// StopWatcher stops the file watcher if running
func (m *ModelReloadManager) StopWatcher() {
	if m.watcherRunning.Load() {
		close(m.stopWatcher)
		// Wait for watcher to stop
		for m.watcherRunning.Load() {
			time.Sleep(10 * time.Millisecond)
		}
		// Reset channel for potential restart
		m.stopWatcher = make(chan struct{})
	}
}

// ============================================================================
// Lifecycle Management
// ============================================================================

// Close stops the file watcher and releases all resources
func (m *ModelReloadManager) Close() error {
	if m.closed.Swap(true) {
		return nil // Already closed
	}

	m.logger.Printf("Closing ModelReloadManager")

	// Stop watcher if running
	m.StopWatcher()

	// Close current model
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentModel != nil {
		if err := m.currentModel.Close(); err != nil {
			return fmt.Errorf("failed to close current model: %w", err)
		}
		m.currentModel = nil
	}

	// Unload the model loader's model
	if err := m.modelLoader.Unload(); err != nil {
		return fmt.Errorf("failed to unload model loader: %w", err)
	}

	m.logger.Printf("ModelReloadManager closed successfully")
	return nil
}

// IsHealthy returns true if the manager has a loaded model and is operational
func (m *ModelReloadManager) IsHealthy() bool {
	if m.closed.Load() {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentModel != nil && m.currentModel.IsLoaded()
}

// ============================================================================
// Inference Execution
// ============================================================================

// Run executes inference on the current model
// This provides a convenience wrapper that handles model access safely
func (m *ModelReloadManager) Run(features []float32) ([]float32, error) {
	if m.closed.Load() {
		return nil, fmt.Errorf("reload manager is closed")
	}

	model := m.GetCurrentModel()
	if model == nil {
		return nil, fmt.Errorf("no model loaded")
	}

	return model.Run(features)
}

// ============================================================================
// Configuration Helpers
// ============================================================================

// SetLogger sets a custom logger for the reload manager
func (m *ModelReloadManager) SetLogger(logger *log.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

// GetConfig returns the current inference configuration
func (m *ModelReloadManager) GetConfig() InferenceConfig {
	return m.config
}
