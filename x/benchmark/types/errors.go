// Package types contains types for the Benchmark module.
package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Error codes for the benchmark module
var (
	// ErrInvalidBenchmark is returned when a benchmark report is invalid
	ErrInvalidBenchmark = sdkerrors.Register(ModuleName, 1, "invalid benchmark report")

	// ErrUnknownProvider is returned when the provider is not registered
	ErrUnknownProvider = sdkerrors.Register(ModuleName, 2, "unknown provider")

	// ErrInvalidSignature is returned when the signature is invalid
	ErrInvalidSignature = sdkerrors.Register(ModuleName, 3, "invalid signature")

	// ErrDuplicateReport is returned when a duplicate report is submitted
	ErrDuplicateReport = sdkerrors.Register(ModuleName, 4, "duplicate benchmark report")

	// ErrOutOfBounds is returned when metric values are out of bounds
	ErrOutOfBounds = sdkerrors.Register(ModuleName, 5, "metric values out of bounds")

	// ErrChallengeNotFound is returned when a challenge is not found
	ErrChallengeNotFound = sdkerrors.Register(ModuleName, 6, "challenge not found")

	// ErrChallengeExpired is returned when a challenge has expired
	ErrChallengeExpired = sdkerrors.Register(ModuleName, 7, "challenge expired")

	// ErrInvalidSuiteVersion is returned when the suite version is invalid
	ErrInvalidSuiteVersion = sdkerrors.Register(ModuleName, 8, "invalid suite version")

	// ErrChallengeAlreadyResponded is returned when a challenge has already been responded to
	ErrChallengeAlreadyResponded = sdkerrors.Register(ModuleName, 9, "challenge already responded")

	// ErrUnauthorized is returned when the action is not authorized
	ErrUnauthorized = sdkerrors.Register(ModuleName, 10, "unauthorized")

	// ErrReportNotFound is returned when a benchmark report is not found
	ErrReportNotFound = sdkerrors.Register(ModuleName, 11, "benchmark report not found")

	// ErrProviderFlagged is returned when a provider is flagged
	ErrProviderFlagged = sdkerrors.Register(ModuleName, 12, "provider is flagged")

	// ErrInvalidMetricSchema is returned when the metric schema version is invalid
	ErrInvalidMetricSchema = sdkerrors.Register(ModuleName, 13, "invalid metric schema version")

	// ErrScoreNotFound is returned when a reliability score is not found
	ErrScoreNotFound = sdkerrors.Register(ModuleName, 14, "reliability score not found")
)
