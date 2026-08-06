// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coordinationclient "k8s.io/client-go/kubernetes/typed/coordination/v1"
)

// KubernetesSubmitterLease uses Kubernetes resourceVersion updates as the
// atomic ownership boundary. LeaseTransitions is the durable monotonically
// increasing fencing token; stale resource versions and stale tokens fail.
type KubernetesSubmitterLease struct {
	client  coordinationclient.LeasesGetter
	ns      string
	ownerID string
	now     func() time.Time
}

func NewKubernetesSubmitterLease(client coordinationclient.LeasesGetter, namespace, ownerID string) (*KubernetesSubmitterLease, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes coordination client is required")
	}
	if strings.TrimSpace(namespace) == "" {
		return nil, fmt.Errorf("kubernetes lease namespace is required")
	}
	if strings.TrimSpace(ownerID) == "" {
		return nil, fmt.Errorf("kubernetes lease owner ID is required")
	}
	return &KubernetesSubmitterLease{client: client, ns: namespace, ownerID: ownerID, now: func() time.Time { return time.Now().UTC() }}, nil
}

func (l *KubernetesSubmitterLease) Acquire(ctx context.Context, name string, ttl time.Duration) (uint64, error) {
	if err := validateLeaseArguments(ctx, name, ttl); err != nil {
		return 0, err
	}
	seconds, err := leaseDurationSeconds(ttl)
	if err != nil {
		return 0, err
	}
	leases := l.client.Leases(l.ns)
	resourceName := kubernetesLeaseName(name)
	now := metav1.NewMicroTime(l.now())
	current, err := leases.Get(ctx, resourceName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		one := int32(1)
		created, createErr := leases.Create(ctx, &coordinationv1.Lease{
			ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: l.ns, Labels: map[string]string{"app.kubernetes.io/managed-by": "virtengine-provider-daemon", "virtengine.com/fenced-lease": "true"}},
			Spec:       coordinationv1.LeaseSpec{HolderIdentity: stringPointer(l.ownerID), LeaseDurationSeconds: &seconds, AcquireTime: &now, RenewTime: &now, LeaseTransitions: &one},
		}, metav1.CreateOptions{})
		if createErr != nil {
			return 0, fmt.Errorf("acquire Kubernetes submitter lease: %w", createErr)
		}
		return leaseToken(created)
	}
	if err != nil {
		return 0, fmt.Errorf("read Kubernetes submitter lease: %w", err)
	}
	if kubernetesLeaseUnexpired(current, l.now()) {
		return 0, fmt.Errorf("submitter lease %s already held", name)
	}
	previous, err := leaseToken(current)
	if err != nil && !errors.Is(err, ErrSubmitterLeaseNotHeld) {
		return 0, err
	}
	if previous >= math.MaxInt32 {
		return 0, fmt.Errorf("kubernetes submitter fencing token exhausted")
	}
	next := int32(previous) + 1 // #nosec G115 -- bounded to math.MaxInt32 above.
	current.Spec.HolderIdentity = stringPointer(l.ownerID)
	current.Spec.LeaseDurationSeconds = &seconds
	current.Spec.AcquireTime = &now
	current.Spec.RenewTime = &now
	current.Spec.LeaseTransitions = &next
	updated, err := leases.Update(ctx, current, metav1.UpdateOptions{})
	if err != nil {
		return 0, fmt.Errorf("acquire Kubernetes submitter lease: %w", err)
	}
	return leaseToken(updated)
}

func (l *KubernetesSubmitterLease) Renew(ctx context.Context, name string, token uint64, ttl time.Duration) error {
	if err := validateLeaseArguments(ctx, name, ttl); err != nil {
		return err
	}
	seconds, err := leaseDurationSeconds(ttl)
	if err != nil {
		return err
	}
	leases := l.client.Leases(l.ns)
	current, err := leases.Get(ctx, kubernetesLeaseName(name), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("%w: %s", ErrSubmitterLeaseNotHeld, name)
	}
	currentToken, tokenErr := leaseToken(current)
	if tokenErr != nil || currentToken != token || !kubernetesLeaseUnexpired(current, l.now()) || pointerValue(current.Spec.HolderIdentity) != l.ownerID {
		return fmt.Errorf("%w: %s", ErrSubmitterLeaseNotHeld, name)
	}
	now := metav1.NewMicroTime(l.now())
	current.Spec.RenewTime = &now
	current.Spec.LeaseDurationSeconds = &seconds
	if _, err := leases.Update(ctx, current, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("renew Kubernetes submitter lease: %w", err)
	}
	return nil
}

func (l *KubernetesSubmitterLease) Release(ctx context.Context, name string, token uint64) error {
	leases := l.client.Leases(l.ns)
	current, err := leases.Get(ctx, kubernetesLeaseName(name), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	currentToken, tokenErr := leaseToken(current)
	if tokenErr != nil || currentToken != token || pointerValue(current.Spec.HolderIdentity) != l.ownerID {
		return nil
	}
	empty := ""
	now := metav1.NewMicroTime(l.now().Add(-time.Second))
	zero := int32(1)
	current.Spec.HolderIdentity = &empty
	current.Spec.RenewTime = &now
	current.Spec.LeaseDurationSeconds = &zero
	_, err = leases.Update(ctx, current, metav1.UpdateOptions{})
	return err
}

func (l *KubernetesSubmitterLease) Held(ctx context.Context, name string, token uint64) bool {
	current, err := l.client.Leases(l.ns).Get(ctx, kubernetesLeaseName(name), metav1.GetOptions{})
	if err != nil {
		return false
	}
	currentToken, err := leaseToken(current)
	return err == nil && currentToken == token && pointerValue(current.Spec.HolderIdentity) == l.ownerID && kubernetesLeaseUnexpired(current, l.now())
}

func kubernetesLeaseName(name string) string {
	digest := sha256.Sum256([]byte(name))
	return "provider-submit-" + hex.EncodeToString(digest[:8])
}

func kubernetesLeaseUnexpired(lease *coordinationv1.Lease, now time.Time) bool {
	if lease == nil || lease.Spec.RenewTime == nil || lease.Spec.LeaseDurationSeconds == nil || pointerValue(lease.Spec.HolderIdentity) == "" {
		return false
	}
	return now.Before(lease.Spec.RenewTime.Add(time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second))
}

func leaseDurationSeconds(ttl time.Duration) (int32, error) {
	seconds := int64((ttl + time.Second - 1) / time.Second)
	if seconds <= 0 || seconds > math.MaxInt32 {
		return 0, fmt.Errorf("kubernetes lease TTL is out of range")
	}
	return int32(seconds), nil
}

func leaseToken(lease *coordinationv1.Lease) (uint64, error) {
	if lease == nil || lease.Spec.LeaseTransitions == nil || *lease.Spec.LeaseTransitions <= 0 {
		return 0, ErrSubmitterLeaseNotHeld
	}
	return uint64(int64(*lease.Spec.LeaseTransitions)), nil // #nosec G115 -- positivity is checked above.
}

func stringPointer(value string) *string { return &value }

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

var _ SubmitterLease = (*KubernetesSubmitterLease)(nil)
