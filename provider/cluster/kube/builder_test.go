package kube

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/virtengine/virtengine/manifest"
	"github.com/virtengine/virtengine/testutil"
	"github.com/stretchr/testify/assert"
)

func TestLidNsSanity(t *testing.T) {
	log := testutil.Logger(t)
	leaseID := testutil.LeaseID(t)

	ns := lidNS(leaseID)
	assert.NotEmpty(t, ns)

	// namespaces must be no more than 63 characters.
	assert.Less(t, len(ns), int(64))
	settings := NewDefaultSettings()
	g := &manifest.Group{}

	b := builder{
		log:      log,
		settings: settings,
		lid:      leaseID,
		group:    g,
	}

	mb := newManifestBuilder(log, settings, ns, leaseID, g)
	assert.Equal(t, b.ns(), mb.ns())

	m, err := mb.create()
	assert.NoError(t, err)
	assert.Equal(t, m.Spec.LeaseID.DSeq, strconv.FormatUint(leaseID.DSeq, 10))

	assert.Equal(t, ns, m.Name)
}

func TestNetworkPolicies(t *testing.T) {
	leaseID := testutil.LeaseID(t)

	g := &manifest.Group{}
	settings := NewDefaultSettings()
	np := newNetPolBuilder(NewDefaultSettings(), leaseID, g)

	// disabled
	netPolicies, err := np.create()
	assert.NoError(t, err)
	assert.Len(t, netPolicies, 0)

	// enabled
	settings.NetworkPoliciesEnabled = true
	np = newNetPolBuilder(settings, leaseID, g)
	netPolicies, err = np.create()
	assert.NoError(t, err)
	assert.Len(t, netPolicies, 1)

	pol0 := netPolicies[0]
	assert.Equal(t, pol0.Name, "virtengine-deployment-restrictions")

	// Change the DSeq ID
	np.lid.DSeq = uint64(100)
	k := virtengineNetworkNamespace
	ns := lidNS(np.lid)
	updatedNetPol, err := np.update(netPolicies[0])
	assert.NoError(t, err)
	updatedNS := updatedNetPol.Labels[k]
	assert.Equal(t, ns, updatedNS)
}

func TestGlobalServiceBuilder(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	service := &manifest.Service{
		Name: "myservice",
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, true)
	require.NotNil(t, serviceBuilder)
	// Should have name ending with suffix
	require.Equal(t, "myservice-np", serviceBuilder.name())
	// Should not have any work to do
	require.False(t, serviceBuilder.any())
}

func TestLocalServiceBuilder(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	service := &manifest.Service{
		Name: "myservice",
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, false)
	require.NotNil(t, serviceBuilder)
	// Should have name verbatim
	require.Equal(t, "myservice", serviceBuilder.name())
	// Should not have any work to do
	require.False(t, serviceBuilder.any())
}

func TestGlobalServiceBuilderWithoutGlobalServices(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	exposesServices := make([]manifest.ServiceExpose, 1)
	exposesServices[0].Global = false
	service := &manifest.Service{
		Name:   "myservice",
		Expose: exposesServices,
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, true)

	// Should not have any work to do
	require.False(t, serviceBuilder.any())
}

func TestGlobalServiceBuilderWithGlobalServices(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	exposesServices := make([]manifest.ServiceExpose, 2)
	exposesServices[0] = manifest.ServiceExpose{
		Global:       true,
		Proto:        "TCP",
		Port:         1000,
		ExternalPort: 1001,
	}
	exposesServices[1] = manifest.ServiceExpose{
		Global:       false,
		Proto:        "TCP",
		Port:         2000,
		ExternalPort: 2001,
	}
	service := &manifest.Service{
		Name:   "myservice",
		Expose: exposesServices,
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, true)

	// Should have work to do
	require.True(t, serviceBuilder.any())

	result, err := serviceBuilder.create()
	require.NoError(t, err)
	require.Equal(t, result.Spec.Type, corev1.ServiceTypeNodePort)
	ports := result.Spec.Ports
	require.Len(t, ports, 1)
	require.Equal(t, ports[0].Port, int32(1001))
	require.Equal(t, ports[0].TargetPort, intstr.FromInt(1000))
	require.Equal(t, ports[0].Name, "0-1001")
}

func TestLocalServiceBuilderWithoutLocalServices(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	exposesServices := make([]manifest.ServiceExpose, 1)
	exposesServices[0].Global = true
	service := &manifest.Service{
		Name:   "myservice",
		Expose: exposesServices,
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, false)

	// Should have work to do
	require.False(t, serviceBuilder.any())
}

func TestLocalServiceBuilderWithLocalServices(t *testing.T) {
	myLog := testutil.Logger(t)
	group := &manifest.Group{}
	exposesServices := make([]manifest.ServiceExpose, 2)
	exposesServices[0] = manifest.ServiceExpose{
		Global:       true,
		Proto:        "TCP",
		Port:         1000,
		ExternalPort: 1001,
	}
	exposesServices[1] = manifest.ServiceExpose{
		Global:       false,
		Proto:        "TCP",
		Port:         2000,
		ExternalPort: 2001,
	}
	service := &manifest.Service{
		Name:   "myservice",
		Expose: exposesServices,
	}
	mySettings := NewDefaultSettings()
	lid := testutil.LeaseID(t)
	serviceBuilder := newServiceBuilder(myLog, mySettings, lid, group, service, false)

	// Should have work to do
	require.True(t, serviceBuilder.any())

	result, err := serviceBuilder.create()
	require.NoError(t, err)
	require.Equal(t, result.Spec.Type, corev1.ServiceTypeClusterIP)
	ports := result.Spec.Ports
	require.Equal(t, ports[0].Port, int32(2001))
	require.Equal(t, ports[0].TargetPort, intstr.FromInt(2000))
	require.Equal(t, ports[0].Name, "1-2001")
}
