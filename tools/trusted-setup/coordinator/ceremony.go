package coordinator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	zkcircuits "github.com/virtengine/virtengine/x/veid/zk/circuits"
)

const (
	phase1InitialFile = "contrib-0000.bin"
	phase2InitialFile = "contrib-0000.bin"
)

func InitCeremony(state State, circuitName string, minContributors int, beacon string, notes []string) (*Config, error) {
	if err := state.EnsureDirs(); err != nil {
		return nil, err
	}
	if existingCfg, err := state.LoadConfig(); err == nil {
		existingTranscript, transcriptErr := loadTranscript(state)
		if transcriptErr != nil {
			return nil, fmt.Errorf("load existing transcript: %w", transcriptErr)
		}
		if existingCfg.CircuitName != circuitName || existingCfg.MinContributors != minContributors || existingCfg.Beacon != beacon {
			return nil, fmt.Errorf("ceremony workspace already initialized with different config")
		}
		if existingTranscript.CeremonyID != existingCfg.CeremonyID {
			return nil, fmt.Errorf("existing transcript ceremony_id mismatch")
		}
		return existingCfg, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	circuit, err := resolveCircuit(circuitName)
	if err != nil {
		return nil, err
	}

	r1csCompiled, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		return nil, fmt.Errorf("compile circuit: %w", err)
	}
	r1csBN, ok := r1csCompiled.(*cs.R1CS)
	if !ok {
		return nil, fmt.Errorf("unexpected r1cs type %T", r1csCompiled)
	}

	r1csBytes, err := writeToBytes(r1csBN)
	if err != nil {
		return nil, fmt.Errorf("serialize r1cs: %w", err)
	}
	r1csHash := transcript.HashBytes(r1csBytes)
	constraints := r1csBN.GetNbConstraints()
	if constraints < 0 {
		return nil, fmt.Errorf("invalid constraint count: %d", constraints)
	}
	domainSize := ecc.NextPowerOfTwo(uint64(constraints))

	phase1 := mpcsetup.NewPhase1(domainSize)
	phase1Bytes, err := writeToBytes(phase1)
	if err != nil {
		return nil, fmt.Errorf("serialize phase1: %w", err)
	}
	phase1Hash := transcript.HashBytes(phase1Bytes)

	if err := writeFileAtomic(filepath.Join(state.Phase1Dir(), phase1InitialFile), phase1Bytes); err != nil {
		return nil, fmt.Errorf("write phase1 initial: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(state.Phase2Dir(), "r1cs.bin"), r1csBytes); err != nil {
		return nil, fmt.Errorf("write r1cs: %w", err)
	}

	circuitHash := r1csHash
	now := state.Now()
	cfg := &Config{
		CeremonyID:      fmt.Sprintf("veid-%d", now.Unix()),
		CircuitName:     circuitName,
		CircuitHash:     circuitHash,
		Curve:           ecc.BN254.String(),
		MinContributors: minContributors,
		Beacon:          beacon,
		DomainSize:      domainSize,
		R1CSHash:        r1csHash,
		CreatedAt:       now.Format(time.RFC3339),
		Notes:           notes,
	}
	if err := state.SaveConfig(cfg); err != nil {
		return nil, err
	}

	tr := transcript.New(cfg.CeremonyID, circuitName, circuitHash, cfg.Curve, domainSize, r1csHash)
	tr.CreatedAt = now.Format(time.RFC3339)
	tr.Notes = notes
	tr.Phase1.InitialHash = phase1Hash
	data, err := tr.Marshal()
	if err != nil {
		return nil, err
	}
	if err := state.SaveTranscript(data); err != nil {
		return nil, err
	}

	return cfg, nil
}

func StartPhase2(state State) error {
	cfg, err := state.LoadConfig()
	if err != nil {
		return err
	}
	tr, err := loadTranscript(state)
	if err != nil {
		return err
	}

	phase1Contribs, phase1Paths, err := loadPhase1Contributions(state)
	if err != nil {
		return err
	}
	if len(phase1Contribs) < cfg.MinContributors {
		return fmt.Errorf("insufficient phase1 contributions: %d < %d", len(phase1Contribs), cfg.MinContributors)
	}

	commons, err := mpcsetup.VerifyPhase1(cfg.DomainSize, decodeBeacon(cfg.Beacon), phase1Contribs...)
	if err != nil {
		return fmt.Errorf("verify phase1: %w", err)
	}
	commonsBytes, err := writeToBytes(&commons)
	if err != nil {
		return fmt.Errorf("serialize commons: %w", err)
	}
	if err := writeFileAtomic(filepath.Join(state.Phase1Dir(), "commons.bin"), commonsBytes); err != nil {
		return fmt.Errorf("write commons: %w", err)
	}

	if len(phase1Paths) > 0 {
		lastBytes, err := os.ReadFile(phase1Paths[len(phase1Paths)-1])
		if err != nil {
			return err
		}
		tr.Phase1.FinalHash = transcript.HashBytes(lastBytes)
	}

	r1csBN, err := loadR1CS(state)
	if err != nil {
		return err
	}

	var phase2 mpcsetup.Phase2
	phase2.Initialize(r1csBN, &commons)
	phase2Bytes, err := writeToBytes(&phase2)
	if err != nil {
		return fmt.Errorf("serialize phase2: %w", err)
	}
	phase2InitialPath := filepath.Join(state.Phase2Dir(), phase2InitialFile)
	if existingBytes, readErr := os.ReadFile(phase2InitialPath); readErr == nil {
		existingHash := transcript.HashBytes(existingBytes)
		newHash := transcript.HashBytes(phase2Bytes)
		if existingHash == newHash && tr.Phase2.InitialHash == newHash {
			return nil
		}
		return fmt.Errorf("phase2 already initialized with different state")
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	if err := writeFileAtomic(phase2InitialPath, phase2Bytes); err != nil {
		return fmt.Errorf("write phase2 initial: %w", err)
	}
	tr.Phase2.InitialHash = transcript.HashBytes(phase2Bytes)

	return saveTranscript(state, tr)
}

func AcceptPhase1Contribution(state State, payload []byte, meta ContributionMeta) error {
	tr, err := loadTranscript(state)
	if err != nil {
		return err
	}
	if err := validateContributionMeta(meta, transcript.Phase1); err != nil {
		return err
	}
	payloadHash := transcript.HashBytes(payload)
	if lastRecord := latestContribution(tr.Phase1.Contributions); lastRecord != nil &&
		lastRecord.OutputHash == payloadHash &&
		lastRecord.ParticipantID == meta.ParticipantID &&
		lastRecord.Signature == meta.Signature {
		return nil
	}
	prev, prevBytes, err := loadLatestPhase1(state)
	if err != nil {
		return err
	}
	if err := verifyMetaAgainstPayload(meta, transcript.Phase1, prevBytes, payload); err != nil {
		return err
	}

	next := new(mpcsetup.Phase1)
	if _, err := next.ReadFrom(bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("decode contribution: %w", err)
	}
	if err := prev.Verify(next); err != nil {
		return fmt.Errorf("verify contribution: %w", err)
	}

	index := len(tr.Phase1.Contributions) + 1
	filename := fmt.Sprintf("contrib-%04d.bin", index)
	outPath := filepath.Join(state.Phase1Dir(), filename)
	if err := writeFileAtomic(outPath, payload); err != nil {
		return fmt.Errorf("write contribution: %w", err)
	}
	now := state.Now()

	record := transcript.ContributionRecord{
		ID:             fmt.Sprintf("%s-%s-%d", tr.CeremonyID, transcript.Phase1, index),
		Phase:          transcript.Phase1,
		ParticipantID:  meta.ParticipantID,
		PublicKey:      meta.PublicKey,
		Signature:      meta.Signature,
		Attestation:    meta.Attestation,
		InputHash:      transcript.HashBytes(prevBytes),
		OutputHash:     payloadHash,
		ContributionAt: now,
		VerifiedAt:     now,
		File:           filepath.Base(outPath),
	}
	tr.AddContribution(record)
	tr.Phase1.FinalHash = record.OutputHash

	if err := saveTranscript(state, tr); err != nil {
		return err
	}

	return nil
}

func AcceptPhase2Contribution(state State, payload []byte, meta ContributionMeta) error {
	tr, err := loadTranscript(state)
	if err != nil {
		return err
	}
	if err := validateContributionMeta(meta, transcript.Phase2); err != nil {
		return err
	}
	payloadHash := transcript.HashBytes(payload)
	if lastRecord := latestContribution(tr.Phase2.Contributions); lastRecord != nil &&
		lastRecord.OutputHash == payloadHash &&
		lastRecord.ParticipantID == meta.ParticipantID &&
		lastRecord.Signature == meta.Signature {
		return nil
	}
	prev, prevBytes, err := loadLatestPhase2(state)
	if err != nil {
		return err
	}
	if err := verifyMetaAgainstPayload(meta, transcript.Phase2, prevBytes, payload); err != nil {
		return err
	}

	next := new(mpcsetup.Phase2)
	if _, err := next.ReadFrom(bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("decode contribution: %w", err)
	}
	if err := prev.Verify(next); err != nil {
		return fmt.Errorf("verify contribution: %w", err)
	}

	index := len(tr.Phase2.Contributions) + 1
	filename := fmt.Sprintf("contrib-%04d.bin", index)
	outPath := filepath.Join(state.Phase2Dir(), filename)
	if err := writeFileAtomic(outPath, payload); err != nil {
		return fmt.Errorf("write contribution: %w", err)
	}
	now := state.Now()

	record := transcript.ContributionRecord{
		ID:             fmt.Sprintf("%s-%s-%d", tr.CeremonyID, transcript.Phase2, index),
		Phase:          transcript.Phase2,
		ParticipantID:  meta.ParticipantID,
		PublicKey:      meta.PublicKey,
		Signature:      meta.Signature,
		Attestation:    meta.Attestation,
		InputHash:      transcript.HashBytes(prevBytes),
		OutputHash:     payloadHash,
		ContributionAt: now,
		VerifiedAt:     now,
		File:           filepath.Base(outPath),
	}
	tr.AddContribution(record)
	tr.Phase2.FinalHash = record.OutputHash

	return saveTranscript(state, tr)
}

func Finalize(state State, parametersVersion string) (*groth16.ProvingKey, *groth16.VerifyingKey, error) {
	cfg, err := state.LoadConfig()
	if err != nil {
		return nil, nil, err
	}
	tr, err := loadTranscript(state)
	if err != nil {
		return nil, nil, err
	}

	r1csBN, err := loadR1CS(state)
	if err != nil {
		return nil, nil, err
	}

	commons, err := loadCommons(state)
	if err != nil {
		return nil, nil, err
	}

	phase2Contribs, phase2Paths, err := loadPhase2Contributions(state)
	if err != nil {
		return nil, nil, err
	}
	if len(phase2Contribs) < cfg.MinContributors {
		return nil, nil, fmt.Errorf("insufficient phase2 contributions: %d < %d", len(phase2Contribs), cfg.MinContributors)
	}

	pk, vk, err := mpcsetup.VerifyPhase2(r1csBN, commons, decodeBeacon(cfg.Beacon), phase2Contribs...)
	if err != nil {
		return nil, nil, fmt.Errorf("verify phase2: %w", err)
	}

	pkBytes, err := writeToBytes(pk)
	if err != nil {
		return nil, nil, fmt.Errorf("serialize proving key: %w", err)
	}
	vkBytes, err := writeToBytes(vk)
	if err != nil {
		return nil, nil, fmt.Errorf("serialize verifying key: %w", err)
	}
	provingKeyPath := filepath.Join(state.Phase2Dir(), "proving_key.bin")
	verifyingKeyPath := filepath.Join(state.Phase2Dir(), "verifying_key.bin")
	if existingPK, readErr := os.ReadFile(provingKeyPath); readErr == nil {
		existingVK, vkErr := os.ReadFile(verifyingKeyPath)
		if vkErr != nil {
			return nil, nil, vkErr
		}
		if transcript.HashBytes(existingPK) == transcript.HashBytes(pkBytes) && transcript.HashBytes(existingVK) == transcript.HashBytes(vkBytes) && tr.Final.ProvingKeyHash != "" && tr.Final.VerifyingKeyHash != "" {
			return &pk, &vk, nil
		}
	} else if !os.IsNotExist(readErr) {
		return nil, nil, readErr
	}
	if err := writeFileAtomic(provingKeyPath, pkBytes); err != nil {
		return nil, nil, err
	}
	if err := writeFileAtomic(verifyingKeyPath, vkBytes); err != nil {
		return nil, nil, err
	}

	if len(phase2Paths) > 0 {
		lastBytes, err := os.ReadFile(phase2Paths[len(phase2Paths)-1])
		if err != nil {
			return nil, nil, err
		}
		tr.Phase2.FinalHash = transcript.HashBytes(lastBytes)
	}

	tr.Final.ProvingKeyHash = transcript.HashBytes(pkBytes)
	tr.Final.VerifyingKeyHash = transcript.HashBytes(vkBytes)
	tr.Final.Beacon = cfg.Beacon
	tr.Final.CompletedAt = state.Now().Format(time.RFC3339)
	tr.Final.ParametersVersion = parametersVersion

	if err := saveTranscript(state, tr); err != nil {
		return nil, nil, err
	}

	return &pk, &vk, nil
}

func StatusSnapshot(state State) (*Status, error) {
	cfg, err := state.LoadConfig()
	if err != nil {
		return nil, err
	}
	tr, err := loadTranscript(state)
	if err != nil {
		return nil, err
	}
	last := time.Time{}
	if len(tr.Phase1.Contributions) > 0 {
		last = tr.Phase1.Contributions[len(tr.Phase1.Contributions)-1].ContributionAt
	}
	if len(tr.Phase2.Contributions) > 0 {
		phase2Last := tr.Phase2.Contributions[len(tr.Phase2.Contributions)-1].ContributionAt
		if phase2Last.After(last) {
			last = phase2Last
		}
	}

	phase := "phase1"
	if tr.Phase2.InitialHash != "" {
		phase = "phase2"
	}
	if tr.Final.VerifyingKeyHash != "" {
		phase = "finalized"
	}

	return &Status{
		CeremonyID:        cfg.CeremonyID,
		CircuitName:       cfg.CircuitName,
		Curve:             cfg.Curve,
		Phase:             phase,
		Phase1Count:       len(tr.Phase1.Contributions),
		Phase2Count:       len(tr.Phase2.Contributions),
		MinContributors:   cfg.MinContributors,
		LastContribution:  last,
		TranscriptHash:    hashTranscript(tr),
		ParametersVersion: tr.Final.ParametersVersion,
	}, nil
}

func resolveCircuit(name string) (frontend.Circuit, error) {
	switch name {
	case "age-range":
		return &zkcircuits.AgeRangeCircuit{}, nil
	case "residency":
		return &zkcircuits.ResidencyCircuit{}, nil
	case "score-range":
		return &zkcircuits.ScoreRangeCircuit{}, nil
	default:
		return nil, fmt.Errorf("unknown circuit %q", name)
	}
}

func decodeBeacon(beacon string) []byte {
	if beacon == "" {
		return nil
	}
	data, err := hex.DecodeString(beacon)
	if err != nil {
		return []byte(beacon)
	}
	return data
}

func latestContribution(records []transcript.ContributionRecord) *transcript.ContributionRecord {
	if len(records) == 0 {
		return nil
	}
	return &records[len(records)-1]
}

func validateContributionMeta(meta ContributionMeta, phase string) error {
	if meta.ParticipantID == "" {
		return fmt.Errorf("%s contribution missing participant_id", phase)
	}
	if meta.PublicKey == "" {
		return fmt.Errorf("%s contribution missing public_key", phase)
	}
	if meta.Signature == "" {
		return fmt.Errorf("%s contribution missing signature", phase)
	}
	return nil
}

func verifyMetaAgainstPayload(meta ContributionMeta, phase string, inputPayload, outputPayload []byte) error {
	inputHash := transcript.HashBytes(inputPayload)
	outputHash := transcript.HashBytes(outputPayload)
	if meta.InputHash != "" && meta.InputHash != inputHash {
		return fmt.Errorf("%s contribution input hash mismatch: %s != %s", phase, meta.InputHash, inputHash)
	}
	if meta.OutputHash != "" && meta.OutputHash != outputHash {
		return fmt.Errorf("%s contribution output hash mismatch: %s != %s", phase, meta.OutputHash, outputHash)
	}
	if err := participant.VerifyContributionSignature(
		phase,
		meta.ParticipantID,
		meta.PublicKey,
		meta.Attestation,
		meta.Signature,
		inputPayload,
		outputPayload,
	); err != nil {
		return fmt.Errorf("verify %s contribution signature: %w", phase, err)
	}
	return nil
}
