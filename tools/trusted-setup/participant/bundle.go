package participant

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

func RespondToPhaseBundle(bundleDir, outDir string, client *Client) (*bundle.PhaseResponse, error) {
	if client == nil {
		return nil, fmt.Errorf("participant client is required")
	}
	var request bundle.PhaseRequest
	requestPath := filepath.Join(bundleDir, "request.json")
	if err := bundle.ReadJSON(requestPath, &request); err != nil {
		return nil, err
	}
	if request.SchemaVersion != bundle.PhaseRequestSchema {
		return nil, fmt.Errorf("unsupported request schema %q", request.SchemaVersion)
	}

	inputPath := filepath.Join(bundleDir, request.InputFile)
	inputPayload, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	actualInputHash := transcript.HashBytes(inputPayload)
	if actualInputHash != request.InputHash {
		return nil, fmt.Errorf("request input hash mismatch: %s != %s", request.InputHash, actualInputHash)
	}

	var outputPayload []byte
	var signature string
	switch request.Phase {
	case transcript.Phase1:
		outputPayload, signature, err = client.ContributePhase1(inputPayload)
	case transcript.Phase2:
		outputPayload, signature, err = client.ContributePhase2(inputPayload)
	default:
		return nil, fmt.Errorf("unsupported phase %q", request.Phase)
	}
	if err != nil {
		return nil, err
	}

	requestBytes, err := os.ReadFile(requestPath)
	if err != nil {
		return nil, err
	}
	response := &bundle.PhaseResponse{
		SchemaVersion: bundle.PhaseResponseSchema,
		CeremonyID:    request.CeremonyID,
		Phase:         request.Phase,
		RequestHash:   transcript.HashBytes(requestBytes),
		PayloadFile:   "contribution.bin",
		ParticipantID: client.Identity.ID,
		PublicKey:     client.Identity.PublicKey,
		Attestation:   client.Attestation,
		InputHash:     request.InputHash,
		OutputHash:    transcript.HashBytes(outputPayload),
		Signature:     signature,
	}

	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return nil, err
	}
	payloadPath := filepath.Join(outDir, response.PayloadFile)
	responsePath := filepath.Join(outDir, "response.json")
	if err := os.WriteFile(payloadPath, outputPayload, 0o600); err != nil {
		return nil, err
	}
	if err := bundle.WriteJSON(responsePath, response); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, response.PayloadFile+".sha256"), payloadPath); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, "response.json.sha256"), responsePath); err != nil {
		return nil, err
	}

	return response, nil
}
