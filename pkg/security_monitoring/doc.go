// Package security_monitoring provides comprehensive security-focused monitoring
// and threat detection for the VirtEngine platform.
//
// MONITOR-002: Security Monitoring & Threat Detection
//
// This package implements:
//   - Suspicious transaction pattern detection
//   - Identity verification fraud detection metrics
//   - Rate limiting breach alerts
//   - Cryptographic operation anomaly detection
//   - Provider daemon compromise detection integration
//   - Security event audit logging
//   - Automated incident response playbooks
//
// # Architecture
//
// The security monitoring system is designed around a central SecurityMonitor
// that orchestrates multiple specialized detectors:
//
//   - TransactionDetector: Analyzes transaction patterns for suspicious activity
//   - FraudDetector: Detects identity verification fraud indicators
//   - CryptoAnomalyDetector: Monitors cryptographic operations for anomalies
//   - ProviderSecurityMonitor: Integrates with provider daemon compromise detection
//
// All security events are logged to a structured AuditLog and can trigger
// automated incident response playbooks.
//
// # Usage
//
//	monitor := security_monitoring.NewSecurityMonitor(config, logger)
//	monitor.Start(ctx)
//	defer monitor.Stop()
//
//	// Record events
//	monitor.RecordTransaction(tx)
//	monitor.RecordVerification(veid)
//	monitor.RecordCryptoOp(op)
//
// # Metrics
//
// All metrics are exported via Prometheus with the "virtengine_security_" prefix.
// See security_metrics.go for the full list of available metrics.
package security_monitoring
