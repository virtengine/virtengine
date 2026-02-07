package provider_daemon

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *PortalAPIServer) handleListTickets(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	status := r.URL.Query().Get("status")
	deploymentID := r.URL.Query().Get("deployment_id")
	tickets, err := s.chainQuery.ListTickets(r.Context(), principal, status, deploymentID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tickets)
}

func (s *PortalAPIServer) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DeploymentID == "" || req.Subject == "" || req.Description == "" {
		writeJSONError(w, http.StatusBadRequest, "deployment_id, subject, and description are required")
		return
	}

	ticket, err := s.chainQuery.CreateTicket(r.Context(), principal, req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (s *PortalAPIServer) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := mux.Vars(r)["ticketId"]
	if ticketID == "" {
		writeJSONError(w, http.StatusBadRequest, "ticket id required")
		return
	}

	ticket, err := s.chainQuery.GetTicket(r.Context(), ticketID)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (s *PortalAPIServer) handleAddTicketComment(w http.ResponseWriter, r *http.Request) {
	ticketID := mux.Vars(r)["ticketId"]
	if ticketID == "" {
		writeJSONError(w, http.StatusBadRequest, "ticket id required")
		return
	}

	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req TicketCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "message is required")
		return
	}

	comment, err := s.chainQuery.AddTicketComment(r.Context(), ticketID, principal, req.Message)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (s *PortalAPIServer) handleUpdateTicket(w http.ResponseWriter, r *http.Request) {
	ticketID := mux.Vars(r)["ticketId"]
	if ticketID == "" {
		writeJSONError(w, http.StatusBadRequest, "ticket id required")
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := s.chainQuery.UpdateTicket(r.Context(), ticketID, req)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}
