// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/payments/offramp"
)

const (
	defaultFiatWebhookBodyLimit = int64(1 << 20)
	defaultFiatWebhookTimeout   = 5 * time.Second
)

// FiatWebhookEventRepository persists verified callbacks before acknowledgement.
type FiatWebhookEventRepository interface {
	PutVerifiedWebhookEvent(context.Context, offramp.WebhookEvent) error
	Durable() bool
}

// FiatWebhookHandlerConfig is intentionally separate from a public listener;
// callers must mount it on a private/authenticated ingress path.
type FiatWebhookHandlerConfig struct {
	Verifier     *offramp.WebhookVerifier
	Events       FiatWebhookEventRepository
	Orchestrator *FiatConversionOrchestrator
	MaxBodyBytes int64
	Timeout      time.Duration
}

type FiatWebhookHandler struct {
	verifier     *offramp.WebhookVerifier
	events       FiatWebhookEventRepository
	orchestrator *FiatConversionOrchestrator
	maxBodyBytes int64
	timeout      time.Duration
}

func NewFiatWebhookHandler(cfg FiatWebhookHandlerConfig) (*FiatWebhookHandler, error) {
	if cfg.Verifier == nil || cfg.Events == nil || !cfg.Events.Durable() || cfg.Orchestrator == nil {
		return nil, fmt.Errorf("webhook verifier, durable event store, and orchestrator are required")
	}
	limit := cfg.MaxBodyBytes
	if limit <= 0 {
		limit = defaultFiatWebhookBodyLimit
	}
	if limit > 8<<20 {
		return nil, errors.New("webhook body limit exceeds safety bound")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultFiatWebhookTimeout
	}
	if timeout > 30*time.Second {
		return nil, errors.New("webhook timeout exceeds safety bound")
	}
	return &FiatWebhookHandler{verifier: cfg.Verifier, events: cfg.Events, orchestrator: cfg.Orchestrator, maxBodyBytes: limit, timeout: timeout}, nil
}

func (h *FiatWebhookHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if request.Host == "" || request.URL == nil || request.URL.User != nil || request.URL.RawQuery != "" || request.URL.Fragment != "" ||
		request.Header.Get("Transfer-Encoding") != "" || len(request.TransferEncoding) != 0 || request.ContentLength < 0 || request.ContentLength > h.maxBodyBytes ||
		!strings.HasPrefix(strings.ToLower(request.Header.Get("Content-Type")), "application/json") || !singleFiatWebhookHeaders(request.Header) {
		writeWebhookError(writer, http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), h.timeout)
	defer cancel()
	limited := http.MaxBytesReader(writer, request.Body, h.maxBodyBytes)
	defer limited.Close()
	verified, err := h.verifier.Verify(ctx, offramp.WebhookHeaders{Timestamp: strings.TrimSpace(request.Header.Get("X-Offramp-Timestamp")), Signature: strings.TrimSpace(request.Header.Get("X-Offramp-Signature")), KeyID: strings.TrimSpace(request.Header.Get("X-Offramp-Key-ID")), KeyVersion: strings.TrimSpace(request.Header.Get("X-Offramp-Key-Version")), APIVersion: strings.TrimSpace(request.Header.Get("X-Offramp-API-Version"))}, limited)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, offramp.ErrAdapterUnavailable) {
			status = http.StatusServiceUnavailable
		}
		writeWebhookError(writer, status)
		return
	}
	if err := h.events.PutVerifiedWebhookEvent(ctx, verified.Event); err != nil {
		writeWebhookError(writer, http.StatusServiceUnavailable)
		return
	}
	h.orchestrator.Wake()
	writer.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(writer).Encode(struct {
		Accepted  bool `json:"accepted"`
		Duplicate bool `json:"duplicate"`
	}{Accepted: true, Duplicate: verified.Duplicate}); err != nil {
		return
	}
}

func singleFiatWebhookHeaders(header http.Header) bool {
	for _, name := range []string{"X-Offramp-Signature", "X-Offramp-Timestamp", "X-Offramp-Key-ID", "X-Offramp-Key-Version", "X-Offramp-API-Version"} {
		values := header.Values(name)
		if len(values) != 1 || strings.TrimSpace(values[0]) == "" || values[0] != strings.TrimSpace(values[0]) || strings.ContainsAny(values[0], "\r\n,;") {
			return false
		}
	}
	return true
}

func writeWebhookError(writer http.ResponseWriter, status int) {
	writer.WriteHeader(status)
	_, _ = io.WriteString(writer, "{\"accepted\":false}\n")
}

var _ http.Handler = (*FiatWebhookHandler)(nil)
