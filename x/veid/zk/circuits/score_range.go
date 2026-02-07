package circuits

import "github.com/consensys/gnark/frontend"

// ScoreRangeCircuit defines the circuit for proving a score exceeds a threshold
// without revealing the exact score.
//
// Public inputs: scoreThreshold, commitmentHash
// Private inputs: actualScore, salt
// Constraint: actualScore >= scoreThreshold and commitment binding.
type ScoreRangeCircuit struct {
	// Public inputs
	ScoreThreshold frontend.Variable `gnark:",public"`
	CommitmentHash frontend.Variable `gnark:",public"`

	// Private inputs (witness)
	ActualScore frontend.Variable
	Salt        frontend.Variable
}

// Define implements the frontend.Circuit interface for score range proofs.
func (circuit *ScoreRangeCircuit) Define(api frontend.API) error {
	// Assert score >= threshold
	api.AssertIsLessOrEqual(circuit.ScoreThreshold, circuit.ActualScore)

	// Verify commitment: hash(actualScore || salt) == commitmentHash
	commitment := api.Add(api.Mul(circuit.ActualScore, 1000000), circuit.Salt)
	api.AssertIsEqual(circuit.CommitmentHash, commitment)

	return nil
}
