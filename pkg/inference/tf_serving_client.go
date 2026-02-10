// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// TensorFlow Serving client for VEID inference.

package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// TFServingConfig configures the TensorFlow Serving client.
type TFServingConfig struct {
	BaseURL       string
	FallbackURL   string
	ModelName     string
	InputName     string
	OutputName    string
	SignatureName string
	Timeout       time.Duration
	HealthPath    string
}

// TFServingClient handles REST calls to TensorFlow Serving.
type TFServingClient struct {
	baseURL       string
	fallbackURL   string
	modelName     string
	inputName     string
	outputName    string
	signatureName string
	healthPath    string
	httpClient    *http.Client
}

// NewTFServingClient creates a new TensorFlow Serving client.
func NewTFServingClient(config TFServingConfig) (*TFServingClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("tf serving base URL is required")
	}
	modelName := strings.TrimSpace(config.ModelName)
	if modelName == "" {
		return nil, fmt.Errorf("tf serving model name is required")
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &TFServingClient{
		baseURL:       baseURL,
		fallbackURL:   strings.TrimRight(strings.TrimSpace(config.FallbackURL), "/"),
		modelName:     modelName,
		inputName:     strings.TrimSpace(config.InputName),
		outputName:    strings.TrimSpace(config.OutputName),
		signatureName: strings.TrimSpace(config.SignatureName),
		healthPath:    strings.TrimSpace(config.HealthPath),
		httpClient:    security.NewSecureHTTPClient(security.WithTimeout(timeout)),
	}, nil
}

// Predict runs inference via TensorFlow Serving.
func (c *TFServingClient) Predict(ctx context.Context, features []float32) ([]float32, string, time.Duration, error) {
	start := time.Now()

	output, err := c.predictOnce(ctx, c.baseURL, features)
	if err == nil {
		return output, c.baseURL, time.Since(start), nil
	}
	if c.fallbackURL == "" {
		return nil, c.baseURL, time.Since(start), err
	}

	fallbackOutput, fallbackErr := c.predictOnce(ctx, c.fallbackURL, features)
	if fallbackErr == nil {
		return fallbackOutput, c.fallbackURL, time.Since(start), nil
	}

	return nil, c.baseURL, time.Since(start), fmt.Errorf("tf serving primary error: %w; fallback error: %v", err, fallbackErr)
}

// CheckHealth checks the model status via TensorFlow Serving REST API.
func (c *TFServingClient) CheckHealth(ctx context.Context) (string, error) {
	endpoint := c.baseURL
	if err := c.checkHealthOnce(ctx, c.baseURL); err == nil {
		return endpoint, nil
	}

	if c.fallbackURL == "" {
		return endpoint, fmt.Errorf("tf serving health check failed at %s", c.baseURL)
	}

	endpoint = c.fallbackURL
	if err := c.checkHealthOnce(ctx, c.fallbackURL); err == nil {
		return endpoint, nil
	}

	return endpoint, fmt.Errorf("tf serving health check failed at %s and fallback %s", c.baseURL, c.fallbackURL)
}

func (c *TFServingClient) checkHealthOnce(ctx context.Context, baseURL string) error {
	path := c.healthPath
	if path == "" {
		path = fmt.Sprintf("/v1/models/%s", c.modelName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tf serving health status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}

	if healthy, ok := parseTFServingHealth(body); ok && !healthy {
		return fmt.Errorf("tf serving model not available")
	}

	return nil
}

func (c *TFServingClient) predictOnce(ctx context.Context, baseURL string, features []float32) ([]float32, error) {
	requestURL := fmt.Sprintf("%s/v1/models/%s:predict", baseURL, c.modelName)

	var payload map[string]any
	if c.inputName != "" {
		payload = map[string]any{
			"instances": []map[string]any{
				{c.inputName: features},
			},
		}
	} else {
		payload = map[string]any{
			"instances": [][]float32{features},
		}
	}

	if c.signatureName != "" {
		payload["signature_name"] = c.signatureName
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tf serving request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build tf serving request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tf serving request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tf serving response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tf serving returned %d: %s", resp.StatusCode, string(respBody))
	}

	output, err := parseTFServingPrediction(respBody, c.outputName)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func parseTFServingPrediction(body []byte, outputName string) ([]float32, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode tf serving response: %w", err)
	}

	rawPredictions, ok := payload["predictions"]
	if !ok {
		rawPredictions = payload["outputs"]
	}
	predictions, ok := rawPredictions.([]any)
	if !ok || len(predictions) == 0 {
		return nil, fmt.Errorf("tf serving response missing predictions")
	}

	return parsePredictionValue(predictions[0], outputName)
}

func parsePredictionValue(value any, outputName string) ([]float32, error) {
	switch typed := value.(type) {
	case []any:
		if len(typed) == 0 {
			return nil, fmt.Errorf("empty prediction output")
		}
		if inner, ok := typed[0].([]any); ok {
			return parseFloatSlice(inner)
		}
		return parseFloatSlice(typed)
	case map[string]any:
		if outputName != "" {
			if inner, ok := typed[outputName]; ok {
				return parsePredictionValue(inner, "")
			}
		}
		for _, key := range []string{"outputs", "scores", "score"} {
			if inner, ok := typed[key]; ok {
				return parsePredictionValue(inner, "")
			}
		}
		return nil, fmt.Errorf("unknown prediction map format")
	case float64:
		return []float32{float32(typed)}, nil
	default:
		return nil, fmt.Errorf("unexpected prediction type %T", value)
	}
}

func parseFloatSlice(values []any) ([]float32, error) {
	result := make([]float32, 0, len(values))
	for _, raw := range values {
		val, ok := raw.(float64)
		if !ok {
			return nil, fmt.Errorf("prediction element not numeric: %T", raw)
		}
		result = append(result, float32(val))
	}
	return result, nil
}

func parseTFServingHealth(body []byte) (bool, bool) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, false
	}

	statuses, ok := payload["model_version_status"].([]any)
	if !ok || len(statuses) == 0 {
		return false, false
	}

	for _, status := range statuses {
		entry, ok := status.(map[string]any)
		if !ok {
			continue
		}
		state, _ := entry["state"].(string)
		if strings.EqualFold(state, "AVAILABLE") {
			return true, true
		}
	}

	return false, true
}
