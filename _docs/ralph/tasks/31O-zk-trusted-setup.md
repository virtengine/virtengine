# Task 31O: ZK Trusted Setup Ceremony

**vibe-kanban ID:** `c994fc4d-9749-420b-928d-f02a24f19815`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31O |
| **Title** | feat(zk): Trusted setup ceremony |
| **Priority** | P0 (Critical) |
| **Wave** | 1 (Pre-Launch Blocker) |
| **Estimated LOC** | 3500 |
| **Duration** | 4-5 weeks |
| **Dependencies** | gnark ZK circuits |
| **Blocking** | Mainnet launch |

---

## Problem Statement

The gnark v0.14.0 ZK circuits for VEID require a trusted setup ceremony before mainnet. Without it:
- ZK proofs cannot be generated/verified in production
- Anyone could forge proofs (security breach)
- Network is not production-ready

The ceremony must be:
- Multi-party computation (MPC) with multiple participants
- Verifiable and auditable
- Reproducible
- Toxic waste must be destroyed

### Current State Analysis

```
go.mod:
  github.com/consensys/gnark v0.14.0  ✅ Dependency exists

x/veid/zk/circuits/                    ⚠️  Circuits implemented
x/veid/zk/params/                      ❌ No production parameters
tools/trusted-setup/                   ❌ Does not exist
```

---

## Acceptance Criteria

### AC-1: Ceremony Infrastructure
- [ ] MPC coordinator service
- [ ] Participant client software
- [ ] Contribution verification
- [ ] Transcript generation
- [ ] Toxic waste secure deletion

### AC-2: Ceremony Execution
- [ ] Minimum 20 participants
- [ ] At least 5 known/reputable contributors
- [ ] Geographically distributed
- [ ] Hardware entropy sources
- [ ] Air-gapped contribution option

### AC-3: Parameter Generation
- [ ] Generate proving key (pk)
- [ ] Generate verification key (vk)
- [ ] Compress parameters for distribution
- [ ] Embed vk in chain binary
- [ ] Parameter versioning

### AC-4: Verification and Audit
- [ ] Verifiable transcript of all contributions
- [ ] Third-party audit of ceremony
- [ ] Public verification tools
- [ ] Documentation of security properties

---

## Technical Requirements

### ZK Circuit Overview

```go
// x/veid/zk/circuits/veid_circuit.go

package circuits

import (
    "github.com/consensys/gnark/frontend"
    "github.com/consensys/gnark/std/hash/mimc"
)

// VEIDCircuit proves identity verification without revealing PII
type VEIDCircuit struct {
    // Private inputs (witness)
    IdentityHash    frontend.Variable `gnark:",secret"`
    BiometricHash   frontend.Variable `gnark:",secret"`
    DocumentHash    frontend.Variable `gnark:",secret"`
    Salt            frontend.Variable `gnark:",secret"`
    
    // Public inputs
    CommitmentHash  frontend.Variable `gnark:",public"`
    ValidatorSetHash frontend.Variable `gnark:",public"`
    TrustScore      frontend.Variable `gnark:",public"`
    Timestamp       frontend.Variable `gnark:",public"`
}

func (c *VEIDCircuit) Define(api frontend.API) error {
    // Hash all identity components
    mimc, _ := mimc.NewMiMC(api)
    
    mimc.Write(c.IdentityHash)
    mimc.Write(c.BiometricHash)
    mimc.Write(c.DocumentHash)
    mimc.Write(c.Salt)
    
    computed := mimc.Sum()
    
    // Verify commitment matches
    api.AssertIsEqual(computed, c.CommitmentHash)
    
    // Range check trust score (0-100)
    api.AssertIsLessOrEqual(c.TrustScore, 100)
    
    return nil
}
```

### Trusted Setup Coordinator

```go
// tools/trusted-setup/coordinator/coordinator.go

package coordinator

import (
    "context"
    "crypto/rand"
    "fmt"
    "sync"
    "time"
    
    "github.com/consensys/gnark-crypto/ecc"
    "github.com/consensys/gnark/backend/groth16"
    "github.com/consensys/gnark/frontend"
    "github.com/consensys/gnark/frontend/cs/r1cs"
)

type Coordinator struct {
    circuit       frontend.Circuit
    currentPhase  Phase
    contributions []Contribution
    mu            sync.RWMutex
    
    // State
    srs          *groth16.ProvingKey
    vk           *groth16.VerifyingKey
    transcript   *Transcript
    
    // Config
    minContributors int
    curve           ecc.ID
}

type Phase int

const (
    PhaseWaiting Phase = iota
    PhaseContributing
    PhaseVerifying
    PhaseFinalized
)

type Contribution struct {
    ID            string
    ParticipantID string
    PublicKey     []byte
    Contribution  []byte
    Hash          []byte
    Timestamp     time.Time
    Verified      bool
    VerifiedAt    *time.Time
}

type Transcript struct {
    CircuitHash     []byte
    InitialParams   []byte
    Contributions   []ContributionRecord
    FinalParams     []byte
    FinalHash       []byte
}

type ContributionRecord struct {
    ParticipantID    string
    ParticipantPubKey []byte
    InputHash        []byte
    OutputHash       []byte
    Signature        []byte
    Timestamp        time.Time
}

func NewCoordinator(circuit frontend.Circuit, minContributors int) (*Coordinator, error) {
    return &Coordinator{
        circuit:         circuit,
        currentPhase:    PhaseWaiting,
        contributions:   make([]Contribution, 0),
        minContributors: minContributors,
        curve:           ecc.BN254,  // Common curve for Ethereum compatibility
        transcript:      &Transcript{},
    }, nil
}

func (c *Coordinator) Initialize(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Compile circuit to R1CS
    r1cs, err := frontend.Compile(c.curve.ScalarField(), r1cs.NewBuilder, c.circuit)
    if err != nil {
        return fmt.Errorf("compile circuit: %w", err)
    }
    
    // Generate initial random parameters (Phase 1)
    // In production, this should use Powers of Tau from a larger ceremony
    pk, vk, err := groth16.Setup(r1cs)
    if err != nil {
        return fmt.Errorf("initial setup: %w", err)
    }
    
    c.srs = &pk
    c.vk = &vk
    
    // Record in transcript
    c.transcript.CircuitHash = hashCircuit(r1cs)
    c.transcript.InitialParams = serializeParams(pk)
    
    c.currentPhase = PhaseContributing
    return nil
}

func (c *Coordinator) AcceptContribution(ctx context.Context, contrib Contribution) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.currentPhase != PhaseContributing {
        return fmt.Errorf("not accepting contributions, current phase: %v", c.currentPhase)
    }
    
    // Verify contribution is valid transformation
    if err := c.verifyContribution(contrib); err != nil {
        return fmt.Errorf("invalid contribution: %w", err)
    }
    
    // Apply contribution to SRS
    if err := c.applyContribution(contrib); err != nil {
        return fmt.Errorf("apply contribution: %w", err)
    }
    
    // Record
    contrib.Verified = true
    now := time.Now()
    contrib.VerifiedAt = &now
    c.contributions = append(c.contributions, contrib)
    
    // Record in transcript
    c.transcript.Contributions = append(c.transcript.Contributions, ContributionRecord{
        ParticipantID:     contrib.ParticipantID,
        ParticipantPubKey: contrib.PublicKey,
        InputHash:         c.transcript.Contributions[len(c.transcript.Contributions)-1].OutputHash,
        OutputHash:        contrib.Hash,
        Signature:         contrib.Signature,
        Timestamp:         contrib.Timestamp,
    })
    
    return nil
}

func (c *Coordinator) Finalize(ctx context.Context) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if len(c.contributions) < c.minContributors {
        return fmt.Errorf("insufficient contributions: %d < %d", len(c.contributions), c.minContributors)
    }
    
    c.currentPhase = PhaseVerifying
    
    // Verify entire transcript
    if err := c.verifyTranscript(); err != nil {
        return fmt.Errorf("transcript verification failed: %w", err)
    }
    
    // Finalize parameters
    c.transcript.FinalParams = serializeParams(*c.srs)
    c.transcript.FinalHash = hashParams(c.transcript.FinalParams)
    
    c.currentPhase = PhaseFinalized
    return nil
}

func (c *Coordinator) ExportParameters() (*Parameters, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.currentPhase != PhaseFinalized {
        return nil, fmt.Errorf("ceremony not finalized")
    }
    
    return &Parameters{
        ProvingKey:      c.srs,
        VerifyingKey:    c.vk,
        TranscriptHash:  c.transcript.FinalHash,
        ContributorCount: len(c.contributions),
    }, nil
}

func (c *Coordinator) verifyContribution(contrib Contribution) error {
    // Verify the contribution is a valid random linear combination
    // This ensures the contributor added entropy without corrupting params
    
    // 1. Verify signature over contribution hash
    // 2. Verify algebraic consistency
    // 3. Verify non-degeneracy
    
    return nil // Simplified
}

func (c *Coordinator) applyContribution(contrib Contribution) error {
    // Apply the contribution's random scalar to the SRS
    // pk' = pk * random_scalar
    return nil // Simplified
}
```

### Participant Client

```go
// tools/trusted-setup/participant/participant.go

package participant

import (
    "context"
    "crypto/rand"
    "fmt"
    "time"
    
    "golang.org/x/crypto/argon2"
)

type ParticipantClient struct {
    coordinatorURL string
    identity       *Identity
    entropy        *EntropySource
}

type Identity struct {
    ID         string
    PublicKey  []byte
    PrivateKey []byte
}

type EntropySource struct {
    hardwareRNG bool
    userInput   []byte
    systemNoise []byte
    timestamp   time.Time
}

func NewParticipant(coordinatorURL string) (*ParticipantClient, error) {
    identity, err := generateIdentity()
    if err != nil {
        return nil, err
    }
    
    return &ParticipantClient{
        coordinatorURL: coordinatorURL,
        identity:       identity,
        entropy:        &EntropySource{},
    }, nil
}

func (p *ParticipantClient) CollectEntropy(ctx context.Context) error {
    // Collect entropy from multiple sources
    
    // 1. Hardware RNG if available
    hwEntropy := make([]byte, 64)
    if _, err := rand.Read(hwEntropy); err != nil {
        return err
    }
    
    // 2. Request user input (keyboard timing, mouse movement)
    fmt.Println("Please type random characters for 30 seconds:")
    userEntropy, err := p.collectUserInput(ctx, 30*time.Second)
    if err != nil {
        return err
    }
    
    // 3. System noise (process list, network stats, disk access)
    systemEntropy := p.collectSystemNoise()
    
    // Combine all entropy sources
    p.entropy = &EntropySource{
        hardwareRNG: true,
        userInput:   userEntropy,
        systemNoise: systemEntropy,
        timestamp:   time.Now(),
    }
    
    return nil
}

func (p *ParticipantClient) GenerateContribution(ctx context.Context) (*Contribution, error) {
    // Fetch current parameters from coordinator
    currentParams, err := p.fetchCurrentParams(ctx)
    if err != nil {
        return nil, err
    }
    
    // Generate random scalar from entropy
    entropyHash := p.hashEntropy()
    randomScalar := deriveScalar(entropyHash)
    
    // Apply transformation
    newParams := applyScalarToParams(currentParams, randomScalar)
    
    // Generate proof of correct transformation
    proof := generateTransformationProof(currentParams, newParams, randomScalar)
    
    // Sign contribution
    signature := p.sign(newParams)
    
    // Securely delete the random scalar (toxic waste)
    secureDelete(randomScalar)
    
    return &Contribution{
        ID:           generateID(),
        ParticipantID: p.identity.ID,
        PublicKey:    p.identity.PublicKey,
        Contribution: newParams,
        Proof:        proof,
        Hash:         hashParams(newParams),
        Signature:    signature,
        Timestamp:    time.Now(),
    }, nil
}

func (p *ParticipantClient) Submit(ctx context.Context, contrib *Contribution) error {
    // Submit to coordinator
    return p.submitToCoordinator(ctx, contrib)
}

func (p *ParticipantClient) hashEntropy() []byte {
    combined := append(p.entropy.userInput, p.entropy.systemNoise...)
    combined = append(combined, []byte(p.entropy.timestamp.String())...)
    
    // Use Argon2 for memory-hard derivation
    return argon2.IDKey(combined, []byte("virtengine-trusted-setup"), 1, 64*1024, 4, 64)
}

// secureDelete overwrites memory before freeing
func secureDelete(data []byte) {
    for i := range data {
        data[i] = 0
    }
    // Multiple overwrites for paranoid deletion
    rand.Read(data)
    for i := range data {
        data[i] = 0xFF
    }
    for i := range data {
        data[i] = 0
    }
}
```

### Parameter Embedding

```go
// x/veid/zk/params/embed.go

package params

import (
    _ "embed"
    
    "github.com/consensys/gnark/backend/groth16"
    "github.com/consensys/gnark-crypto/ecc"
)

//go:embed veid_vk.bin
var verifyingKeyBytes []byte

//go:embed params_metadata.json
var metadataBytes []byte

type Metadata struct {
    Version          string   `json:"version"`
    CeremonyHash     string   `json:"ceremony_hash"`
    ContributorCount int      `json:"contributor_count"`
    Contributors     []string `json:"contributors"`
    CircuitHash      string   `json:"circuit_hash"`
    GeneratedAt      string   `json:"generated_at"`
}

var (
    cachedVK *groth16.VerifyingKey
)

func GetVerifyingKey() (*groth16.VerifyingKey, error) {
    if cachedVK != nil {
        return cachedVK, nil
    }
    
    vk := groth16.NewVerifyingKey(ecc.BN254)
    if err := vk.ReadFrom(bytes.NewReader(verifyingKeyBytes)); err != nil {
        return nil, fmt.Errorf("parse verifying key: %w", err)
    }
    
    cachedVK = &vk
    return cachedVK, nil
}

func GetMetadata() (*Metadata, error) {
    var meta Metadata
    if err := json.Unmarshal(metadataBytes, &meta); err != nil {
        return nil, err
    }
    return &meta, nil
}

// VerifyProof verifies a VEID ZK proof against the embedded verifying key
func VerifyProof(proof []byte, publicInputs [][]byte) (bool, error) {
    vk, err := GetVerifyingKey()
    if err != nil {
        return false, err
    }
    
    // Deserialize proof
    p := groth16.NewProof(ecc.BN254)
    if _, err := p.ReadFrom(bytes.NewReader(proof)); err != nil {
        return false, fmt.Errorf("parse proof: %w", err)
    }
    
    // Convert public inputs
    witness, err := frontend.NewWitness(&VEIDPublicInputs{
        // Map publicInputs to circuit public variables
    }, ecc.BN254.ScalarField(), frontend.PublicOnly())
    if err != nil {
        return false, err
    }
    
    // Verify
    return groth16.Verify(p, *vk, witness) == nil, nil
}
```

### Ceremony CLI

```go
// tools/trusted-setup/cmd/ceremony/main.go

package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "ceremony",
    Short: "VirtEngine ZK trusted setup ceremony tools",
}

var coordinatorCmd = &cobra.Command{
    Use:   "coordinator",
    Short: "Run the ceremony coordinator",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := context.Background()
        
        coordinator, err := NewCoordinator(circuit, minContributors)
        if err != nil {
            return err
        }
        
        return coordinator.Run(ctx, port)
    },
}

var participateCmd = &cobra.Command{
    Use:   "participate",
    Short: "Participate in the ceremony",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := context.Background()
        
        participant, err := NewParticipant(coordinatorURL)
        if err != nil {
            return err
        }
        
        fmt.Println("=== VirtEngine Trusted Setup Ceremony ===")
        fmt.Println()
        fmt.Println("You are about to contribute to the VirtEngine ZK trusted setup.")
        fmt.Println("Your contribution will help secure the network.")
        fmt.Println()
        
        // Collect entropy
        fmt.Println("Step 1: Collecting entropy...")
        if err := participant.CollectEntropy(ctx); err != nil {
            return err
        }
        
        // Generate contribution
        fmt.Println("Step 2: Generating contribution...")
        contrib, err := participant.GenerateContribution(ctx)
        if err != nil {
            return err
        }
        
        // Submit
        fmt.Println("Step 3: Submitting contribution...")
        if err := participant.Submit(ctx, contrib); err != nil {
            return err
        }
        
        fmt.Println()
        fmt.Println("✅ Contribution submitted successfully!")
        fmt.Printf("Contribution hash: %x\n", contrib.Hash)
        fmt.Println()
        fmt.Println("Thank you for participating!")
        
        return nil
    },
}

var verifyCmd = &cobra.Command{
    Use:   "verify",
    Short: "Verify the ceremony transcript",
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := context.Background()
        
        transcript, err := loadTranscript(transcriptPath)
        if err != nil {
            return err
        }
        
        fmt.Println("Verifying ceremony transcript...")
        
        if err := verifyTranscript(ctx, transcript); err != nil {
            return fmt.Errorf("verification failed: %w", err)
        }
        
        fmt.Println("✅ Transcript verified successfully!")
        fmt.Printf("Contributors: %d\n", len(transcript.Contributions))
        fmt.Printf("Final hash: %x\n", transcript.FinalHash)
        
        return nil
    },
}

func main() {
    rootCmd.AddCommand(coordinatorCmd)
    rootCmd.AddCommand(participateCmd)
    rootCmd.AddCommand(verifyCmd)
    
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

---

## Directory Structure

```
tools/trusted-setup/
├── cmd/
│   └── ceremony/
│       └── main.go           # CLI entry point
├── coordinator/
│   ├── coordinator.go        # MPC coordinator
│   ├── server.go             # HTTP/gRPC server
│   └── transcript.go         # Transcript management
├── participant/
│   ├── participant.go        # Participant client
│   ├── entropy.go            # Entropy collection
│   └── airgap.go             # Air-gapped contribution
├── verify/
│   └── verify.go             # Transcript verification
└── docs/
    ├── CEREMONY_GUIDE.md
    └── SECURITY_MODEL.md

x/veid/zk/
├── circuits/
│   └── veid_circuit.go       # ZK circuit definition
├── params/
│   ├── embed.go              # Parameter embedding
│   ├── veid_vk.bin           # Verifying key (post-ceremony)
│   └── params_metadata.json  # Ceremony metadata
└── verifier.go               # Proof verification
```

---

## Ceremony Timeline

| Phase | Duration | Description |
|-------|----------|-------------|
| 1. Preparation | 1 week | Circuit finalization, tooling, security review |
| 2. Beta Testing | 1 week | Test ceremony with internal team |
| 3. Public Ceremony | 2 weeks | Open contribution period (min 20 contributors) |
| 4. Verification | 3 days | Verify transcript, generate final params |
| 5. Audit | 1 week | Third-party audit of ceremony and parameters |
| 6. Integration | 3 days | Embed parameters, release new binary |

---

## Security Requirements

1. **Multi-Party**: Minimum 20 participants
2. **Diversity**: Contributors from different organizations/geographies
3. **Entropy**: Hardware + user + system entropy for each contribution
4. **Toxic Waste**: Secure deletion with memory overwrite
5. **Transcript**: Full audit trail of all contributions
6. **Air-Gap Option**: Support for offline contribution
7. **Verification**: Public tools to verify ceremony integrity

---

## Testing Requirements

### Unit Tests
- Contribution verification
- Entropy hashing
- Parameter serialization

### Integration Tests
- Full ceremony flow with test participants
- Transcript verification

### Security Tests
- Memory inspection for toxic waste
- Entropy quality analysis

---

## Success Criteria

| Metric | Target |
|--------|--------|
| Minimum contributors | 20+ |
| Known reputable contributors | 5+ |
| Geographic distribution | 5+ countries |
| Ceremony duration | < 2 weeks |
| Verification passes | 100% |
| Third-party audit | Pass |
