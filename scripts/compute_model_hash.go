// Copyright 2024 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// compute_model_hash is a standalone CLI tool for computing deterministic
// SHA256 hashes of ML model files for VEID governance. It mirrors the hash
// computation logic used by the on-chain keeper (ComputeLocalModelHash).
//
// Usage:
//
//	go run scripts/compute_model_hash.go -dir ml/facial_verification/weights -type face_verification
//	go run scripts/compute_model_hash.go -dir ml/liveness_detection/weights -type liveness
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// modelExtensions lists file extensions that contain model weights / frozen graphs.
var modelExtensions = map[string]bool{
	".pb":          true,
	".h5":          true,
	".tflite":      true,
	".onnx":        true,
	".savedmodel":  true,
	".pt":          true,
	".pth":         true,
	".bin":         true,
	".safetensors": true,
}

// excludePatterns lists filename substrings to exclude from hashing (metadata files).
var excludePatterns = []string{
	"readme", "license", "changelog", "manifest",
	"checksum", "metadata",
}

// HashResult is the JSON output structure.
type HashResult struct {
	SHA256Hash string `json:"sha256_hash"`
	ModelType  string `json:"model_type"`
	Version    string `json:"version"`
	FileCount  int    `json:"file_count"`
	Timestamp  string `json:"timestamp"`
	Directory  string `json:"directory"`
}

func main() {
	dir := flag.String("dir", "", "Path to model directory (required)")
	modelType := flag.String("type", "", "Model type (defaults to directory name)")
	flag.Parse()

	if *dir == "" {
		fmt.Fprintln(os.Stderr, "Usage: compute_model_hash -dir <model_dir> [-type <model_type>]")
		os.Exit(1)
	}

	if *modelType == "" {
		*modelType = filepath.Base(*dir)
	}

	info, err := os.Stat(*dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: directory not found: %s\n", *dir)
		os.Exit(1)
	}

	hash, fileCount, err := computeHash(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	version := readVersion(*dir)

	result := HashResult{
		SHA256Hash: hash,
		ModelType:  *modelType,
		Version:    version,
		FileCount:  fileCount,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Directory:  *dir,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

// computeHash walks the directory, finds model files sorted lexicographically,
// hashes each file with SHA256, then hashes the concatenated hex digests.
func computeHash(dir string) (string, int, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isModelFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", 0, fmt.Errorf("walking directory: %w", err)
	}

	if len(files) == 0 {
		return "", 0, fmt.Errorf("no model files found in %s", dir)
	}

	sort.Strings(files)

	var combined strings.Builder
	for _, f := range files {
		h, err := hashFile(f)
		if err != nil {
			return "", 0, fmt.Errorf("hashing %s: %w", f, err)
		}
		combined.WriteString(h)
	}

	finalHash := sha256.Sum256([]byte(combined.String()))
	return hex.EncodeToString(finalHash[:]), len(files), nil
}

// hashFile computes SHA256 of a single file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// isModelFile returns true if the file has a model extension and is not excluded.
func isModelFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))

	if !modelExtensions[ext] {
		return false
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(base, pattern) {
			return false
		}
	}
	return true
}

// readVersion attempts to read a version file from the model directory or parent.
func readVersion(dir string) string {
	for _, path := range []string{
		filepath.Join(dir, "version.txt"),
		filepath.Join(dir, "..", "version.txt"),
	} {
		data, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return "unknown"
}
