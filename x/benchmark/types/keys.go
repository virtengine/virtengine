// Package types contains types for the Benchmark module.
//
// VE-600 through VE-603: Benchmarking on-chain module
package types

const (
	// ModuleName is the name of the benchmark module
	ModuleName = "benchmark"

	// StoreKey is the store key for the benchmark module
	StoreKey = ModuleName

	// RouterKey is the router key for the benchmark module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the benchmark module
	QuerierRoute = ModuleName
)

// Key prefixes for benchmark store
var (
	// BenchmarkReportPrefix is the prefix for benchmark report storage
	BenchmarkReportPrefix = []byte{0x01}

	// ReliabilityScorePrefix is the prefix for reliability score storage
	ReliabilityScorePrefix = []byte{0x02}

	// ChallengePrefix is the prefix for benchmark challenges
	ChallengePrefix = []byte{0x03}

	// ChallengeResponsePrefix is the prefix for challenge responses
	ChallengeResponsePrefix = []byte{0x04}

	// AnomalyFlagPrefix is the prefix for anomaly flags
	AnomalyFlagPrefix = []byte{0x05}

	// ProviderFlagPrefix is the prefix for provider moderation flags
	ProviderFlagPrefix = []byte{0x06}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeyReport is the sequence key for benchmark reports
	SequenceKeyReport = []byte{0x20}

	// SequenceKeyChallenge is the sequence key for challenges
	SequenceKeyChallenge = []byte{0x21}

	// SequenceKeyAnomaly is the sequence key for anomaly flags
	SequenceKeyAnomaly = []byte{0x22}

	// IndexProviderClusterPrefix is the index for provider+cluster
	IndexProviderClusterPrefix = []byte{0x30}

	// IndexRegionPrefix is the index for region
	IndexRegionPrefix = []byte{0x31}
)

// GetBenchmarkReportKey returns the key for a benchmark report
func GetBenchmarkReportKey(reportID string) []byte {
	return append(BenchmarkReportPrefix, []byte(reportID)...)
}

// GetReliabilityScoreKey returns the key for a reliability score
func GetReliabilityScoreKey(providerAddr string) []byte {
	return append(ReliabilityScorePrefix, []byte(providerAddr)...)
}

// GetChallengeKey returns the key for a challenge
func GetChallengeKey(challengeID string) []byte {
	return append(ChallengePrefix, []byte(challengeID)...)
}

// GetChallengeResponseKey returns the key for a challenge response
func GetChallengeResponseKey(challengeID string) []byte {
	return append(ChallengeResponsePrefix, []byte(challengeID)...)
}

// GetAnomalyFlagKey returns the key for an anomaly flag
func GetAnomalyFlagKey(flagID string) []byte {
	return append(AnomalyFlagPrefix, []byte(flagID)...)
}

// GetProviderFlagKey returns the key for a provider flag
func GetProviderFlagKey(providerAddr string) []byte {
	return append(ProviderFlagPrefix, []byte(providerAddr)...)
}

// GetProviderClusterIndexKey returns the index key for provider+cluster
func GetProviderClusterIndexKey(providerAddr, clusterID string) []byte {
	key := append([]byte(nil), IndexProviderClusterPrefix...)
	key = append(key, []byte(providerAddr)...)
	key = append(key, byte(':'))
	key = append(key, []byte(clusterID)...)
	return key
}

// GetRegionIndexKey returns the index key for region
func GetRegionIndexKey(region string) []byte {
	return append(IndexRegionPrefix, []byte(region)...)
}
