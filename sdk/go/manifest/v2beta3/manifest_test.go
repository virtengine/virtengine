package v2beta3

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	resources "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

const (
	nameOfTestService = "test-service"
	nameOfTestGroup   = "testGroup"
)

var (
	randCPU1    = uint64(testutil.RandCPUUnits())
	randCPU2    = randCPU1 + 1
	randGPU1    = uint64(testutil.RandGPUUnits())
	randMemory  = testutil.RandMemoryQuantity()
	randStorage = testutil.RandStorageQuantity()
)

var randUnits1 = resources.Resources{
	ID: 1,
	CPU: &resources.CPU{
		Units: resources.NewResourceValue(randCPU1),
	},
	GPU: &resources.GPU{
		Units: resources.NewResourceValue(randGPU1),
	},
	Memory: &resources.Memory{
		Quantity: resources.NewResourceValue(randMemory),
	},
	Storage: resources.Volumes{
		resources.Storage{
			Quantity: resources.NewResourceValue(randStorage),
		},
	},
	Endpoints: resources.Endpoints{},
}

var randUnits3 = resources.Resources{
	ID: 1,
	CPU: &resources.CPU{
		Units: resources.NewResourceValue(randCPU2),
	},
	Memory: &resources.Memory{
		Quantity: resources.NewResourceValue(randMemory),
	},
	Storage: resources.Volumes{
		resources.Storage{
			Quantity: resources.NewResourceValue(randStorage),
		},
	},
	Endpoints: resources.Endpoints{},
}

func simpleResources(exposes ServiceExposes) resources.Resources {
	return resources.Resources{
		ID: 1,
		CPU: &resources.CPU{
			Units: resources.ResourceValue{
				Val: sdkmath.NewIntFromUint64(randCPU1),
			},
			Attributes: nil,
		},
		Memory: &resources.Memory{
			Quantity: resources.ResourceValue{
				Val: sdkmath.NewIntFromUint64(randMemory),
			},
			Attributes: nil,
		},
		GPU: &resources.GPU{
			Units: resources.ResourceValue{
				Val: sdkmath.NewIntFromUint64(randGPU1),
			},
			Attributes: nil,
		},
		Storage: resources.Volumes{
			resources.Storage{
				Name: "default",
				Quantity: resources.ResourceValue{
					Val: sdkmath.NewIntFromUint64(randStorage),
				},
			},
		},
		Endpoints: exposes.GetEndpoints(),
	}
}

func TestNilManifestIsInvalid(t *testing.T) {
	m := Manifest{}
	err := m.Validate()

	require.Error(t, err)
	require.Regexp(t, "^.*manifest is empty.*$", err)
}

func simpleManifest(svcCount uint32) Manifest {
	expose := make([]ServiceExpose, 1)
	expose[0].Global = true
	expose[0].Port = 80
	expose[0].Proto = TCP
	expose[0].Hosts = make([]string, 1)
	expose[0].Hosts[0] = "host.test"

	services := make([]Service, 1)
	services[0] = Service{
		Name:      nameOfTestService,
		Image:     "test/image:1.0",
		Command:   nil,
		Args:      nil,
		Env:       nil,
		Resources: simpleResources(expose),
		Count:     svcCount,
		Expose:    expose,
	}

	m := make(Manifest, 1)
	m[0] = Group{
		Name:     nameOfTestGroup,
		Services: services,
	}

	return m
}

func TestSimpleManifestIsValid(t *testing.T) {
	m := simpleManifest(1)
	err := m.Validate()
	require.NoError(t, err)
}

func TestSimpleManifestInvalidResourcesID(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Resources.ID = 0
	err := m.Validate()
	require.Error(t, err)
}

func TestManifestWithNoGlobalServicesIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Expose[0].Global = false
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*zero global services.*$", err)
}

func TestManifestWithBadServiceNameIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Name = "a_bad_service_name" // should not contain underscores
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*name is invalid.*$", err)

	m[0].Services[0].Name = "a-name-" // should not end with dash
	err = m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*name is invalid.*$", err)
}

func TestManifestWithServiceNameIsValid(t *testing.T) {
	m := simpleManifest(1)

	m[0].Services[0].Name = "9aaa-bar" // does not allow starting with a number
	err := m.Validate()
	require.ErrorIs(t, err, ErrInvalidManifest)
	require.Regexp(t, "^.*name is invalid.*$", err)
}

func TestManifestWithDuplicateHostIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	hosts := make([]string, 2)
	const hostname = "a.test"
	hosts[0] = hostname
	hosts[1] = hostname
	m[0].Services[0].Expose[0].Hosts = hosts
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*hostname.+is duplicated.*$", err)
}

func TestManifestWithDashInHostname(t *testing.T) {
	m := simpleManifest(1)
	hosts := make([]string, 1)
	hosts[0] = "a-test.com"
	m[0].Services[0].Expose[0].Hosts = hosts
	err := m.Validate()
	require.NoError(t, err)
}

func TestManifestWithBadHostIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	hosts := make([]string, 2)
	hosts[0] = "bob.test" // valid
	hosts[1] = "-bob"     // invalid
	m[0].Services[0].Expose[0].Hosts = hosts
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid hostname.*$", err)
}

func TestManifestWithLongHostIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	hosts := make([]string, 1)
	buf := &bytes.Buffer{}
	for i := 0; i != 255; i++ {
		_, err := buf.WriteRune('a')
		require.NoError(t, err)
	}
	_, err := buf.WriteString(".com")
	require.NoError(t, err)

	hosts[0] = buf.String()
	m[0].Services[0].Expose[0].Hosts = hosts
	err = m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid hostname.*$", err)
}

func TestManifestWithDuplicateGroupIsInvalid(t *testing.T) {
	mDuplicate := make(Manifest, 2)
	mDuplicate[0] = simpleManifest(1)[0]
	mDuplicate[1] = simpleManifest(1)[0]
	mDuplicate[1].Services[0].Expose[0].Hosts[0] = "anotherhost.test"
	err := mDuplicate.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*duplicate group.*$", err)
}

func TestManifestWithNoServicesInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services = nil
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*contains no services.*$", err)
}

func TestManifestWithEmptyServiceNameInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Name = ""
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*service name is empty.*$", err)
}

func TestManifestWithEmptyImageNameInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Image = ""
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, "^.*service.+has empty image name.*$", err)
}

func TestManifestWithEmptyEnvValueIsValid(t *testing.T) {
	m := simpleManifest(1)
	envVars := make([]string, 1)
	envVars[0] = "FOO=" // sets FOO to empty string
	m[0].Services[0].Env = envVars
	err := m.Validate()
	require.NoError(t, err)
}

func TestManifestWithEmptyEnvNameIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	envVars := make([]string, 1)
	envVars[0] = "=FOO" // invalid
	m[0].Services[0].Env = envVars
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, `^.*var\. with an invalid name.*$`, err)
}

func TestManifestWithBadEnvNameIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	envVars := make([]string, 1)
	envVars[0] = "9VAR=FOO" // invalid because it starts with a digit
	m[0].Services[0].Env = envVars
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, `^.*var\. with an invalid name.*$`, err)
}

func TestManifestServiceUnknownProtocolIsInvalid(t *testing.T) {
	m := simpleManifest(1)
	m[0].Services[0].Expose[0].Proto = "ICMP"
	err := m.Validate()
	require.Error(t, err)
	require.Regexp(t, `^.*protocol .+ unknown.*$`, err)
}

func Test_ValidateManifest(t *testing.T) {
	expose := make([]ServiceExpose, 1)
	expose[0].Global = true
	expose[0].Port = 80
	expose[0].Proto = TCP
	expose[0].Hosts = make([]string, 1)
	expose[0].Hosts[0] = "host.test"

	tests := []struct {
		name    string
		ok      bool
		mani    Manifest
		dgroups []dtypes.GroupSpec
	}{
		{
			name: "empty",
			ok:   false,
		},

		{
			name: "single",
			ok:   true,
			mani: []Group{
				{
					Name: "foo",
					Services: []Service{
						{
							Name:      "svc1",
							Image:     "test",
							Resources: simpleResources(expose),
							Count:     3,
							Expose:    expose,
						},
					},
				},
			},
			dgroups: []dtypes.GroupSpec{
				{
					Name: "foo",
					Resources: dtypes.ResourceUnits{
						{
							Resources: simpleResources(expose),
							Count:     3,
						},
					},
				},
			},
		},

		{
			name: "multi-mgroup",
			ok:   true,
			mani: []Group{
				{
					Name: "foo",
					Services: []Service{
						{
							Name:      "svc1",
							Image:     "test",
							Resources: simpleResources(expose),
							Count:     1,
							Expose:    expose,
						},
						{
							Name:      "svc1",
							Image:     "test",
							Resources: simpleResources(expose),
							Count:     2,
						},
					},
				},
			},
			dgroups: []dtypes.GroupSpec{
				{
					Name: "foo",
					Resources: dtypes.ResourceUnits{
						{
							Resources: simpleResources(expose),
							Count:     3,
						},
					},
				},
			},
		},

		{
			name: "mismatch-name",
			ok:   false,
			mani: []Group{
				{
					Name: "foo-bad",
					Services: []Service{
						{
							Name:      "svc1",
							Image:     "test",
							Resources: randUnits1,
							Count:     3,
						},
					},
				},
			},
			dgroups: []dtypes.GroupSpec{
				{
					Name: "foo",
					Resources: dtypes.ResourceUnits{
						{
							Resources: randUnits1,
							Count:     3,
						},
					},
				},
			},
		},

		{
			name: "mismatch-cpu",
			ok:   false,
			mani: []Group{
				{
					Name: "foo",
					Services: []Service{
						{
							Name:      "svc1",
							Image:     "test",
							Resources: randUnits3,
							Count:     3,
						},
					},
				},
			},
			dgroups: []dtypes.GroupSpec{
				{
					Name: "foo",
					Resources: dtypes.ResourceUnits{
						{
							Resources: randUnits1,
							Count:     3,
						},
					},
				},
			},
		},

		{
			name: "mismatch-group-count",
			ok:   false,
			mani: []Group{
				{
					Name: "foo",
					Services: []Service{
						{
							Name:      "svc1",
							Image:     "test",
							Resources: randUnits3,
							Count:     3,
						},
					},
				},
			},
			dgroups: []dtypes.GroupSpec{},
		},
	}

	for _, test := range tests {
		err := test.mani.CheckAgainstGSpecs(test.dgroups)
		if test.ok {
			assert.NoError(t, err, test.name)
		} else {
			assert.Error(t, err, test.name)
		}
	}
}
