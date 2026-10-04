// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Regression test for the CPU-architecture divergence in the golden-vector
// input hashes. The VEID conformance workflow computes these hashes on
// linux/amd64, windows/amd64 and darwin/arm64 and requires all three artifacts
// to agree exactly, so the values are pinned here as literals.

package inference

import "testing"

// canonicalGoldenInputHashes are the per-vector input hashes produced by the
// reference platform, taken verbatim from the `conformance-output.txt` evidence
// artifact of virtengine/virtengine run 35366514684
// (job 105672065639, "Cross-platform hash mismatch for
// conformance-evidence-linux-amd64").
//
// They are the values every platform must reproduce; a mismatch means a
// validator would compute a different input hash from the same inputs, which is
// a consensus hazard and not merely a CI annoyance.
var canonicalGoldenInputHashes = map[string]string{
	"golden_high_quality_v1":   "5946b3adb8d8be0fdf3bc2df5bbeea641e359a469b3d49431f4cc8332bea6a3e",
	"golden_medium_quality_v1": "edd131dc625fff493f31b1dbd31d8a4164ac3517c8c9dda4fa3029d807f0fc56",
	"golden_low_quality_v1":    "102a693277cb4ad12a26b4d3070f4ae8d1a44888a165f315ed0a3ff842c02a64",
	"golden_boundary_v1":       "68f0c3a3f56369edbb4eb78312df609775fa06e8943ea881681cfbc6df587524",
	"golden_perfect_v1":        "c3ab6c64d65ee2f38e353bc69401d44b2ade489e0712b086c2bace2b1d2214ab",
}

// TestGoldenVectorInputHashIsArchIndependent pins the golden-vector input hashes
// to the canonical values so that a backend-specific floating-point contraction
// cannot silently reintroduce a cross-architecture divergence.
//
// The original defect: generateDeterministicEmbedding computed
//
//	normalized := float32(state)/float32(m)*2*scale - scale
//
// a plain float32 `a*b + c` expression. The arm64 backend contracts that into a
// single FMADDS - one rounding for the whole `base*scale - scale` - while amd64
// emits MULSS followed by SUBSS - two roundings. Three of the five vectors have
// a non-zero scale and differed by ~250 of 512 embedding elements between the
// two architectures, which is what turned the cross-platform verifier red.
//
// This test passes on amd64 both before and after the fix, so it cannot be
// proven to have teeth by an amd64-only run. It is exercised under
// qemu-user-static as GOARCH=arm64, where the pre-fix code fails it and the
// fixed code passes.
func TestGoldenVectorInputHashIsArchIndependent(t *testing.T) {
	dc := NewDeterminismController(GoldenVectorSeed, true)

	seen := make(map[string]bool, len(canonicalGoldenInputHashes))

	for i := range GoldenVectors {
		vec := &GoldenVectors[i]

		want, ok := canonicalGoldenInputHashes[vec.ID]
		if !ok {
			t.Errorf("golden vector %q has no pinned canonical input hash; add one to canonicalGoldenInputHashes", vec.ID)
			continue
		}
		seen[vec.ID] = true

		t.Run(vec.ID, func(t *testing.T) {
			got := dc.ComputeInputHash(vec.Inputs)
			if got != want {
				t.Errorf("input hash for %s is not architecture-independent:\n got: %s\nwant: %s\n"+
					"A differing value means the embedding generator's arithmetic is being "+
					"contracted differently by this backend (e.g. arm64 FMADDS vs amd64 MULSS+SUBSS). "+
					"Do not re-pin this literal without evidence from all three conformance platforms.",
					vec.ID, got, want)
			}
		})
	}

	// Guard against a pinned vector being deleted or renamed: the map and the
	// vector set must stay in sync.
	for id := range canonicalGoldenInputHashes {
		if !seen[id] {
			t.Errorf("pinned canonical hash for %q does not correspond to any golden vector", id)
		}
	}
}
