package provider_daemon

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/virtengine/virtengine/pkg/data_vault"
	"github.com/virtengine/virtengine/pkg/observability"
	portalauth "github.com/virtengine/virtengine/pkg/provider_daemon/auth"
)

const (
	shellCodeStdout  = 100
	shellCodeStderr  = 101
	shellCodeResult  = 102
	shellCodeFailure = 103
	shellCodeStdin   = 104
	shellCodeResize  = 105
)

type PortalAPIServerConfig struct {
	ListenAddr              string
	AuthSecret              string
	AllowInsecure           bool
	RequireVEID             bool
	MinVEIDScore            int
	ShellSessionTTL         time.Duration
	TokenTTL                time.Duration
	LifecycleExecutor       LifecycleExecutor
	LifecycleAllowedRoles   []string
	LifecycleConsentScope   string
	LifecycleRequireConsent bool
	AuditLogger             *AuditLogger
	LogStore                *DeploymentLogStore
	ChainQuery              ChainQuery
	WalletAuthChainID       string
	WalletAuthNonceStore    portalauth.NonceStore
	WalletAuthChainQuery    portalauth.ChainQuerier
	WalletAuthMaxAge        time.Duration
	WalletAuthFutureDrift   time.Duration
	WalletAuthCacheTTL      time.Duration
	ProviderInfo            ProviderInfo
	ProviderPricing         ProviderPricing
	ProviderCapacity        ProviderCapacity
	ProviderAttributes      ProviderAttributes
	ProviderInfoProvider    ProviderInfoProvider
	RateLimit               RateLimitConfig
	VaultService            data_vault.VaultService
	VaultMaxPayloadBytes    int64
}

func DefaultPortalAPIServerConfig() PortalAPIServerConfig {
	return PortalAPIServerConfig{
		ListenAddr:              ":8080",
		AllowInsecure:           true,
		RequireVEID:             true,
		MinVEIDScore:            80,
		ShellSessionTTL:         10 * time.Minute,
		TokenTTL:                5 * time.Minute,
		VaultMaxPayloadBytes:    10 * 1024 * 1024,
		LifecycleAllowedRoles:   []string{"customer", "administrator", "support_agent"},
		LifecycleConsentScope:   "marketplace:lifecycle",
		LifecycleRequireConsent: true,
		RateLimit: RateLimitConfig{
			RequestsPerMinute: 120,
		},
	}
}

type PortalAPIServer struct {
	cfg           PortalAPIServerConfig
	server        *http.Server
	logStore      *DeploymentLogStore
	shellSessions *ShellSessionManager
	upgrader      websocket.Upgrader
	chainQuery    ChainQuery
	providerInfo  ProviderInfoProvider
	rateLimiter   *PortalRateLimiter
	authVerifier  *portalauth.Verifier
	vault         data_vault.VaultService
	lifecycleExec LifecycleExecutor
}

func NewPortalAPIServer(cfg PortalAPIServerConfig) (*PortalAPIServer, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.ShellSessionTTL == 0 {
		cfg.ShellSessionTTL = 10 * time.Minute
	}
	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = 5 * time.Minute
	}
	if cfg.VaultMaxPayloadBytes == 0 {
		cfg.VaultMaxPayloadBytes = 10 * 1024 * 1024
	}
	if cfg.AuditLogger == nil {
		logger, err := NewAuditLogger(DefaultAuditLogConfig())
		if err != nil {
			return nil, err
		}
		cfg.AuditLogger = logger
	}
	if cfg.LogStore == nil {
		cfg.LogStore = NewDeploymentLogStore()
	}
	if cfg.ChainQuery == nil {
		cfg.ChainQuery = NoopChainQuery{}
	}
	if cfg.WalletAuthNonceStore == nil {
		cfg.WalletAuthNonceStore = portalauth.NewInMemoryNonceStore()
	}
	if cfg.WalletAuthChainQuery == nil {
		cfg.WalletAuthChainQuery = portalauth.NoopChainQuerier{}
	}

	srv := &PortalAPIServer{
		cfg:           cfg,
		logStore:      cfg.LogStore,
		shellSessions: NewShellSessionManager(cfg.TokenTTL),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		chainQuery: cfg.ChainQuery,
		authVerifier: portalauth.NewVerifier(portalauth.VerifierConfig{
			ChainID:           cfg.WalletAuthChainID,
			NonceStore:        cfg.WalletAuthNonceStore,
			ChainQuerier:      cfg.WalletAuthChainQuery,
			MaxTimestampAge:   cfg.WalletAuthMaxAge,
			FutureTimeDrift:   cfg.WalletAuthFutureDrift,
			OwnershipCacheTTL: cfg.WalletAuthCacheTTL,
		}),
	}

	if cfg.ProviderInfoProvider != nil {
		srv.providerInfo = cfg.ProviderInfoProvider
	} else {
		srv.providerInfo = NewStaticProviderInfoProvider(cfg.ProviderInfo, cfg.ProviderPricing, cfg.ProviderCapacity, cfg.ProviderAttributes)
	}

	srv.vault = cfg.VaultService
	srv.rateLimiter = NewPortalRateLimiter(cfg.RateLimit.RequestsPerMinute, time.Minute)
	srv.lifecycleExec = cfg.LifecycleExecutor

	return srv, nil
}

func (s *PortalAPIServer) Start(ctx context.Context) error {
	router := mux.NewRouter()
	s.setupRoutes(router)

	s.server = &http.Server{
		Addr:              s.cfg.ListenAddr,
		Handler:           observability.HTTPTracingHandler(router, "provider.portal_api"),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()

	return s.server.ListenAndServe()
}

func (s *PortalAPIServer) setupRoutes(router *mux.Router) {
	router.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)
	router.HandleFunc("/deployments/{id}/logs", s.handleLogs).Methods(http.MethodGet)
	router.HandleFunc("/deployments/{id}/shell/session", s.handleShellSession).Methods(http.MethodPost)
	router.HandleFunc("/deployments/{id}/shell", s.handleShell).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.rateLimitMiddleware())

	api.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)
	api.HandleFunc("/deployments/{deploymentId}/logs", s.handleLogs).Methods(http.MethodGet)
	api.HandleFunc("/deployments/{deploymentId}/shell/session", s.handleShellSession).Methods(http.MethodPost)
	api.HandleFunc("/deployments/{deploymentId}/shell", s.handleShell).Methods(http.MethodGet)

	api.Handle("/organizations", s.authMiddleware(true)(http.HandlerFunc(s.handleListOrganizations))).Methods(http.MethodGet)
	api.Handle("/organizations/{orgId}", s.authMiddleware(true)(http.HandlerFunc(s.handleGetOrganization))).Methods(http.MethodGet)
	api.Handle("/organizations/{orgId}/members", s.authMiddleware(true)(http.HandlerFunc(s.handleOrganizationMembers))).Methods(http.MethodGet)
	api.Handle("/organizations/{orgId}/invite", s.authMiddleware(true)(http.HandlerFunc(s.handleInviteOrganizationMember))).Methods(http.MethodPost)
	api.Handle("/organizations/{orgId}/members/{address}", s.authMiddleware(true)(http.HandlerFunc(s.handleRemoveOrganizationMember))).Methods(http.MethodDelete)

	api.Handle("/tickets", s.authMiddleware(true)(http.HandlerFunc(s.handleListTickets))).Methods(http.MethodGet)
	api.Handle("/tickets", s.authMiddleware(true)(http.HandlerFunc(s.handleCreateTicket))).Methods(http.MethodPost)
	api.Handle("/tickets/{ticketId}", s.authMiddleware(true)(http.HandlerFunc(s.handleGetTicket))).Methods(http.MethodGet)
	api.Handle("/tickets/{ticketId}/comments", s.authMiddleware(true)(http.HandlerFunc(s.handleAddTicketComment))).Methods(http.MethodPost)
	api.Handle("/tickets/{ticketId}", s.authMiddleware(true)(http.HandlerFunc(s.handleUpdateTicket))).Methods(http.MethodPatch)

	api.Handle("/invoices", s.authMiddleware(true)(http.HandlerFunc(s.handleListInvoices))).Methods(http.MethodGet)
	api.Handle("/invoices/{invoiceId}", s.authMiddleware(true)(http.HandlerFunc(s.handleGetInvoice))).Methods(http.MethodGet)
	api.Handle("/usage", s.authMiddleware(true)(http.HandlerFunc(s.handleGetUsage))).Methods(http.MethodGet)
	api.Handle("/usage/history", s.authMiddleware(true)(http.HandlerFunc(s.handleGetUsageHistory))).Methods(http.MethodGet)

	api.Handle("/deployments/{deploymentId}/metrics", s.authMiddleware(true)(s.leaseOwnerMiddleware()(http.HandlerFunc(s.handleDeploymentMetrics)))).Methods(http.MethodGet)
	api.Handle("/deployments/{deploymentId}/metrics/history", s.authMiddleware(true)(s.leaseOwnerMiddleware()(http.HandlerFunc(s.handleDeploymentMetricsHistory)))).Methods(http.MethodGet)
	api.Handle("/deployments/{deploymentId}/events", s.authMiddleware(true)(s.leaseOwnerMiddleware()(http.HandlerFunc(s.handleDeploymentEvents)))).Methods(http.MethodGet)
	api.Handle("/deployments/{deploymentId}/actions", s.authMiddleware(true)(s.leaseOwnerMiddleware()(http.HandlerFunc(s.handleDeploymentAction)))).Methods(http.MethodPost)
	api.Handle("/metrics/aggregate", s.authMiddleware(true)(http.HandlerFunc(s.handleAggregatedMetrics))).Methods(http.MethodGet)

	api.Handle("/provider/info", s.authMiddleware(false)(http.HandlerFunc(s.handleProviderInfo))).Methods(http.MethodGet)
	api.Handle("/provider/pricing", s.authMiddleware(false)(http.HandlerFunc(s.handleProviderPricing))).Methods(http.MethodGet)
	api.Handle("/provider/capacity", s.authMiddleware(false)(http.HandlerFunc(s.handleProviderCapacity))).Methods(http.MethodGet)
	api.Handle("/provider/attributes", s.authMiddleware(false)(http.HandlerFunc(s.handleProviderAttributes))).Methods(http.MethodGet)

	if s.vault != nil {
		api.Handle("/vault/blobs", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultUpload))).Methods(http.MethodPost)
		api.Handle("/vault/blobs/{blobId}", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultRetrieve))).Methods(http.MethodGet)
		api.Handle("/vault/blobs/{blobId}/metadata", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultMetadata))).Methods(http.MethodGet)
		api.Handle("/vault/blobs/{blobId}", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultDelete))).Methods(http.MethodDelete)
		api.Handle("/vault/keys", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultKeyList))).Methods(http.MethodGet)
		api.Handle("/vault/keys/{keyId}", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultKeyMetadata))).Methods(http.MethodGet)
		api.Handle("/vault/audit", s.authMiddleware(true)(http.HandlerFunc(s.handleVaultAuditSearch))).Methods(http.MethodGet)
	}
}

func (s *PortalAPIServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *PortalAPIServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *PortalAPIServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	deploymentID := deploymentIDFromVars(r)

	authCtx, authErr := s.authenticateRequest(r)
	if authErr != nil {
		s.auditDenied(authCtx.Address, deploymentID, "logs", authErr)
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	if err := s.verifyLeaseOwnership(r.Context(), authCtx, deploymentID); err != nil {
		s.auditDenied(authCtx.Address, deploymentID, "logs", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if websocket.IsWebSocketUpgrade(r) {
		s.handleLogStreamWS(w, r, deploymentID, authCtx.Address)
		return
	}

	tail := parseIntQuery(r, "tail", 200)
	levelFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("level")))
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	lines := s.logStore.Tail(deploymentID, tail)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"deployment-%s-logs.txt\"", deploymentID))
	}

	for _, entry := range lines {
		if levelFilter != "" && strings.ToLower(entry.Level) != levelFilter {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(entry.Message), search) {
			continue
		}
		_, _ = fmt.Fprintf(w, "%s %s %s\n", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
	}
}

func (s *PortalAPIServer) handleLogStreamWS(w http.ResponseWriter, r *http.Request, deploymentID, principal string) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	levelFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("level")))
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	tail := parseIntQuery(r, "tail", 200)
	follow := r.URL.Query().Get("follow") != "0"

	_ = s.cfg.AuditLogger.Log(&AuditEvent{
		Type:        AuditEventLogStreamStarted,
		Operation:   "log_stream",
		Success:     true,
		PrincipalID: principal,
		Details: map[string]interface{}{
			"deployment_id": deploymentID,
		},
	})

	defer func() {
		_ = s.cfg.AuditLogger.Log(&AuditEvent{
			Type:        AuditEventLogStreamEnded,
			Operation:   "log_stream",
			Success:     true,
			PrincipalID: principal,
			Details: map[string]interface{}{
				"deployment_id": deploymentID,
			},
		})
	}()

	lines := s.logStore.Tail(deploymentID, tail)
	for _, entry := range lines {
		if !shouldSendLog(entry, levelFilter, search) {
			continue
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(formatLogEntry(entry)))
	}

	if !follow {
		return
	}

	ch, cancel := s.logStore.Subscribe(deploymentID)
	defer cancel()

	for entry := range ch {
		if !shouldSendLog(entry, levelFilter, search) {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(formatLogEntry(entry))); err != nil {
			return
		}
	}
}

func (s *PortalAPIServer) handleShellSession(w http.ResponseWriter, r *http.Request) {
	deploymentID := deploymentIDFromVars(r)
	authCtx, authErr := s.authenticateRequest(r)
	if authErr != nil {
		s.auditDenied(authCtx.Address, deploymentID, "shell_session", authErr)
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	if err := s.verifyLeaseOwnership(r.Context(), authCtx, deploymentID); err != nil {
		s.auditDenied(authCtx.Address, deploymentID, "shell_session", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if s.cfg.RequireVEID {
		if err := s.verifyVEID(r); err != nil {
			s.auditDenied(authCtx.Address, deploymentID, "shell_session", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	var req struct {
		Container string `json:"container"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	session, err := s.shellSessions.Issue(deploymentID, authCtx.Address, req.Container, s.cfg.ShellSessionTTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = s.cfg.AuditLogger.Log(&AuditEvent{
		Type:        AuditEventShellSessionCreated,
		Operation:   "shell_session",
		Success:     true,
		PrincipalID: authCtx.Address,
		Details: map[string]interface{}{
			"deployment_id": deploymentID,
			"container":     req.Container,
		},
	})

	resp := map[string]interface{}{
		"token":       session.Token,
		"expires_at":  session.ExpiresAt.Format(time.RFC3339),
		"deployment":  deploymentID,
		"container":   req.Container,
		"session_ttl": int(s.cfg.ShellSessionTTL.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("[portal-api] shell session encode error: %v\n", err)
	}
}

func (s *PortalAPIServer) handleShell(w http.ResponseWriter, r *http.Request) {
	deploymentID := deploymentIDFromVars(r)
	authCtx, authErr := s.authenticateRequest(r)
	if authErr != nil {
		s.auditDenied(authCtx.Address, deploymentID, "shell", authErr)
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	if err := s.verifyLeaseOwnership(r.Context(), authCtx, deploymentID); err != nil {
		s.auditDenied(authCtx.Address, deploymentID, "shell", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if s.cfg.RequireVEID {
		if err := s.verifyVEID(r); err != nil {
			s.auditDenied(authCtx.Address, deploymentID, "shell", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	sessionToken := r.URL.Query().Get("token")
	if sessionToken == "" && !s.cfg.AllowInsecure {
		s.auditDenied(authCtx.Address, deploymentID, "shell", errors.New("missing session token"))
		http.Error(w, "missing session token", http.StatusUnauthorized)
		return
	}

	var session *ShellSession
	if sessionToken != "" {
		var err error
		session, err = s.shellSessions.Validate(sessionToken, deploymentID, authCtx.Address)
		if err != nil {
			s.auditDenied(authCtx.Address, deploymentID, "shell", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = s.cfg.AuditLogger.Log(&AuditEvent{
		Type:        AuditEventShellSessionStarted,
		Operation:   "shell",
		Success:     true,
		PrincipalID: authCtx.Address,
		Details: map[string]interface{}{
			"deployment_id": deploymentID,
			"container":     r.URL.Query().Get("container"),
		},
	})

	if session != nil {
		s.shellSessions.Activate(session.Token)
	}

	defer func() {
		if session != nil {
			s.shellSessions.Deactivate(session.Token)
		}
		_ = s.cfg.AuditLogger.Log(&AuditEvent{
			Type:        AuditEventShellSessionEnded,
			Operation:   "shell",
			Success:     true,
			PrincipalID: authCtx.Address,
			Details: map[string]interface{}{
				"deployment_id": deploymentID,
			},
		})
	}()

	writeShellMessage(conn, shellCodeStdout, []byte("Connected to VirtEngine shell.\r\n"))

	expired := make(chan struct{})
	if session != nil {
		go func() {
			timer := time.NewTimer(time.Until(session.ExpiresAt))
			defer timer.Stop()
			select {
			case <-timer.C:
				close(expired)
			case <-r.Context().Done():
			}
		}()
	}

	for {
		select {
		case <-expired:
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "session expired"),
				time.Now().Add(time.Second),
			)
			return
		default:
		}

		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.BinaryMessage || len(data) == 0 {
			continue
		}

		switch data[0] {
		case shellCodeStdin:
			payload := data[1:]
			if len(payload) == 0 {
				continue
			}
			writeShellMessage(conn, shellCodeStdout, payload)
		case shellCodeResize:
			// Terminal resize, ignore for now.
		default:
		}
	}
}

func (s *PortalAPIServer) authenticateRequest(r *http.Request) (portalauth.AuthContext, error) {
	if portalauth.HasWalletAuth(r) {
		if s.authVerifier == nil {
			return portalauth.AuthContext{}, errors.New("wallet auth not configured")
		}
		signed, err := s.authVerifier.Verify(r)
		if err != nil {
			return portalauth.AuthContext{}, err
		}
		return portalauth.AuthContext{Address: signed.Address, Kind: portalauth.AuthKindWallet}, nil
	}

	if s.cfg.AllowInsecure && s.cfg.AuthSecret == "" {
		return portalauth.AuthContext{Address: "dev", Kind: portalauth.AuthKindInsecure}, nil
	}

	signature := r.Header.Get("X-VE-Signature")
	if signature == "" {
		signature = r.URL.Query().Get("sig")
	}
	timestamp := r.Header.Get("X-VE-Timestamp")
	if timestamp == "" {
		timestamp = r.URL.Query().Get("ts")
	}
	principal := r.Header.Get("X-VE-Principal")
	if principal == "" {
		principal = r.URL.Query().Get("principal")
	}

	if signature == "" || timestamp == "" || principal == "" {
		if s.cfg.AllowInsecure {
			return portalauth.AuthContext{Address: "dev", Kind: portalauth.AuthKindInsecure}, nil
		}
		return portalauth.AuthContext{}, errors.New("missing signed request headers")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return portalauth.AuthContext{}, errors.New("invalid timestamp")
	}

	now := time.Now().Unix()
	if ts < now-300 || ts > now+300 {
		return portalauth.AuthContext{}, errors.New("request timestamp out of range")
	}

	if s.cfg.AuthSecret == "" {
		return portalauth.AuthContext{}, errors.New("hmac secret not configured")
	}

	payload := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		principal,
		timestamp,
	)
	expected := computeHMAC(payload, s.cfg.AuthSecret)
	if !strings.EqualFold(expected, signature) {
		return portalauth.AuthContext{}, errors.New("invalid signature")
	}

	return portalauth.AuthContext{Address: principal, Kind: portalauth.AuthKindHMAC}, nil
}

func (s *PortalAPIServer) authMiddleware(required bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !required && !portalauth.HasWalletAuth(r) && r.Header.Get("X-VE-Principal") == "" && r.URL.Query().Get("principal") == "" {
				next.ServeHTTP(w, r)
				return
			}

			authCtx, err := s.authenticateRequest(r)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, err.Error())
				return
			}

			ctx := withAuth(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *PortalAPIServer) leaseOwnerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := authFromContext(r.Context())
			if authCtx.Address == "" {
				writeJSONError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			leaseID := deploymentIDFromVars(r)
			if leaseID == "" {
				writeJSONError(w, http.StatusBadRequest, "deployment id required")
				return
			}

			if err := s.verifyLeaseOwnership(r.Context(), authCtx, leaseID); err != nil {
				writeJSONError(w, http.StatusForbidden, err.Error())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (s *PortalAPIServer) verifyLeaseOwnership(ctx context.Context, authCtx portalauth.AuthContext, leaseID string) error {
	if authCtx.Kind == portalauth.AuthKindInsecure {
		return nil
	}
	if authCtx.Address == "" {
		return errors.New("authentication required")
	}
	if s.authVerifier == nil {
		return errors.New("lease ownership verification unavailable")
	}
	return s.authVerifier.VerifyLeaseOwnership(ctx, authCtx.Address, leaseID)
}

func (s *PortalAPIServer) verifyVEID(r *http.Request) error {
	rawScore := r.Header.Get("X-VEID-Score")
	if rawScore == "" {
		rawScore = r.URL.Query().Get("veid_score")
	}
	if rawScore == "" {
		return errors.New("missing VEID score")
	}
	score, err := strconv.Atoi(rawScore)
	if err != nil {
		return errors.New("invalid VEID score")
	}
	if score < s.cfg.MinVEIDScore {
		return fmt.Errorf("VEID score %d below minimum %d", score, s.cfg.MinVEIDScore)
	}
	return nil
}

func (s *PortalAPIServer) auditDenied(principal, deploymentID, operation string, err error) {
	_ = s.cfg.AuditLogger.Log(&AuditEvent{
		Type:         AuditEventShellAccessDenied,
		Operation:    operation,
		Success:      false,
		PrincipalID:  principal,
		ErrorMessage: err.Error(),
		Details: map[string]interface{}{
			"deployment_id": deploymentID,
		},
	})
}

func computeHMAC(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	if val := r.URL.Query().Get(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func shouldSendLog(entry LogEntry, levelFilter, search string) bool {
	if levelFilter != "" && strings.ToLower(entry.Level) != levelFilter {
		return false
	}
	if search != "" && !strings.Contains(strings.ToLower(entry.Message), search) {
		return false
	}
	return true
}

func formatLogEntry(entry LogEntry) string {
	return fmt.Sprintf("%s %s %s",
		entry.Timestamp.Format(time.RFC3339),
		strings.ToUpper(entry.Level),
		entry.Message,
	)
}

func deploymentIDFromVars(r *http.Request) string {
	vars := mux.Vars(r)
	if vars == nil {
		return ""
	}
	if id := vars["deploymentId"]; id != "" {
		return id
	}
	return vars["id"]
}

func writeShellMessage(conn *websocket.Conn, messageType byte, payload []byte) {
	buf := make([]byte, 1+len(payload))
	buf[0] = messageType
	copy(buf[1:], payload)
	_ = conn.WriteMessage(websocket.BinaryMessage, buf)
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

type DeploymentLogStore struct {
	mu    sync.RWMutex
	logs  map[string][]LogEntry
	subs  map[string]map[chan LogEntry]struct{}
	limit int
}

func NewDeploymentLogStore() *DeploymentLogStore {
	return &DeploymentLogStore{
		logs:  make(map[string][]LogEntry),
		subs:  make(map[string]map[chan LogEntry]struct{}),
		limit: 2000,
	}
}

func (s *DeploymentLogStore) Append(deploymentID string, entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	s.logs[deploymentID] = append(s.logs[deploymentID], entry)
	if len(s.logs[deploymentID]) > s.limit {
		s.logs[deploymentID] = s.logs[deploymentID][len(s.logs[deploymentID])-s.limit:]
	}

	for ch := range s.subs[deploymentID] {
		select {
		case ch <- entry:
		default:
		}
	}
}

func (s *DeploymentLogStore) Tail(deploymentID string, tail int) []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lines := s.logs[deploymentID]
	if tail <= 0 || tail >= len(lines) {
		return append([]LogEntry(nil), lines...)
	}
	return append([]LogEntry(nil), lines[len(lines)-tail:]...)
}

func (s *DeploymentLogStore) Subscribe(deploymentID string) (<-chan LogEntry, func()) {
	ch := make(chan LogEntry, 50)

	s.mu.Lock()
	if s.subs[deploymentID] == nil {
		s.subs[deploymentID] = make(map[chan LogEntry]struct{})
	}
	s.subs[deploymentID][ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		if subs, ok := s.subs[deploymentID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(s.subs, deploymentID)
			}
		}
		s.mu.Unlock()
		close(ch)
	}

	return ch, cancel
}

type ShellSession struct {
	Token        string
	DeploymentID string
	Container    string
	PrincipalID  string
	ExpiresAt    time.Time
	Active       bool
}

type ShellSessionManager struct {
	mu        sync.RWMutex
	ttl       time.Duration
	sessions  map[string]*ShellSession
	principal map[string]map[string]struct{}
}

func NewShellSessionManager(ttl time.Duration) *ShellSessionManager {
	return &ShellSessionManager{
		ttl:       ttl,
		sessions:  make(map[string]*ShellSession),
		principal: make(map[string]map[string]struct{}),
	}
}

func (m *ShellSessionManager) Issue(deploymentID, principalID, container string, ttl time.Duration) (*ShellSession, error) {
	if ttl <= 0 {
		ttl = m.ttl
	}
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}

	session := &ShellSession{
		Token:        token,
		DeploymentID: deploymentID,
		Container:    container,
		PrincipalID:  principalID,
		ExpiresAt:    time.Now().Add(ttl),
	}

	m.mu.Lock()
	m.sessions[token] = session
	if m.principal[principalID] == nil {
		m.principal[principalID] = make(map[string]struct{})
	}
	m.principal[principalID][token] = struct{}{}
	m.mu.Unlock()

	return session, nil
}

func (m *ShellSessionManager) Validate(token, deploymentID, principalID string) (*ShellSession, error) {
	m.mu.RLock()
	session, ok := m.sessions[token]
	m.mu.RUnlock()

	if !ok {
		return nil, errors.New("invalid session token")
	}
	if session.PrincipalID != principalID {
		return nil, errors.New("session principal mismatch")
	}
	if session.DeploymentID != deploymentID {
		return nil, errors.New("session deployment mismatch")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}
	return session, nil
}

func (m *ShellSessionManager) Activate(token string) {
	m.mu.Lock()
	if session, ok := m.sessions[token]; ok {
		session.Active = true
	}
	m.mu.Unlock()
}

func (m *ShellSessionManager) Deactivate(token string) {
	m.mu.Lock()
	if session, ok := m.sessions[token]; ok {
		session.Active = false
	}
	m.mu.Unlock()
}

func generateToken(length int) (string, error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
