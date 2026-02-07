package provider_daemon

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *PortalAPIServer) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)
	orgs, _, err := s.chainQuery.ListOrganizations(r.Context(), principal, limit, cursor)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, orgs)
}

func (s *PortalAPIServer) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if orgID == "" {
		writeJSONError(w, http.StatusBadRequest, "organization id required")
		return
	}

	org, err := s.chainQuery.GetOrganization(r.Context(), orgID)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "organization not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, org)
}

func (s *PortalAPIServer) handleOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if orgID == "" {
		writeJSONError(w, http.StatusBadRequest, "organization id required")
		return
	}

	members, err := s.chainQuery.ListOrganizationMembers(r.Context(), orgID)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "organization not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (s *PortalAPIServer) handleInviteOrganizationMember(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if orgID == "" {
		writeJSONError(w, http.StatusBadRequest, "organization id required")
		return
	}

	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	isAdmin, err := s.chainQuery.IsOrganizationAdmin(r.Context(), orgID, principal)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isAdmin {
		writeJSONError(w, http.StatusForbidden, "not authorized to invite members")
		return
	}

	var req OrganizationInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Address == "" || req.Role == "" {
		writeJSONError(w, http.StatusBadRequest, "address and role are required")
		return
	}

	member, err := s.chainQuery.InviteOrganizationMember(r.Context(), orgID, req.Address, req.Role, principal)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, member)
}

func (s *PortalAPIServer) handleRemoveOrganizationMember(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	address := mux.Vars(r)["address"]
	if orgID == "" || address == "" {
		writeJSONError(w, http.StatusBadRequest, "organization id and address required")
		return
	}

	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	isAdmin, err := s.chainQuery.IsOrganizationAdmin(r.Context(), orgID, principal)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isAdmin {
		writeJSONError(w, http.StatusForbidden, "not authorized to remove members")
		return
	}

	if err := s.chainQuery.RemoveOrganizationMember(r.Context(), orgID, address, principal); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
