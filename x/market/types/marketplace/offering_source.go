// Package marketplace provides types for the marketplace on-chain module.
//
// This file implements the canonical supply-model extensions defined by
// _docs/adr/ADR-010-unified-market-resolution-and-waldur-supply.md:
// offering source, visibility, Waldur references, backend types, acquisition
// modes, and deterministic offer selectors.
package marketplace

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OfferingSource identifies where a supply listing originates.
type OfferingSource string

const (
	// OfferingSourceNative is a listing published directly by the provider.
	OfferingSourceNative OfferingSource = "native"

	// OfferingSourceWaldur is a listing mirrored from a Waldur instance.
	OfferingSourceWaldur OfferingSource = "waldur"

	// OfferingSourceEmpty mirrors the zero value for legacy offerings.
	OfferingSourceEmpty OfferingSource = ""
)

// IsValid returns true if the source is valid (empty is tolerated for backwards compatibility).
func (s OfferingSource) IsValid() bool {
	switch s {
	case OfferingSourceEmpty, OfferingSourceNative, OfferingSourceWaldur:
		return true
	default:
		return false
	}
}

// Effective returns the concrete source, treating the zero value as native.
func (s OfferingSource) Effective() OfferingSource {
	if s == OfferingSourceEmpty {
		return OfferingSourceNative
	}
	return s
}

// OfferingVisibility controls how a listing is exposed in the unified catalog.
type OfferingVisibility string

const (
	// OfferingVisibilityPublic listings appear in the public catalog.
	OfferingVisibilityPublic OfferingVisibility = "public"

	// OfferingVisibilityUnlisted listings are orderable but not browsable.
	OfferingVisibilityUnlisted OfferingVisibility = "unlisted"

	// OfferingVisibilityPrivate listings are only usable by explicit order.
	OfferingVisibilityPrivate OfferingVisibility = "private"

	// OfferingVisibilityEmpty mirrors the zero value for legacy offerings.
	OfferingVisibilityEmpty OfferingVisibility = ""
)

// IsValid returns true if the visibility is valid.
func (v OfferingVisibility) IsValid() bool {
	switch v {
	case OfferingVisibilityEmpty, OfferingVisibilityPublic, OfferingVisibilityUnlisted, OfferingVisibilityPrivate:
		return true
	default:
		return false
	}
}

// Effective returns the concrete visibility, treating the zero value as public.
func (v OfferingVisibility) Effective() OfferingVisibility {
	if v == OfferingVisibilityEmpty {
		return OfferingVisibilityPublic
	}
	return v
}

// IsPublic returns true when the listing is publicly browsable.
func (v OfferingVisibility) IsPublic() bool {
	return v.Effective() == OfferingVisibilityPublic
}

// AcquisitionMode is the commercial pathway used to acquire a listing.
type AcquisitionMode string

const (
	// AcquisitionModeDirect buys a listing at its published price.
	AcquisitionModeDirect AcquisitionMode = "direct"

	// AcquisitionModeBid opens a bidding window resolved by the engine.
	AcquisitionModeBid AcquisitionMode = "bid"
)

// IsValid returns true if the acquisition mode is valid.
func (m AcquisitionMode) IsValid() bool {
	return m == AcquisitionModeDirect || m == AcquisitionModeBid
}

// Backend type identifiers mirror the provider execution backends.
const (
	BackendKubernetes = "kubernetes"
	BackendOpenStack  = "openstack"
	BackendVMware     = "vmware"
	BackendAWS        = "aws"
	BackendAzure      = "azure"
	BackendSLURM      = "slurm"
	BackendMOAB       = "moab"
	BackendOOD        = "ood"
)

// WaldurOfferingRef links an on-chain offering to its Waldur source.
type WaldurOfferingRef struct {
	// InstanceID identifies the registered Waldur instance (source registry key).
	InstanceID string `json:"instance_id"`

	// OfferingUUID is the Waldur offering UUID.
	OfferingUUID string `json:"offering_uuid"`

	// CustomerUUID is the Waldur customer (provider organization) UUID.
	CustomerUUID string `json:"customer_uuid,omitempty"`

	// BackendType is the Waldur/default backend type for this offering.
	BackendType string `json:"backend_type,omitempty"`

	// SnapshotHash is the checksum of the ingested Waldur snapshot.
	SnapshotHash string `json:"snapshot_hash,omitempty"`

	// SnapshotHeight is the monotonic Waldur revision used for replay protection.
	SnapshotHeight uint64 `json:"snapshot_height,omitempty"`
}

// Validate validates the Waldur offering reference.
func (r *WaldurOfferingRef) Validate() error {
	if r == nil {
		return fmt.Errorf("waldur reference is required")
	}
	if strings.TrimSpace(r.OfferingUUID) == "" {
		return fmt.Errorf("waldur offering_uuid is required")
	}
	return nil
}

// OfferSelector is a bounded, deterministic predicate used to resolve an order
// against eligible supply. It intentionally avoids regular expressions and
// unbounded scans.
type OfferSelector struct {
	// Category restricts candidates to a single offering category.
	Category OfferingCategory `json:"category,omitempty"`

	// Regions restricts candidates to listings serving at least one region.
	Regions []string `json:"regions,omitempty"`

	// MinSpecs requires the offering specification to meet or exceed each value.
	MinSpecs map[string]uint64 `json:"min_specs,omitempty"`

	// MaxPrice caps the total resolved price. Nil means no cap.
	MaxPrice *sdk.Coin `json:"max_price,omitempty"`

	// IdentityRequired requires the offering to declare an identity requirement.
	IdentityRequired bool `json:"identity_required,omitempty"`

	// Backends allow-lists provider backend types (empty means any).
	Backends []string `json:"backends,omitempty"`
}

// Validate validates the selector.
func (s *OfferSelector) Validate() error {
	if s == nil {
		return nil
	}
	if s.Category != "" {
		switch s.Category {
		case OfferingCategoryCompute, OfferingCategoryStorage, OfferingCategoryNetwork,
			OfferingCategoryHPC, OfferingCategoryGPU, OfferingCategoryML, OfferingCategoryOther:
		default:
			return fmt.Errorf("invalid selector category: %s", s.Category)
		}
	}
	if s.MaxPrice != nil {
		if !s.MaxPrice.IsValid() || !s.MaxPrice.Amount.IsPositive() {
			return fmt.Errorf("invalid selector max_price")
		}
	}
	for key := range s.MinSpecs {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("selector min_specs contains an empty key")
		}
	}
	return nil
}

// Matches reports whether an offering satisfies the selector. Price is not
// evaluated here; callers apply the price cap after quoting.
func (s *OfferSelector) Matches(offering *Offering) bool {
	if s == nil || offering == nil {
		return false
	}
	if !offering.State.IsAcceptingOrders() {
		return false
	}
	if s.Category != "" && offering.Category != s.Category {
		return false
	}
	if len(s.Regions) > 0 && !intersects(s.Regions, offering.Regions) {
		return false
	}
	for key, min := range s.MinSpecs {
		raw, ok := offering.Specifications[key]
		if !ok {
			return false
		}
		value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
		if err != nil || value < min {
			return false
		}
	}
	if s.IdentityRequired && offering.IdentityRequirement == (IdentityRequirement{}) {
		return false
	}
	if len(s.Backends) > 0 && !containsString(s.Backends, offering.EffectiveBackendType()) {
		return false
	}
	return true
}

// EffectiveBackendType returns the offering's backend type, inferring from the
// service specifications when unset.
func (o *Offering) EffectiveBackendType() string {
	if o == nil {
		return ""
	}
	if o.BackendType != "" {
		return o.BackendType
	}
	if ref := o.Waldur; ref != nil && ref.BackendType != "" {
		return ref.BackendType
	}
	switch ServiceTypeFromSpecs(o.Specifications) {
	case ServiceTypeContainer:
		return BackendKubernetes
	case ServiceTypeVM:
		return BackendOpenStack
	}
	if o.Category == OfferingCategoryHPC {
		return BackendSLURM
	}
	return ""
}

// SupportsAcquisitionMode reports whether the offering may be acquired using the
// given mode. When AcquisitionModes is unset, direct is always supported and bid
// is supported when AllowBidding is set.
func (o *Offering) SupportsAcquisitionMode(mode AcquisitionMode) bool {
	if o == nil {
		return false
	}
	if len(o.AcquisitionModes) > 0 {
		return containsMode(o.AcquisitionModes, mode)
	}
	switch mode {
	case AcquisitionModeDirect:
		return true
	case AcquisitionModeBid:
		return o.AllowBidding
	default:
		return false
	}
}

// AdmitsOrder reports whether the offering may be auto-resolved for an order
// with the given mode, visibility expectation, and selector.
func (o *Offering) AdmitsOrder(mode AcquisitionMode, selector *OfferSelector) bool {
	if o == nil {
		return false
	}
	if !o.State.IsAcceptingOrders() {
		return false
	}
	if !o.SupportsAcquisitionMode(mode) {
		return false
	}
	if err := o.CanAcceptOrder(); err != nil {
		return false
	}
	if selector == nil {
		return true
	}
	return selector.Matches(o)
}

func intersects(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

func containsString(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func containsMode(list []AcquisitionMode, value AcquisitionMode) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// NormalizeAcquisitionModes returns a sorted, de-duplicated list of modes.
func NormalizeAcquisitionModes(modes []AcquisitionMode) []AcquisitionMode {
	if len(modes) == 0 {
		return nil
	}
	seen := make(map[AcquisitionMode]struct{}, len(modes))
	out := make([]AcquisitionMode, 0, len(modes))
	for _, m := range modes {
		if !m.IsValid() {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	if len(out) == 0 {
		return nil
	}
	return out
}
