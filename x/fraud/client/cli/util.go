package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
	"github.com/virtengine/virtengine/x/fraud/types"
)

func normalizeEnumInput(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "fraud_category_")
	value = strings.TrimPrefix(value, "fraud_report_status_")
	value = strings.TrimPrefix(value, "resolution_type_")
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func parseFraudCategory(value string) (fraudv1.FraudCategory, error) {
	value = normalizeEnumInput(value)
	if value == "" {
		return fraudv1.FraudCategoryUnspecified, fmt.Errorf("category is required")
	}
	if numeric, err := strconv.ParseInt(value, 10, 32); err == nil {
		if numeric < 0 || numeric > int64(fraudv1.FraudCategoryOther) {
			return fraudv1.FraudCategoryUnspecified, fmt.Errorf("invalid fraud category: %s", value)
		}
		return fraudv1.FraudCategory(numeric), nil
	}

	for cat, name := range types.FraudCategoryNames {
		if value == name {
			return types.FraudCategoryToProto(cat), nil
		}
	}

	return fraudv1.FraudCategoryUnspecified, fmt.Errorf("unknown fraud category: %s", value)
}

func parseFraudStatus(value string) (fraudv1.FraudReportStatus, error) {
	value = normalizeEnumInput(value)
	if value == "" {
		return fraudv1.FraudReportStatusUnspecified, fmt.Errorf("status is required")
	}
	if numeric, err := strconv.ParseInt(value, 10, 32); err == nil {
		if numeric < 0 || numeric > int64(fraudv1.FraudReportStatusEscalated) {
			return fraudv1.FraudReportStatusUnspecified, fmt.Errorf("invalid fraud report status: %s", value)
		}
		return fraudv1.FraudReportStatus(numeric), nil
	}

	for status, name := range types.FraudReportStatusNames {
		if value == name {
			return types.FraudReportStatusToProto(status), nil
		}
	}

	return fraudv1.FraudReportStatusUnspecified, fmt.Errorf("unknown fraud report status: %s", value)
}

func parseResolutionType(value string) (fraudv1.ResolutionType, error) {
	value = normalizeEnumInput(value)
	if value == "" {
		return fraudv1.ResolutionTypeUnspecified, fmt.Errorf("resolution is required")
	}
	if numeric, err := strconv.ParseInt(value, 10, 32); err == nil {
		if numeric < 0 || numeric > int64(fraudv1.ResolutionTypeNoAction) {
			return fraudv1.ResolutionTypeUnspecified, fmt.Errorf("invalid resolution type: %s", value)
		}
		return fraudv1.ResolutionType(numeric), nil
	}

	for resolution, name := range types.ResolutionTypeNames {
		if value == name {
			return types.ResolutionTypeToProto(resolution), nil
		}
	}

	return fraudv1.ResolutionTypeUnspecified, fmt.Errorf("unknown resolution type: %s", value)
}

func readEvidenceFromFlags(evidenceJSON string, evidenceFile string) ([]fraudv1.EncryptedEvidence, error) {
	var evidence []fraudv1.EncryptedEvidence

	if strings.HasPrefix(strings.TrimSpace(evidenceJSON), "@") {
		evidenceFile = strings.TrimPrefix(strings.TrimSpace(evidenceJSON), "@")
		evidenceJSON = ""
	}

	if evidenceFile != "" {
		data, err := os.ReadFile(evidenceFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read evidence file: %w", err)
		}
		parsed, err := parseEvidenceJSON(string(data))
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, parsed...)
	}

	if strings.TrimSpace(evidenceJSON) != "" {
		parsed, err := parseEvidenceJSON(evidenceJSON)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, parsed...)
	}

	if len(evidence) == 0 {
		return nil, fmt.Errorf("evidence is required; provide --%s or --%s", flagEvidence, flagEvidenceFile)
	}

	for i := range evidence {
		if err := types.ValidateEncryptedEvidenceValue(evidence[i]); err != nil {
			return nil, fmt.Errorf("invalid evidence[%d]: %w", i, err)
		}
	}

	return evidence, nil
}

func parseEvidenceJSON(raw string) ([]fraudv1.EncryptedEvidence, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("evidence payload is empty")
	}

	var evidence []fraudv1.EncryptedEvidence
	if strings.HasPrefix(trimmed, "{") {
		var single fraudv1.EncryptedEvidence
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			return nil, fmt.Errorf("invalid evidence JSON object: %w", err)
		}
		evidence = append(evidence, single)
	} else {
		if err := json.Unmarshal([]byte(trimmed), &evidence); err != nil {
			return nil, fmt.Errorf("invalid evidence JSON array: %w", err)
		}
	}

	if len(evidence) == 0 {
		return nil, fmt.Errorf("evidence payload is empty")
	}

	return evidence, nil
}
