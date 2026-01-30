#!/bin/bash -eu
# Copyright 2024 VirtEngine Authors
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# OSS-Fuzz build script for VirtEngine
# This script compiles all fuzz targets for OSS-Fuzz infrastructure
#
# Environment variables set by OSS-Fuzz:
#   $SRC - Source directory
#   $OUT - Output directory for fuzzer binaries
#   $WORK - Working directory
#   $FUZZING_ENGINE - libfuzzer, afl, honggfuzz, etc.
#   $LIB_FUZZING_ENGINE - Libraries to link against
#   $ARCHITECTURE - x86_64, i386, etc.
#   $SANITIZER - address, memory, undefined, coverage
#
# Task Reference: QUALITY-002 - Fuzz Testing Implementation

cd $SRC/virtengine

# Ensure go modules are available
go mod download

# Compile native Go fuzz targets using the compile_native_go_fuzzer helper
# This creates instrumented binaries compatible with libFuzzer

# Encryption types fuzz targets
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzEnvelopeValidate fuzz_envelope_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzEnvelopeSigningPayload fuzz_envelope_signing_payload
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzEnvelopeHash fuzz_envelope_hash
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzEnvelopeDeterministicBytes fuzz_envelope_deterministic_bytes
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzRecipientKeyRecordValidate fuzz_recipient_key_record_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzComputeKeyFingerprint fuzz_compute_key_fingerprint
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/types FuzzAlgorithmValidation fuzz_algorithm_validation

# Encryption crypto fuzz targets (existing)
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzCreateEnvelope fuzz_create_envelope
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzOpenEnvelope fuzz_open_envelope
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzNonceUniqueness fuzz_nonce_uniqueness
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzMultiRecipientEnvelope fuzz_multi_recipient_envelope
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzInvalidKeySize fuzz_invalid_key_size
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzKeyPairGeneration fuzz_key_pair_generation
compile_native_go_fuzzer github.com/virtengine/virtengine/x/encryption/crypto FuzzAlgorithmEncryption fuzz_algorithm_encryption

# VEID types fuzz targets
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzIdentityScopeValidate fuzz_identity_scope_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzScopeTypeValidation fuzz_scope_type_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzVerificationStatusValidation fuzz_verification_status_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzVerificationStatusTransitions fuzz_verification_status_transitions
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzUploadMetadataValidate fuzz_upload_metadata_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzVerificationEventValidate fuzz_verification_event_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzSimpleVerificationResultValidate fuzz_simple_verification_result_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzApprovedClientValidate fuzz_approved_client_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/veid/types FuzzComputeSaltHash fuzz_compute_salt_hash

# Marketplace types fuzz targets
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOrderValidate fuzz_order_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOrderIDValidate fuzz_order_id_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOrderStateTransitions fuzz_order_state_transitions
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOfferingValidate fuzz_offering_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOfferingIDValidate fuzz_offering_id_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzAllocationIDValidate fuzz_allocation_id_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzIdentityRequirementValidate fuzz_identity_requirement_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzPricingInfoValidate fuzz_pricing_info_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOrderSetState fuzz_order_set_state
compile_native_go_fuzzer github.com/virtengine/virtengine/x/market/types/marketplace FuzzOrderCanAcceptBid fuzz_order_can_accept_bid

# Provider daemon fuzz targets
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzManifestParse fuzz_manifest_parse
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzManifestValidate fuzz_manifest_validate
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzServiceSpecValidation fuzz_service_spec_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzPortSpecValidation fuzz_port_spec_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzVolumeSpecValidation fuzz_volume_spec_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzNetworkSpecValidation fuzz_network_spec_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzHealthCheckSpecValidation fuzz_health_check_spec_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzConstraintsValidation fuzz_constraints_validation
compile_native_go_fuzzer github.com/virtengine/virtengine/pkg/provider_daemon FuzzManifestTotalResources fuzz_manifest_total_resources

# Copy seed corpus directories to output
# OSS-Fuzz expects corpus in $OUT/<fuzzer_name>_seed_corpus.zip
for corpus_dir in fuzz/corpus/*; do
    if [ -d "$corpus_dir" ]; then
        fuzzer_name=$(basename "$corpus_dir")
        zip -j "$OUT/${fuzzer_name}_seed_corpus.zip" "$corpus_dir"/* 2>/dev/null || true
    fi
done

# Copy dictionary files
cp fuzz/dictionaries/*.dict "$OUT/" 2>/dev/null || true

echo "OSS-Fuzz build completed successfully"
echo "Built $(ls -1 $OUT/fuzz_* 2>/dev/null | wc -l) fuzz targets"
