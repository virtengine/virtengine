// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Store keys and prefixes
package types

const (
	// ModuleName is the name of the fraud module
	ModuleName = "fraud"

	// StoreKey is the store key for the fraud module
	StoreKey = ModuleName

	// RouterKey is the router key for the fraud module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the fraud module
	QuerierRoute = ModuleName
)

// Key prefixes for fraud store
var (
	// FraudReportPrefix is the prefix for fraud report storage
	FraudReportPrefix = []byte{0x01}

	// ModeratorQueuePrefix is the prefix for moderator queue storage
	ModeratorQueuePrefix = []byte{0x02}

	// ReporterIndexPrefix is the prefix for reporter-to-reports index
	ReporterIndexPrefix = []byte{0x03}

	// ReportedPartyIndexPrefix is the prefix for reported-party-to-reports index
	ReportedPartyIndexPrefix = []byte{0x04}

	// StatusIndexPrefix is the prefix for status-to-reports index
	StatusIndexPrefix = []byte{0x05}

	// AuditLogPrefix is the prefix for audit log storage
	AuditLogPrefix = []byte{0x06}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeyFraudReport is the sequence key for fraud reports
	SequenceKeyFraudReport = []byte{0x20}

	// SequenceKeyAuditLog is the sequence key for audit logs
	SequenceKeyAuditLog = []byte{0x21}
)

// GetFraudReportKey returns the key for a fraud report
func GetFraudReportKey(reportID string) []byte {
	return append(FraudReportPrefix, []byte(reportID)...)
}

// GetModeratorQueueKey returns the key for a report in the moderator queue
func GetModeratorQueueKey(reportID string) []byte {
	return append(ModeratorQueuePrefix, []byte(reportID)...)
}

// GetReporterIndexKey returns the index key for a reporter
func GetReporterIndexKey(reporterAddr string) []byte {
	return append(ReporterIndexPrefix, []byte(reporterAddr)...)
}

// GetReportedPartyIndexKey returns the index key for a reported party
func GetReportedPartyIndexKey(reportedAddr string) []byte {
	return append(ReportedPartyIndexPrefix, []byte(reportedAddr)...)
}

// GetStatusIndexKey returns the index key for a status
func GetStatusIndexKey(status FraudReportStatus) []byte {
	return append(StatusIndexPrefix, byte(status))
}

// GetAuditLogKey returns the key for an audit log entry
func GetAuditLogKey(logID string) []byte {
	return append(AuditLogPrefix, []byte(logID)...)
}

// GetReportAuditLogsKey returns the prefix for audit logs of a specific report
func GetReportAuditLogsKey(reportID string) []byte {
	return append(AuditLogPrefix, append([]byte(reportID), '/')...)
}
