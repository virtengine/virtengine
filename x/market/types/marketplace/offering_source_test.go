package marketplace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func activeOffering(id string, seq uint64) *Offering {
	return &Offering{
		ID:                  OfferingID{ProviderAddress: id, Sequence: seq},
		State:               OfferingStateActive,
		Category:            OfferingCategoryCompute,
		Name:                "test",
		Pricing:             PricingInfo{Model: PricingModelHourly, Currency: "uvirt", BasePrice: 100},
		IdentityRequirement: DefaultIdentityRequirement(),
		Specifications:      map[string]string{"vcpu": "4", "memory_gb": "16"},
		Regions:             []string{"us-east-1"},
	}
}

func TestAdmitsOrderBasic(t *testing.T) {
	offering := activeOffering("ve1p", 1)
	require.True(t, offering.AdmitsOrder(AcquisitionModeDirect, nil))
	require.False(t, offering.AdmitsOrder(AcquisitionModeBid, nil), "bidding disabled by default")

	offering.AllowBidding = true
	require.True(t, offering.AdmitsOrder(AcquisitionModeBid, nil))
}

func TestAdmitsOrderAcquisitionModes(t *testing.T) {
	offering := activeOffering("ve1p", 1)
	offering.AcquisitionModes = []AcquisitionMode{AcquisitionModeBid}
	require.False(t, offering.AdmitsOrder(AcquisitionModeDirect, nil))
	require.True(t, offering.AdmitsOrder(AcquisitionModeBid, nil))
}

func TestAdmitsOrderSelector(t *testing.T) {
	offering := activeOffering("ve1p", 1)

	require.True(t, (&OfferSelector{
		Category: OfferingCategoryCompute,
		Regions:  []string{"us-east-1", "eu-west-1"},
		MinSpecs: map[string]uint64{"vcpu": 2, "memory_gb": 8},
	}).Matches(offering))

	require.False(t, (&OfferSelector{Category: OfferingCategoryStorage}).Matches(offering))
	require.False(t, (&OfferSelector{Regions: []string{"ap-south-1"}}).Matches(offering))
	require.False(t, (&OfferSelector{MinSpecs: map[string]uint64{"vcpu": 8}}).Matches(offering))
	require.False(t, (&OfferSelector{MinSpecs: map[string]uint64{"gpu": 1}}).Matches(offering))
}

func TestAdmitsOrderBackend(t *testing.T) {
	offering := activeOffering("ve1p", 1)
	offering.BackendType = BackendKubernetes
	require.True(t, (&OfferSelector{Backends: []string{BackendKubernetes}}).Matches(offering))
	require.False(t, (&OfferSelector{Backends: []string{BackendSLURM}}).Matches(offering))
}

func TestOfferingVisibilityDefaults(t *testing.T) {
	require.Equal(t, OfferingVisibilityPublic, OfferingVisibilityEmpty.Effective())
	require.True(t, OfferingVisibility("").IsPublic())
	require.True(t, OfferingVisibilityUnlisted.IsValid())
}

func TestOfferingValidationWaldurSource(t *testing.T) {
	offering := activeOffering("ve1p", 1)
	offering.Source = OfferingSourceWaldur
	require.Error(t, offering.Validate(), "waldur source requires a reference")

	offering.Waldur = &WaldurOfferingRef{OfferingUUID: "uuid-1"}
	require.NoError(t, offering.Validate())
}

func TestWaldurIngestToOfferingSetsSupplyMetadata(t *testing.T) {
	imp := &WaldurOfferingImport{
		UUID:         "uuid-1",
		InstanceID:   "waldur-eu",
		Name:         "SLURM Cluster",
		Description:  "desc",
		Type:         "VirtEngine.HPC",
		State:        "Active",
		CustomerUUID: "cust-1",
		Components: []WaldurPricingComponent{
			{Name: "cpu", Type: "usage", BillingType: "usage", MeasuredUnit: "cpu_hour", Price: "0.10"},
		},
		Attributes: map[string]interface{}{
			"ve_allow_bidding": true,
			"spec_cpu":         "128",
		},
	}

	cfg := DefaultIngestConfig()
	offering := imp.ToOfferingAt("ve1provider", 1, cfg, time.Unix(0, 0))

	require.Equal(t, OfferingSourceWaldur, offering.Source)
	require.Equal(t, OfferingVisibilityPublic, offering.Visibility)
	require.NotNil(t, offering.Waldur)
	require.Equal(t, "uuid-1", offering.Waldur.OfferingUUID)
	require.Equal(t, "waldur-eu", offering.Waldur.InstanceID)
	require.Equal(t, BackendSLURM, offering.BackendType)
	require.True(t, offering.AllowBidding)
	require.True(t, offering.SupportsAcquisitionMode(AcquisitionModeBid))
	require.True(t, offering.SupportsAcquisitionMode(AcquisitionModeDirect))
	require.NoError(t, offering.Validate())
}
