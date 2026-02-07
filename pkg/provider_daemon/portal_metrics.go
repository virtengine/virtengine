package provider_daemon

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (s *PortalAPIServer) handleDeploymentMetrics(w http.ResponseWriter, r *http.Request) {
	deploymentID := mux.Vars(r)["deploymentId"]
	if deploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "deployment id required")
		return
	}

	metrics, err := s.chainQuery.GetDeploymentMetrics(r.Context(), deploymentID)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

func (s *PortalAPIServer) handleDeploymentMetricsHistory(w http.ResponseWriter, r *http.Request) {
	deploymentID := mux.Vars(r)["deploymentId"]
	if deploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "deployment id required")
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	interval, err := parseInterval(r, time.Minute*5)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	series, err := s.chainQuery.GetDeploymentMetricsHistory(r.Context(), deploymentID, start, end, interval)
	if err != nil {
		if errorsIsNotFound(err) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, series)
}

func (s *PortalAPIServer) handleAggregatedMetrics(w http.ResponseWriter, r *http.Request) {
	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	interval, err := parseInterval(r, time.Minute*5)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	series, err := s.chainQuery.GetAggregatedMetrics(r.Context(), start, end, interval)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, series)
}
