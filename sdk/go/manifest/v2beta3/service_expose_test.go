package v2beta3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldBeIngress(t *testing.T) {
	// Should not create ingress for something on port 81
	exp := ServiceExpose{
		Global: true,
		Proto:  TCP,
		Port:   81,
	}

	require.False(t, exp.IsIngress())

	exp = ServiceExpose{
		Global: true,
		Proto:  TCP,
		Port:   80,
	}

	// Should create ingress for something on port 80
	require.True(t, exp.IsIngress())

	exp = ServiceExpose{
		Global: false,
		Proto:  TCP,
		Port:   80,
	}

	// Should not create ingress for something on port 80 that is not Global
	require.False(t, exp.IsIngress())

	exp = ServiceExpose{
		Global: true,
		Proto:  UDP,
		Port:   80,
	}

	// Should not create ingress for something on port 80 that is UDP
	require.False(t, exp.IsIngress())
}
