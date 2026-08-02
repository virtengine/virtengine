package provider_daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	portalauth "github.com/virtengine/virtengine/pkg/provider_daemon/auth"
)

func TestUnavailableOrganizationRoutesReturnTypedFeatureUnavailable(t *testing.T) {
	server := &PortalAPIServer{chainQuery: NoopChainQuery{}}
	tests := []struct {
		name    string
		method  string
		body    string
		vars    map[string]string
		handler http.HandlerFunc
	}{
		{name: "list", method: http.MethodGet, handler: server.handleListOrganizations},
		{name: "get", method: http.MethodGet, vars: map[string]string{"orgId": "org-1"}, handler: server.handleGetOrganization},
		{name: "members", method: http.MethodGet, vars: map[string]string{"orgId": "org-1"}, handler: server.handleOrganizationMembers},
		{name: "invite", method: http.MethodPost, body: `{"address":"member","role":"member"}`, vars: map[string]string{"orgId": "org-1"}, handler: server.handleInviteOrganizationMember},
		{name: "remove", method: http.MethodDelete, vars: map[string]string{"orgId": "org-1", "address": "member"}, handler: server.handleRemoveOrganizationMember},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertUnavailableRoute(t, test.method, "/", test.body, test.vars, PortalCapabilityOrganizations, "89C", test.handler)
		})
	}
}

func TestUnavailableEnabledDataRoutesReturnTypedFeatureUnavailable(t *testing.T) {
	server := &PortalAPIServer{chainQuery: NoopChainQuery{}}
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		vars       map[string]string
		capability PortalRouteCapability
		handler    http.HandlerFunc
	}{
		{name: "list tickets", method: http.MethodGet, capability: PortalCapabilityTickets, handler: server.handleListTickets},
		{name: "create ticket", method: http.MethodPost, body: `{"deployment_id":"deployment-1","subject":"subject","description":"description"}`, capability: PortalCapabilityTickets, handler: server.handleCreateTicket},
		{name: "get ticket", method: http.MethodGet, vars: map[string]string{"ticketId": "ticket-1"}, capability: PortalCapabilityTickets, handler: server.handleGetTicket},
		{name: "add ticket comment", method: http.MethodPost, body: `{"message":"message"}`, vars: map[string]string{"ticketId": "ticket-1"}, capability: PortalCapabilityTickets, handler: server.handleAddTicketComment},
		{name: "update ticket", method: http.MethodPatch, body: `{"status":"closed"}`, vars: map[string]string{"ticketId": "ticket-1"}, capability: PortalCapabilityTickets, handler: server.handleUpdateTicket},
		{name: "list invoices", method: http.MethodGet, capability: PortalCapabilityBilling, handler: server.handleListInvoices},
		{name: "get invoice", method: http.MethodGet, vars: map[string]string{"invoiceId": "invoice-1"}, capability: PortalCapabilityBilling, handler: server.handleGetInvoice},
		{name: "usage summary", method: http.MethodGet, capability: PortalCapabilityUsage, handler: server.handleGetUsage},
		{name: "usage history", method: http.MethodGet, path: "/?start=2026-01-01T00:00:00Z&end=2026-01-02T00:00:00Z", capability: PortalCapabilityUsage, handler: server.handleGetUsageHistory},
		{name: "deployment events", method: http.MethodGet, vars: map[string]string{"deploymentId": "deployment-1"}, capability: PortalCapabilityEvents, handler: server.handleDeploymentEvents},
		{name: "deployment metrics", method: http.MethodGet, vars: map[string]string{"deploymentId": "deployment-1"}, capability: PortalCapabilityMetrics, handler: server.handleDeploymentMetrics},
		{name: "metrics history", method: http.MethodGet, path: "/?start=2026-01-01T00:00:00Z&end=2026-01-02T00:00:00Z", vars: map[string]string{"deploymentId": "deployment-1"}, capability: PortalCapabilityMetrics, handler: server.handleDeploymentMetricsHistory},
		{name: "aggregate metrics", method: http.MethodGet, path: "/?start=2026-01-01T00:00:00Z&end=2026-01-02T00:00:00Z", capability: PortalCapabilityMetrics, handler: server.handleAggregatedMetrics},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertUnavailableRoute(t, test.method, test.path, test.body, test.vars, test.capability, "86C", test.handler)
		})
	}
}

func assertUnavailableRoute(t *testing.T, method, path, body string, vars map[string]string, capability PortalRouteCapability, owner string, handler http.HandlerFunc) {
	t.Helper()
	if path == "" {
		path = "/"
	}
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request = request.WithContext(withAuth(request.Context(), portalauth.AuthContext{Address: "customer"}))
	request = mux.SetURLVars(request, vars)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	var payload map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, "feature_unavailable", payload["code"])
	require.Equal(t, string(capability), payload["capability"])
	require.Equal(t, owner, payload["owner"])
	require.NotEmpty(t, payload["error"])
}
