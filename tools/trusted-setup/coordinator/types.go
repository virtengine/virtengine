package coordinator

import (
	"time"
)

type Config struct {
	CeremonyID      string   `json:"ceremony_id"`
	CircuitName     string   `json:"circuit_name"`
	CircuitHash     string   `json:"circuit_hash"`
	Curve           string   `json:"curve"`
	MinContributors int      `json:"min_contributors"`
	Beacon          string   `json:"beacon"`
	DomainSize      uint64   `json:"domain_size"`
	R1CSHash        string   `json:"r1cs_hash"`
	CreatedAt       string   `json:"created_at"`
	Notes           []string `json:"notes"`
}

type ContributionMeta struct {
	ParticipantID string
	PublicKey     string
	Signature     string
	Attestation   string
}

type Status struct {
	CeremonyID        string    `json:"ceremony_id"`
	CircuitName       string    `json:"circuit_name"`
	Curve             string    `json:"curve"`
	Phase             string    `json:"phase"`
	Phase1Count       int       `json:"phase1_count"`
	Phase2Count       int       `json:"phase2_count"`
	MinContributors   int       `json:"min_contributors"`
	LastContribution  time.Time `json:"last_contribution"`
	TranscriptHash    string    `json:"transcript_hash"`
	ParametersVersion string    `json:"parameters_version"`
}
