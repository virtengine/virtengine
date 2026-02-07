package provider_daemon

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (s *PortalAPIServer) handleListInvoices(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)
	status := r.URL.Query().Get("status")

	invoices, nextCursor, err := s.chainQuery.ListInvoices(r.Context(), principal, status, limit, cursor)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := InvoiceListResponse{
		Invoices:   invoices,
		NextCursor: nextCursor,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *PortalAPIServer) handleGetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := mux.Vars(r)["invoiceId"]
	if invoiceID == "" {
		writeJSONError(w, http.StatusBadRequest, "invoice id required")
		return
	}

	invoice, err := s.chainQuery.GetInvoice(r.Context(), invoiceID)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, invoice)
}

func (s *PortalAPIServer) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	summary, err := s.chainQuery.GetUsageSummary(r.Context(), principal)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (s *PortalAPIServer) handleGetUsageHistory(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	interval, err := parseInterval(r, time.Hour)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	history, err := s.chainQuery.GetUsageHistory(r.Context(), principal, start, end, interval)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, history)
}
