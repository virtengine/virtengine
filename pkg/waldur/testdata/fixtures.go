// Package testdata provides test fixtures for Waldur marketplace mapping tests.
//
// VE-303: Waldur bridge test fixtures for mapping validation
package testdata

import (
	"encoding/json"
	"time"
)

// Offering Export Fixtures

// VirtEngineOfferingJSON is an example VirtEngine offering payload.
const VirtEngineOfferingJSON = `{
  "id": {
    "provider_address": "ve1abc123xyz789def456ghi",
    "sequence": 42
  },
  "state": 1,
  "category": "compute",
  "name": "Standard Compute Instance",
  "description": "General-purpose compute instance with balanced CPU and memory",
  "version": "1.2.0",
  "pricing": {
    "model": "hourly",
    "base_price": 50000,
    "currency": "uvirt",
    "usage_rates": {},
    "minimum_commitment": 3600
  },
  "identity_requirement": {
    "min_score": 30,
    "required_status": "",
    "require_verified_email": true,
    "require_verified_domain": false,
    "require_mfa": false
  },
  "require_mfa_for_orders": false,
  "public_metadata": {
    "sla": "99.9%",
    "support": "24x7"
  },
  "specifications": {
    "vcpu": "4",
    "memory_gb": "16",
    "disk_gb": "100",
    "network": "1Gbps"
  },
  "tags": ["compute", "linux", "general-purpose"],
  "regions": ["us-east-1", "eu-west-1"],
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-01-20T14:30:00Z",
  "max_concurrent_orders": 100,
  "total_order_count": 1500,
  "active_order_count": 45
}`

// WaldurOfferingExportJSON is the expected Waldur offering export.
const WaldurOfferingExportJSON = `{
  "name": "Standard Compute Instance",
  "description": "General-purpose compute instance with balanced CPU and memory",
  "type": "Support.PerHour",
  "category": "uuid-of-compute-category",
  "shared": true,
  "billable": true,
  "attributes": {
    "ve_offering_id": "ve1abc123xyz789def456ghi/42",
    "ve_version": "1.2.0",
    "min_identity_score": 30,
    "require_mfa": false,
    "require_verified_email": true,
    "max_concurrent_orders": 100,
    "sla": "99.9%",
    "support": "24x7",
    "vcpu": "4",
    "memory_gb": "16",
    "disk_gb": "100",
    "network": "1Gbps",
    "tags": ["compute", "linux", "general-purpose"]
  },
  "components": [
    {
      "type": "usage",
      "name": "Hourly Rate",
      "measured_unit": "hour",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "0.050000"
    }
  ],
  "plans": [
    {
      "name": "Standard",
      "unit": "hour",
      "unit_price": "0.050000"
    }
  ],
  "locations": ["uuid-us-east-1", "uuid-eu-west-1"]
}`

// Order Export Fixtures

// VirtEngineOrderJSON is an example VirtEngine order payload.
const VirtEngineOrderJSON = `{
  "id": {
    "customer_address": "ve1customer789xyz456def123",
    "sequence": 101
  },
  "offering_id": {
    "provider_address": "ve1abc123xyz789def456ghi",
    "sequence": 42
  },
  "state": 5,
  "public_metadata": {
    "project_name": "ML Training Pipeline",
    "environment": "production"
  },
  "region": "us-east-1",
  "requested_quantity": 2,
  "allocated_provider_address": "ve1provider456xyz789abc123",
  "max_bid_price": 60000,
  "accepted_price": 50000,
  "created_at": "2026-01-25T09:00:00Z",
  "updated_at": "2026-01-25T09:15:00Z",
  "matched_at": "2026-01-25T09:05:00Z",
  "activated_at": "2026-01-25T09:15:00Z",
  "bid_count": 3
}`

// WaldurOrderExportJSON is the expected Waldur order export.
const WaldurOrderExportJSON = `{
  "offering": "uuid-of-offering",
  "project": "uuid-of-customer-project",
  "type": "Create",
  "attributes": {
    "name": "ML Training Pipeline",
    "description": "VirtEngine order ve1customer789xyz456def123/101",
    "ve_order_id": "ve1customer789xyz456def123/101",
    "region": "us-east-1",
    "environment": "production",
    "provider": "ve1provider456xyz789abc123"
  },
  "limits": {
    "instances": 2
  }
}`

// Callback Fixtures

// WaldurCallbackTerminateJSON is an example terminate callback from Waldur.
const WaldurCallbackTerminateJSON = `{
  "id": "wcb_ve1customer789xyz456def123_101_1_a1b2c3d4",
  "action_type": "terminate",
  "waldur_id": "550e8400-e29b-41d4-a716-446655440003",
  "chain_entity_type": "allocation",
  "chain_entity_id": "ve1customer789xyz456def123/101/1",
  "payload": {
    "reason": "customer_request",
    "requested_by": "user@example.com"
  },
  "signature": "YmFzZTY0LWVuY29kZWQtZWQyNTUxOS1zaWduYXR1cmU=",
  "signer_id": "waldur-bridge-signer-01",
  "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "timestamp": "2026-01-30T12:00:00Z",
  "expires_at": "2026-01-30T13:00:00Z"
}`

// WaldurCallbackProvisionJSON is an example provision callback from Waldur.
const WaldurCallbackProvisionJSON = `{
  "id": "wcb_ve1customer789xyz456def123_102_1_b2c3d4e5",
  "action_type": "provision",
  "waldur_id": "550e8400-e29b-41d4-a716-446655440004",
  "chain_entity_type": "allocation",
  "chain_entity_id": "ve1customer789xyz456def123/102/1",
  "payload": {
    "backend_id": "i-0abc123def456789",
    "external_ip": "203.0.113.45",
    "internal_ip": "10.0.1.25"
  },
  "signature": "YmFzZTY0LWVuY29kZWQtZWQyNTUxOS1zaWduYXR1cmU=",
  "signer_id": "waldur-bridge-signer-01",
  "nonce": "b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7",
  "timestamp": "2026-01-30T12:00:00Z",
  "expires_at": "2026-01-30T13:00:00Z"
}`

// WaldurCallbackUsageReportJSON is an example usage report callback from Waldur.
const WaldurCallbackUsageReportJSON = `{
  "id": "wcb_ve1customer789xyz456def123_101_1_c3d4e5f6",
  "action_type": "usage_report",
  "waldur_id": "550e8400-e29b-41d4-a716-446655440003",
  "chain_entity_type": "allocation",
  "chain_entity_id": "ve1customer789xyz456def123/101/1",
  "payload": {
    "cpu_hours": "720",
    "memory_gb_hours": "11520",
    "storage_gb_months": "100",
    "period_start": "2026-01-01T00:00:00Z",
    "period_end": "2026-01-31T23:59:59Z"
  },
  "signature": "YmFzZTY0LWVuY29kZWQtZWQyNTUxOS1zaWduYXR1cmU=",
  "signer_id": "waldur-bridge-signer-01",
  "nonce": "c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8",
  "timestamp": "2026-01-30T12:00:00Z",
  "expires_at": "2026-01-30T13:00:00Z"
}`

// State Mapping Fixtures

// OfferingStateMappings contains test cases for offering state mapping.
var OfferingStateMappings = []struct {
	VirtEngineState uint8
	WaldurState     string
}{
	{1, "Active"},
	{2, "Paused"},
	{3, "Archived"},
	{4, "Paused"},
	{5, "Archived"},
}

// OrderStateMappings contains test cases for order state mapping.
var OrderStateMappings = []struct {
	VirtEngineState uint8
	WaldurState     string
}{
	{1, "pending-consumer"},
	{2, "pending-provider"},
	{3, "executing"},
	{4, "executing"},
	{5, "done"},
	{6, "done"},
	{7, "terminating"},
	{8, "terminated"},
	{9, "erred"},
	{10, "canceled"},
}

// AllocationStateMappings contains test cases for allocation state mapping.
var AllocationStateMappings = []struct {
	VirtEngineState uint8
	WaldurState     string
}{
	{1, "Creating"},
	{2, "Creating"},
	{3, "Creating"},
	{4, "OK"},
	{5, "OK"},
	{6, "Terminating"},
	{7, "Terminated"},
	{8, "Erred"},
	{9, "Erred"},
}

// Pricing Fixtures

// PricingModelMappings contains test cases for pricing model mapping.
var PricingModelMappings = []struct {
	VirtEngineModel string
	WaldurType      string
	BillingType     string
	MeasuredUnit    string
}{
	{"hourly", "Support.PerHour", "usage", "hour"},
	{"daily", "Support.PerDay", "usage", "day"},
	{"monthly", "Support.Monthly", "fixed", "month"},
	{"usage_based", "Support.Usage", "usage", "unit"},
	{"fixed", "Support.OneTime", "one", "item"},
}

// PriceConversionCases contains test cases for price normalization.
var PriceConversionCases = []struct {
	ChainPrice uint64
	Decimals   int
	Expected   string
}{
	{50000, 6, "0.050000"},
	{1000000, 6, "1.000000"},
	{123456789, 6, "123.456789"},
	{100, 2, "1.000000"},
	{1, 6, "0.000001"},
}

// Category Mapping Fixtures

// CategoryMappings contains test cases for category to Waldur UUID mapping.
var CategoryMappings = map[string]string{
	"compute": "550e8400-e29b-41d4-a716-446655440100",
	"storage": "550e8400-e29b-41d4-a716-446655440101",
	"gpu":     "550e8400-e29b-41d4-a716-446655440102",
	"hpc":     "550e8400-e29b-41d4-a716-446655440103",
	"ml":      "550e8400-e29b-41d4-a716-446655440104",
	"network": "550e8400-e29b-41d4-a716-446655440105",
	"other":   "550e8400-e29b-41d4-a716-446655440199",
}

// Region Mapping Fixtures

// RegionMappings contains test cases for region to Waldur location UUID mapping.
var RegionMappings = map[string]string{
	"us-east-1":      "550e8400-e29b-41d4-a716-446655440200",
	"us-west-2":      "550e8400-e29b-41d4-a716-446655440201",
	"eu-west-1":      "550e8400-e29b-41d4-a716-446655440202",
	"eu-central-1":   "550e8400-e29b-41d4-a716-446655440203",
	"ap-northeast-1": "550e8400-e29b-41d4-a716-446655440204",
}

// Waldur API Response Fixtures

// WaldurOfferingResponseJSON is an example Waldur API offering response.
const WaldurOfferingResponseJSON = `{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Standard Compute Instance",
  "description": "General-purpose compute instance",
  "type": "Support.PerHour",
  "state": "Active",
  "shared": true,
  "billable": true,
  "created": "2026-01-15T10:00:00Z"
}`

// WaldurOrderResponseJSON is an example Waldur API order response.
const WaldurOrderResponseJSON = `{
  "uuid": "550e8400-e29b-41d4-a716-446655440001",
  "state": "done",
  "type": "Create",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440002",
  "created": "2026-01-25T09:00:00Z",
  "error_message": ""
}`

// WaldurResourceResponseJSON is an example Waldur API resource response.
const WaldurResourceResponseJSON = `{
  "uuid": "550e8400-e29b-41d4-a716-446655440003",
  "name": "test-resource-1",
  "state": "OK",
  "offering_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440002",
  "resource_type": "Support.PerHour",
  "created": "2026-01-25T09:15:00Z"
}`

// Sync Record Fixtures

// SyncRecordPendingJSON is an example pending sync record.
const SyncRecordPendingJSON = `{
  "entity_type": "offering",
  "entity_id": "ve1abc123xyz789def456ghi/42",
  "waldur_id": "",
  "state": 0,
  "sync_version": 0,
  "chain_version": 1,
  "failure_count": 0,
  "last_error": "",
  "checksum": ""
}`

// SyncRecordSyncedJSON is an example synced sync record.
const SyncRecordSyncedJSON = `{
  "entity_type": "offering",
  "entity_id": "ve1abc123xyz789def456ghi/42",
  "waldur_id": "550e8400-e29b-41d4-a716-446655440000",
  "state": 1,
  "sync_version": 1,
  "chain_version": 1,
  "last_synced_at": "2026-01-20T14:30:00Z",
  "failure_count": 0,
  "last_error": "",
  "checksum": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
}`

// SyncRecordFailedJSON is an example failed sync record.
const SyncRecordFailedJSON = `{
  "entity_type": "offering",
  "entity_id": "ve1abc123xyz789def456ghi/43",
  "waldur_id": "",
  "state": 2,
  "sync_version": 0,
  "chain_version": 1,
  "last_sync_attempt_at": "2026-01-20T14:30:00Z",
  "failure_count": 3,
  "last_error": "waldur server error: status 503",
  "checksum": ""
}`

// Helper Functions

// MustParseJSON parses JSON into a map for testing.
func MustParseJSON(data string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		panic(err)
	}
	return result
}

// MustParseTime parses an RFC3339 time string.
func MustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// TestUUIDs provides consistent test UUIDs.
var TestUUIDs = struct {
	Offering   string
	Order      string
	Resource   string
	Project    string
	Customer   string
	ComputeCat string
	USEast1    string
}{
	Offering:   "550e8400-e29b-41d4-a716-446655440000",
	Order:      "550e8400-e29b-41d4-a716-446655440001",
	Resource:   "550e8400-e29b-41d4-a716-446655440003",
	Project:    "550e8400-e29b-41d4-a716-446655440002",
	Customer:   "550e8400-e29b-41d4-a716-446655440010",
	ComputeCat: "550e8400-e29b-41d4-a716-446655440100",
	USEast1:    "550e8400-e29b-41d4-a716-446655440200",
}

// TestAddresses provides consistent test blockchain addresses.
var TestAddresses = struct {
	Provider1 string
	Provider2 string
	Customer1 string
	Customer2 string
}{
	Provider1: "ve1testprovider123abc456def789",
	Provider2: "ve1testprovider987zyx654wvu321",
	Customer1: "ve1testcustomer456def789abc123",
	Customer2: "ve1testcustomer321cba987zyx654",
}

// ErrorCases provides test cases for error handling.
var ErrorCases = []struct {
	Name           string
	HTTPStatusCode int
	ExpectedError  string
	ShouldRetry    bool
}{
	{"Unauthorized", 401, "unauthorized: check API token", false},
	{"Forbidden", 403, "forbidden: insufficient permissions", false},
	{"NotFound", 404, "resource not found", false},
	{"Conflict", 409, "resource conflict", false},
	{"RateLimited", 429, "rate limited", true},
	{"ServerError", 500, "waldur server error", true},
	{"ServiceUnavailable", 503, "waldur server error", true},
}
