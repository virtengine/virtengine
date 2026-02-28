package types

const (
	ModuleName   = "issuancepolicy"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	PrefixPolicy          = []byte{0x01}
	PrefixActivePolicy    = []byte{0x02}
	PrefixCounters        = []byte{0x03}
	PrefixProofMintRecord = []byte{0x04}
	PrefixParams          = []byte{0x05}
)

func PolicyKey(policyID string) []byte {
	return append(append([]byte{}, PrefixPolicy...), []byte(policyID)...)
}

func PolicyPrefixKey() []byte {
	return PrefixPolicy
}

func ActivePolicyKey() []byte {
	return PrefixActivePolicy
}

func CountersKey() []byte {
	return PrefixCounters
}

func ProofMintRecordKey(proofID string) []byte {
	return append(append([]byte{}, PrefixProofMintRecord...), []byte(proofID)...)
}

func ParamsKey() []byte {
	return PrefixParams
}
