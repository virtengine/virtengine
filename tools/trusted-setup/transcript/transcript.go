package transcript

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	Phase1 = "phase1"
	Phase2 = "phase2"
)

type ContributionRecord struct {
	ID             string    `json:"id"`
	Phase          string    `json:"phase"`
	ParticipantID  string    `json:"participant_id"`
	PublicKey      string    `json:"public_key"`
	Signature      string    `json:"signature"`
	Attestation    string    `json:"attestation"`
	InputHash      string    `json:"input_hash"`
	OutputHash     string    `json:"output_hash"`
	ContributionAt time.Time `json:"contribution_at"`
	VerifiedAt     time.Time `json:"verified_at"`
	File           string    `json:"file"`
}

type PhaseSummary struct {
	InitialHash      string               `json:"initial_hash"`
	FinalHash        string               `json:"final_hash"`
	ContributorCount int                  `json:"contributor_count"`
	Contributions    []ContributionRecord `json:"contributions"`
}

type FinalSummary struct {
	ProvingKeyHash    string `json:"proving_key_hash"`
	VerifyingKeyHash  string `json:"verifying_key_hash"`
	Beacon            string `json:"beacon"`
	CompletedAt       string `json:"completed_at"`
	ParametersVersion string `json:"parameters_version"`
}

type Transcript struct {
	CeremonyID  string       `json:"ceremony_id"`
	CircuitName string       `json:"circuit_name"`
	CircuitHash string       `json:"circuit_hash"`
	Curve       string       `json:"curve"`
	DomainSize  uint64       `json:"domain_size"`
	R1CSHash    string       `json:"r1cs_hash"`
	CreatedAt   string       `json:"created_at"`
	Phase1      PhaseSummary `json:"phase1"`
	Phase2      PhaseSummary `json:"phase2"`
	Final       FinalSummary `json:"final"`
	Notes       []string     `json:"notes"`
}

func New(ceremonyID, circuitName, circuitHash, curve string, domainSize uint64, r1csHash string) *Transcript {
	return &Transcript{
		CeremonyID:  ceremonyID,
		CircuitName: circuitName,
		CircuitHash: circuitHash,
		Curve:       curve,
		DomainSize:  domainSize,
		R1CSHash:    r1csHash,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func (t *Transcript) AddContribution(record ContributionRecord) {
	switch record.Phase {
	case Phase1:
		t.Phase1.Contributions = append(t.Phase1.Contributions, record)
		t.Phase1.ContributorCount = len(t.Phase1.Contributions)
	case Phase2:
		t.Phase2.Contributions = append(t.Phase2.Contributions, record)
		t.Phase2.ContributorCount = len(t.Phase2.Contributions)
	}
}

func (t *Transcript) Marshal() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func HashJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

func VerifyHash(expected string, data []byte) error {
	actual := HashBytes(data)
	if actual != expected {
		return fmt.Errorf("hash mismatch: expected %s got %s", expected, actual)
	}
	return nil
}
