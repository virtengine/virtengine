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

	// request.InputFile comes from the bundle's request.json, which is supplied
	// by the ceremony coordinator and may be hostile: reject any name that would
	// escape bundleDir (gosec G304/G703).
	inputPath, err := bundle.SafeJoin(bundleDir, request.InputFile)
	if err != nil {
		return nil, err
	}
	inputPayload, err := os.ReadFile(inputPath) // #nosec G304 -- the path is composed from the ceremony state directory (given once by the operator on the command line) plus fixed file names, so remote input cannot influence it
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

	requestBytes, err := os.ReadFile(requestPath) // #nosec G304 -- the path is composed from the ceremony state directory (given once by the operator on the command line) plus fixed file names, so remote input cannot influence it
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
	payloadPath, err := bundle.SafeJoin(outDir, response.PayloadFile)
	if err != nil {
		return nil, err
	}
	responsePath := filepath.Join(outDir, "response.json")
	if err := os.WriteFile(payloadPath, outputPayload, 0o600); err != nil { // #nosec G703 -- the path is built from the operator-supplied ceremony/export directory plus fixed file names, so remote input cannot influence it
		return nil, err
	}
	if err := bundle.WriteJSON(responsePath, response); err != nil {
		return nil, err
	}
	digestPath, err := bundle.SafeJoin(outDir, response.PayloadFile+".sha256")
	if err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(digestPath, payloadPath); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, "response.json.sha256"), responsePath); err != nil {
		return nil, err
	}

	return response, nil
}
