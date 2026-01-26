// Package keeper implements the Fraud module keeper.
//
// VE-912: Fraud reporting flow - Store keys for keeper
package keeper

import (
	"pkg.akt.dev/node/x/fraud/types"
)

// Key prefixes for store operations
var (
	FraudReportPrefix        = types.FraudReportPrefix
	ModeratorQueuePrefix     = types.ModeratorQueuePrefix
	ReporterIndexPrefix      = types.ReporterIndexPrefix
	ReportedPartyIndexPrefix = types.ReportedPartyIndexPrefix
	StatusIndexPrefix        = types.StatusIndexPrefix
	AuditLogPrefix           = types.AuditLogPrefix
	ParamsKey                = types.ParamsKey
	SequenceKeyFraudReport   = types.SequenceKeyFraudReport
	SequenceKeyAuditLog      = types.SequenceKeyAuditLog
)

// FraudReportKey returns the key for a fraud report
func FraudReportKey(reportID string) []byte {
	return types.GetFraudReportKey(reportID)
}

// ModeratorQueueKey returns the key for a moderator queue entry
func ModeratorQueueKey(reportID string) []byte {
	return types.GetModeratorQueueKey(reportID)
}

// ReporterIndexKey returns the index key for a reporter
func ReporterIndexKey(reporterAddr string) []byte {
	return types.GetReporterIndexKey(reporterAddr)
}

// ReportedPartyIndexKey returns the index key for a reported party
func ReportedPartyIndexKey(reportedAddr string) []byte {
	return types.GetReportedPartyIndexKey(reportedAddr)
}

// StatusIndexKey returns the index key for a status
func StatusIndexKey(status types.FraudReportStatus) []byte {
	return types.GetStatusIndexKey(status)
}

// AuditLogKey returns the key for an audit log entry
func AuditLogKey(logID string) []byte {
	return types.GetAuditLogKey(logID)
}

// ReportAuditLogsKey returns the prefix for audit logs of a specific report
func ReportAuditLogsKey(reportID string) []byte {
	return types.GetReportAuditLogsKey(reportID)
}
