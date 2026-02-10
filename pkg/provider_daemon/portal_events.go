package provider_daemon

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *PortalAPIServer) handleDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	deploymentID := mux.Vars(r)["deploymentId"]
	if deploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "deployment id required")
		return
	}

	limit := parseLimit(r, 20, 100)
	cursor := parseCursor(r)
	events, nextCursor, err := s.chainQuery.GetDeploymentEvents(r.Context(), deploymentID, limit, cursor)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := DeploymentEventListResponse{
		Events:     events,
		NextCursor: nextCursor,
	}
	writeJSON(w, http.StatusOK, resp)
}
