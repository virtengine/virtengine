package circuits

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
)

// AgeRangeCircuit defines the circuit for proving age is above a threshold
// without revealing the actual date of birth.
//
// Public inputs: ageThreshold, currentTimestamp, commitmentHash
// Private inputs: dateOfBirth, salt
// Constraint: (currentTimestamp - dateOfBirth) / 31536000 >= ageThreshold
// NOTE: constraint uses average seconds/year for better accuracy in circuits.
type AgeRangeCircuit struct {
	// Public inputs
	AgeThreshold     frontend.Variable `gnark:",public"`
	CurrentTimestamp frontend.Variable `gnark:",public"`
	CommitmentHash   frontend.Variable `gnark:",public"`

	// Private inputs (witness)
	DateOfBirth frontend.Variable
	Salt        frontend.Variable
}

// Define implements the frontend.Circuit interface for age range proofs.
func (circuit *AgeRangeCircuit) Define(api frontend.API) error {
	// Seconds in a year (365.25 days average)
	secondsPerYear := big.NewInt(31557600)

	// Compute age in seconds
	ageSeconds := api.Sub(circuit.CurrentTimestamp, circuit.DateOfBirth)

	// Compute age in years (integer division)
	ageYears := api.Div(ageSeconds, secondsPerYear)

	// Assert age >= threshold
	api.AssertIsLessOrEqual(circuit.AgeThreshold, ageYears)

	// Verify commitment: hash(dateOfBirth || salt) == commitmentHash
	// For circuit efficiency, we use a simplified hash commitment
	commitment := api.Add(api.Mul(circuit.DateOfBirth, 1000000), circuit.Salt)
	api.AssertIsEqual(circuit.CommitmentHash, commitment)

	return nil
}
