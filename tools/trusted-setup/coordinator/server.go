package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Server exposes coordinator operations over HTTP.
type Server struct {
	State State
}

// NewServer creates a coordinator server bound to the ceremony state directory.
func NewServer(state State) *Server {
	return &Server{State: state}
}

// Listen runs the coordinator HTTP server until it stops.
func (s *Server) Listen(addr string) error {
	return s.Run(context.Background(), addr)
}

// Run starts the HTTP server and blocks until it exits or the context is canceled.
func (s *Server) Run(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/transcript", s.handleTranscript)
	mux.HandleFunc("/api/v1/phase1/current", s.handlePhase1Current)
	mux.HandleFunc("/api/v1/phase1/contribute", s.handlePhase1Contribute)
	mux.HandleFunc("/api/v1/phase2/current", s.handlePhase2Current)
	mux.HandleFunc("/api/v1/phase2/contribute", s.handlePhase2Contribute)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := StatusSnapshot(s.State)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, status)
}

func (s *Server) handleTranscript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	data, err := s.State.LoadTranscript()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (s *Server) handlePhase1Current(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_, data, err := loadLatestPhase1(s.State)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func (s *Server) handlePhase2Current(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_, data, err := loadLatestPhase2(s.State)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func (s *Server) handlePhase1Contribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	payload, err := readLimited(r.Body, 128<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	meta := extractMeta(r)
	if err := AcceptPhase1Contribution(s.State, payload, meta); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handlePhase2Contribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	payload, err := readLimited(r.Body, 256<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	meta := extractMeta(r)
	if err := AcceptPhase2Contribution(s.State, payload, meta); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func extractMeta(r *http.Request) ContributionMeta {
	return ContributionMeta{
		ParticipantID: r.Header.Get("X-Participant-Id"),
		PublicKey:     r.Header.Get("X-Public-Key"),
		Signature:     r.Header.Get("X-Signature"),
		Attestation:   r.Header.Get("X-Attestation"),
	}
}

func readLimited(reader io.Reader, max int64) ([]byte, error) {
	limited := io.LimitReader(reader, max)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) >= max {
		return nil, fmt.Errorf("payload exceeds limit")
	}
	return data, nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
