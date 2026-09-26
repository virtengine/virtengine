// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Store keys and prefixes
package types

import (
	"encoding/binary"
)

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

	// SequenceKeyFraudResponse is the sequence key for fraud responses
	SequenceKeyFraudResponse = []byte{0x22}

	// DedupIndexPrefix indexes a report's submission fingerprint per reporter so
	// an identical resubmission is rejected instead of creating a second queue entry.
	DedupIndexPrefix = []byte{0x07}

	// ResponsePrefix is the prefix for fraud response records
	ResponsePrefix = []byte{0x08}

	// ReportResponseIndexPrefix indexes responses by their parent report
	ReportResponseIndexPrefix = []byte{0x09}

	// ReporterActivityPrefix indexes a reporter's submission heights for the
	// per-reporter rate limit (block-height based, host-clock independent).
	ReporterActivityPrefix = []byte{0x0a}
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
	return append(AuditLogPrefix, []byte(reportID+"/")...)
}

// GetDedupIndexKey returns the dedup index key for a reporter + submission fingerprint
func GetDedupIndexKey(reporter, fingerprint string) []byte {
	return append(DedupIndexPrefix, []byte(reporter+"/"+fingerprint)...)
}

// GetResponseKey returns the key for a fraud response record
func GetResponseKey(responseID string) []byte {
	return append(ResponsePrefix, []byte(responseID)...)
}

// GetReportResponsesKey returns the prefix for responses belonging to a report
func GetReportResponsesKey(reportID string) []byte {
	return append(ReportResponseIndexPrefix, []byte(reportID+"/")...)
}

// GetReportResponseKey returns the index key for one response of a report
func GetReportResponseKey(reportID, responseID string) []byte {
	return append(GetReportResponsesKey(reportID), []byte(responseID)...)
}

// GetReporterActivityKey returns the activity index key for a reporter submission
func GetReporterActivityKey(reporter string, height int64, reportID string) []byte {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	key := append(ReporterActivityPrefix, []byte(reporter+"/")...)
	key = append(key, heightBytes...)
	return append(key, []byte("/"+reportID)...)
}

// GetReporterActivityPrefix returns the prefix for all of a reporter's submissions
func GetReporterActivityPrefix(reporter string) []byte {
	return append(ReporterActivityPrefix, []byte(reporter+"/")...)
}
