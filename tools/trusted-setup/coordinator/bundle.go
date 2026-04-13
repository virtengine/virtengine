package coordinator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

const (
	phaseBundlePayloadFile  = "current.bin"
	phaseBundleRequestFile  = "request.json"
	phaseBundleResponseFile = "response.json"
)

func ExportPhaseBundle(state State, phase, outDir string) (*bundle.PhaseRequest, error) {
	cfg, err := state.LoadConfig()
	if err != nil {
		return nil, err
	}
	tr, err := loadTranscript(state)
	if err != nil {
		return nil, err
	}
	payload, err := currentPhasePayload(state, phase)
	if err != nil {
		return nil, err
	}

	request := &bundle.PhaseRequest{
		SchemaVersion:  bundle.PhaseRequestSchema,
		CeremonyID:     cfg.CeremonyID,
		CircuitName:    cfg.CircuitName,
		Phase:          phase,
		TranscriptHash: hashTranscript(tr),
		InputFile:      phaseBundlePayloadFile,
		InputHash:      transcript.HashBytes(payload),
		Notes:          append([]string(nil), cfg.Notes...),
	}

	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return nil, err
	}
	payloadPath := filepath.Join(outDir, phaseBundlePayloadFile)
	requestPath := filepath.Join(outDir, phaseBundleRequestFile)
	if err := writeFileAtomic(payloadPath, payload); err != nil {
		return nil, err
	}
	if err := bundle.WriteJSON(requestPath, request); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, phaseBundlePayloadFile+".sha256"), payloadPath); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, phaseBundleRequestFile+".sha256"), requestPath); err != nil {
		return nil, err
	}

	return request, nil
}

func AcceptPhaseBundle(state State, phase, bundleDir string) error {
	responsePath := filepath.Join(bundleDir, phaseBundleResponseFile)
	payloadPath := filepath.Join(bundleDir, "contribution.bin")

	var response bundle.PhaseResponse
	if err := bundle.ReadJSON(responsePath, &response); err != nil {
		return err
	}
	if response.SchemaVersion != bundle.PhaseResponseSchema {
		return fmt.Errorf("unsupported response schema %q", response.SchemaVersion)
	}
	if response.Phase != phase {
		return fmt.Errorf("response phase mismatch: expected %s got %s", phase, response.Phase)
	}

	payload, err := os.ReadFile(payloadPath)
	if err != nil {
		return err
	}
	if transcript.HashBytes(payload) != response.OutputHash {
		return fmt.Errorf("response output hash mismatch: %s != %s", response.OutputHash, transcript.HashBytes(payload))
	}

	meta := ContributionMeta{
		ParticipantID: response.ParticipantID,
		PublicKey:     response.PublicKey,
		Signature:     response.Signature,
		Attestation:   response.Attestation,
		InputHash:     response.InputHash,
		OutputHash:    response.OutputHash,
	}
	switch phase {
	case transcript.Phase1:
		return AcceptPhase1Contribution(state, payload, meta)
	case transcript.Phase2:
		return AcceptPhase2Contribution(state, payload, meta)
	default:
		return fmt.Errorf("unsupported phase %q", phase)
	}
}

func currentPhasePayload(state State, phase string) ([]byte, error) {
	switch phase {
	case transcript.Phase1:
		_, data, err := loadLatestPhase1(state)
		return data, err
	case transcript.Phase2:
		_, data, err := loadLatestPhase2(state)
		return data, err
	default:
		return nil, fmt.Errorf("unsupported phase %q", phase)
	}
}
