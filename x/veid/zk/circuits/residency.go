package circuits

import "github.com/consensys/gnark/frontend"

// ResidencyCircuit defines the circuit for proving residency in a country
// without revealing the full address.
//
// Public inputs: countryCodeHash, commitmentHash
// Private inputs: fullAddressHash, salt
// Constraint: countryCodeHash == addressCountry and commitment binding.
type ResidencyCircuit struct {
	// Public inputs
	CountryCodeHash frontend.Variable `gnark:",public"`
	CommitmentHash  frontend.Variable `gnark:",public"`

	// Private inputs (witness)
	FullAddressHash frontend.Variable
	AddressCountry  frontend.Variable
	Salt            frontend.Variable
}

// Define implements the frontend.Circuit interface for residency proofs.
func (circuit *ResidencyCircuit) Define(api frontend.API) error {
	// Assert that the address country matches the claimed country
	api.AssertIsEqual(circuit.CountryCodeHash, circuit.AddressCountry)

	// Verify commitment: simplified commitment scheme
	commitment := api.Add(api.Mul(circuit.FullAddressHash, 1000), circuit.Salt)
	api.AssertIsEqual(circuit.CommitmentHash, commitment)

	return nil
}
