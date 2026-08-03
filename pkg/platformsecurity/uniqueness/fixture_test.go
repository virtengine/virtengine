package uniqueness

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func requestFixture(idempotency, label string) EnrollmentRequest {
	now := time.Now().Unix()
	return EnrollmentRequest{Version: Version1, RequestDigest: digestFixture("request-" + label), EvidenceDigest: digestFixture("evidence"), ModelDigest: digestFixture("model"), RuntimeDigest: digestFixture("runtime"), ProfileDigest: digestFixture("profile"), ProgramIDDigest: digestFixture("program"), PolicyDigest: digestFixture("policy"), IdempotencyKey: idempotency, KeyEpoch: 7, Freshness: Freshness{IssuedAt: now - 1, ExpiresAt: now + 60, Nonce: "nonce-" + idempotency}}
}

func finalRecord(request EnrollmentRequest, artifact TemplateArtifact) EnrollmentRecord {
	return EnrollmentRecord{Version: Version1, RequestDigest: request.RequestDigest, EvidenceDigest: request.EvidenceDigest, ModelDigest: request.ModelDigest, RuntimeDigest: request.RuntimeDigest, ProfileDigest: request.ProfileDigest, ProgramIDDigest: request.ProgramIDDigest, PolicyDigest: request.PolicyDigest, IdempotencyKey: request.IdempotencyKey, KeyEpoch: request.KeyEpoch, Freshness: request.Freshness, Outcome: OutcomeFinalUnique, Reason: ReasonFinalUnique, TemplateCommitment: artifact.Commitment(), ScopedNullifier: digestFixture("nullifier-" + request.IdempotencyKey)}
}

func enrollFinal(ctx context.Context, store AtomicEnrollmentStore, request EnrollmentRequest, artifact TemplateArtifact, threshold int64) (EnrollmentRecord, error) {
	return store.Enroll(ctx, request, func(tx EnrollmentTransaction) (EnrollmentRecord, error) {
		if _, err := tx.Search(ctx, artifact, threshold); err != nil {
			return EnrollmentRecord{}, err
		}
		if err := tx.Advance(StageAdjudication); err != nil {
			return EnrollmentRecord{}, err
		}
		if err := tx.Advance(StageNullifier); err != nil {
			return EnrollmentRecord{}, err
		}
		record := finalRecord(request, artifact)
		if err := tx.Insert(record, artifact); err != nil {
			return EnrollmentRecord{}, err
		}
		return record, nil
	})
}

func templateArtifactFixture(t *testing.T, input string) TemplateArtifact {
	t.Helper()
	transformer, err := NewDeterministicTemplateFixture([]byte("fixture-transform-key-32-bytes!!"))
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := transformer.Transform(context.Background(), profileFixture(), []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}

func finalAuthorizationFixture(t *testing.T, nodes CustodyNodeSet, keys map[string]ed25519.PrivateKey, now time.Time) (NullifierInput, FinalUniqueAuthorization) {
	t.Helper()
	setDigest, _ := nodes.Digest()
	input := NullifierInput{Version: Version1, Domain: NullifierDomain{Version: Version1, Kind: ExactProgramDomain, ProgramIDDigest: digestFixture("program")}, PolicyDigest: digestFixture("policy"), ProfileDigest: digestFixture("profile"), KeyEpoch: 7, StableInput: digestFixture("stable")}
	authorization := FinalUniqueAuthorization{Version: Version1, RequestDigest: digestFixture("request"), ProgramDigest: input.Domain.ProgramIDDigest, PolicyDigest: input.PolicyDigest, ProfileDigest: input.ProfileDigest, StableInputCommitment: input.StableInput, Decision: OutcomeFinalUnique, KeyEpoch: 7, ProfileEpoch: 3, Freshness: Freshness{IssuedAt: now.Unix() - 1, ExpiresAt: now.Unix() + 60, Nonce: "auth-nonce"}, ParticipantSetDigest: setDigest, Threshold: nodes.Threshold}
	attestor := DeterministicQuorumAttestorFixture{Nodes: nodes, PrivateKeys: keys}
	attestation, err := attestor.AttestFinalUnique(context.Background(), authorization)
	if err != nil {
		t.Fatal(err)
	}
	authorization.Attestation = attestation
	return input, authorization
}

func TestThresholdPRFRequiresFinalUniqueAuthorization(t *testing.T) {
	now := time.Unix(1000, 0)
	nodes, keys := custodyFixture(t)
	input, authorization := finalAuthorizationFixture(t, nodes, keys, now)
	prf, err := NewDeterministicThresholdPRFFixture([]byte("fixture-threshold-prf-key-32bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prf.Evaluate(context.Background(), input, nil); err == nil {
		t.Fatal("PRF evaluated without authorization")
	}
	verified, err := VerifyFinalUniqueAuthorization(authorization, nodes, now, 2*time.Minute, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	first, err := prf.Evaluate(context.Background(), input, verified)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := prf.Evaluate(context.Background(), input, verified)
	if first != again {
		t.Fatal("fixture PRF is unstable")
	}
	mutations := []func(*NullifierInput){
		func(i *NullifierInput) { i.Domain.ProgramIDDigest = digestFixture("other-program") },
		func(i *NullifierInput) { i.PolicyDigest = digestFixture("other-policy") },
		func(i *NullifierInput) { i.ProfileDigest = digestFixture("other-profile") },
		func(i *NullifierInput) { i.KeyEpoch++ },
		func(i *NullifierInput) { i.StableInput = digestFixture("other-stable") },
	}
	for index, mutate := range mutations {
		candidate := input
		mutate(&candidate)
		if _, err := prf.Evaluate(context.Background(), candidate, verified); err == nil {
			t.Fatalf("input mismatch %d accepted", index)
		}
	}
	attacks := []func(*FinalUniqueAuthorization){
		func(a *FinalUniqueAuthorization) { a.Decision = OutcomeDuplicateConfirmed },
		func(a *FinalUniqueAuthorization) { a.KeyEpoch++ },
		func(a *FinalUniqueAuthorization) { a.ProgramDigest = digestFixture("other-program") },
		func(a *FinalUniqueAuthorization) { a.Attestation.Signatures[0].Signature[0] ^= 1 },
		func(a *FinalUniqueAuthorization) {
			a.Freshness = Freshness{IssuedAt: 800, ExpiresAt: 900, Nonce: "stale"}
		},
	}
	for index, attack := range attacks {
		candidate := authorization
		candidate.Attestation.Signatures = append([]NodeSignature(nil), authorization.Attestation.Signatures...)
		for i := range candidate.Attestation.Signatures {
			candidate.Attestation.Signatures[i].Signature = append([]byte(nil), candidate.Attestation.Signatures[i].Signature...)
		}
		attack(&candidate)
		if _, err := VerifyFinalUniqueAuthorization(candidate, nodes, now, 2*time.Minute, time.Second); err == nil {
			t.Fatalf("authorization attack %d accepted", index)
		}
	}
	if prf.ProductionReady() || prf.FixtureState() != FixtureOnlyState {
		t.Fatal("fixture PRF claimed production readiness")
	}
}

func TestCandidateBoundaryIsOpaqueAndDeterministic(t *testing.T) {
	first := templateArtifactFixture(t, "opaque-a")
	second := templateArtifactFixture(t, "opaque-b")
	searcher := &DeterministicCandidateSearcherFixture{}
	if err := searcher.Add(digestFixture("record-a"), first); err != nil {
		t.Fatal(err)
	}
	result, err := searcher.Search(context.Background(), first, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !result.PossibleMatch || result.CandidateCount != 1 || result.ReviewState != CandidateReviewRequired || result.Validate() != nil {
		t.Fatalf("unexpected candidate result: %#v", result)
	}
	again, _ := searcher.Search(context.Background(), first, 0)
	if result != again {
		t.Fatal("candidate result is not deterministic")
	}
	clear, _ := searcher.Search(context.Background(), second, 0)
	if clear.PossibleMatch || clear.CandidateCount != 0 || clear.ReviewState != CandidateClear {
		t.Fatalf("unexpected clear result: %#v", clear)
	}
	encoded, _ := json.Marshal(result)
	lower := strings.ToLower(string(encoded))
	if strings.Contains(lower, "record-a") || strings.Contains(lower, "distance") || strings.Contains(lower, digestFixture("record-a")) {
		t.Fatal("candidate result leaked fixture internals")
	}
	if _, ok := any(first).(interface{ Bytes() []byte }); ok {
		t.Fatal("template artifact exposes bytes")
	}
}

func TestAtomicStoreIdempotencyPoisonAndRollback(t *testing.T) {
	artifact := templateArtifactFixture(t, "opaque")
	request := requestFixture("idem", "a")
	store := NewMemoryAtomicEnrollmentFixture()
	current := time.Unix(2000, 0)
	store.now = func() time.Time { return current }
	request.Freshness = Freshness{IssuedAt: current.Unix() - 1, ExpiresAt: current.Unix() + 30, Nonce: "nonce-idem"}
	first, err := enrollFinal(context.Background(), store, request, artifact, 0)
	if err != nil {
		t.Fatal(err)
	}
	current = current.Add(time.Minute)
	var called atomic.Bool
	retry, err := store.Enroll(context.Background(), request, func(EnrollmentTransaction) (EnrollmentRecord, error) {
		called.Store(true)
		return EnrollmentRecord{}, errors.New("must not run")
	})
	if err != nil || retry != first || called.Load() {
		t.Fatalf("exact retry failed: %v", err)
	}
	changed := request
	changed.RequestDigest = digestFixture("changed")
	if _, err := store.Enroll(context.Background(), changed, func(EnrollmentTransaction) (EnrollmentRecord, error) { return EnrollmentRecord{}, nil }); !errors.Is(err, ErrConflict) {
		t.Fatalf("changed digest did not conflict: %v", err)
	}
	for _, stage := range []EnrollmentStage{StageSearch, StageAdjudication, StageNullifier, StageInsertion} {
		t.Run("ignored-"+string(stage), func(t *testing.T) {
			rollback := NewMemoryAtomicEnrollmentFixture()
			rollback.InjectFailureAfter(stage)
			req := requestFixture("rollback-"+string(stage), string(stage))
			_, err := rollback.Enroll(context.Background(), req, func(tx EnrollmentTransaction) (EnrollmentRecord, error) {
				_, _ = tx.Search(context.Background(), artifact, 0)
				_ = tx.Advance(StageAdjudication)
				_ = tx.Advance(StageNullifier)
				record := finalRecord(req, artifact)
				_ = tx.Insert(record, artifact)
				return record, nil
			})
			if err == nil || rollback.Count() != 0 {
				t.Fatalf("ignored %s error committed", stage)
			}
		})
	}
}

func TestAtomicStoreCancellationBoundaries(t *testing.T) {
	artifact := templateArtifactFixture(t, "cancel")
	for _, boundary := range []string{"before", "after-search", "after-insert"} {
		t.Run(boundary, func(t *testing.T) {
			store := NewMemoryAtomicEnrollmentFixture()
			ctx, cancel := context.WithCancel(context.Background())
			request := requestFixture("cancel-"+boundary, boundary)
			if boundary == "before" {
				cancel()
			}
			_, err := store.Enroll(ctx, request, func(tx EnrollmentTransaction) (EnrollmentRecord, error) {
				_, _ = tx.Search(ctx, artifact, 0)
				if boundary == "after-search" {
					cancel()
				}
				_ = tx.Advance(StageAdjudication)
				_ = tx.Advance(StageNullifier)
				record := finalRecord(request, artifact)
				_ = tx.Insert(record, artifact)
				if boundary == "after-insert" {
					cancel()
				}
				return record, nil
			})
			if err == nil || store.Count() != 0 {
				t.Fatalf("cancellation at %s committed", boundary)
			}
		})
	}
}

func TestAtomicStoreConcurrentDuplicatesCommitAtMostOne(t *testing.T) {
	store := NewMemoryAtomicEnrollmentFixture()
	artifact := templateArtifactFixture(t, "same")
	const workers = 16
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(workers)
	for index := 0; index < workers; index++ {
		go func(index int) {
			defer wait.Done()
			request := requestFixture("idem-"+string(rune('a'+index)), string(rune('a'+index)))
			if _, err := enrollFinal(context.Background(), store, request, artifact, 0); err == nil {
				successes.Add(1)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || store.Count() != 1 {
		t.Fatalf("duplicate race committed successes=%d records=%d", successes.Load(), store.Count())
	}
}
