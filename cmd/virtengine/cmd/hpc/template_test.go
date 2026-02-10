package hpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestFilterTemplates(t *testing.T) {
	templates := []*hpctypes.WorkloadTemplate{
		{
			TemplateID:     "gpu-1",
			Name:           "GPU",
			Type:           hpctypes.WorkloadTypeGPU,
			ApprovalStatus: hpctypes.WorkloadApprovalApproved,
			Tags:           []string{"gpu", "fast"},
		},
		{
			TemplateID:     "mpi-1",
			Name:           "MPI",
			Type:           hpctypes.WorkloadTypeMPI,
			ApprovalStatus: hpctypes.WorkloadApprovalPending,
			Tags:           []string{"mpi"},
		},
	}

	filtered, err := filterTemplates(templates, templateFilter{workloadType: "gpu"})
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "gpu-1", filtered[0].TemplateID)

	filtered, err = filterTemplates(templates, templateFilter{status: "pending"})
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "mpi-1", filtered[0].TemplateID)

	filtered, err = filterTemplates(templates, templateFilter{tags: []string{"fast"}})
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "gpu-1", filtered[0].TemplateID)

	_, err = filterTemplates(templates, templateFilter{workloadType: "invalid"})
	require.Error(t, err)
}

func TestRenderTemplateTable(t *testing.T) {
	templates := []*hpctypes.WorkloadTemplate{
		{TemplateID: "t1", Name: "Template One", Version: "1.0.0", Type: hpctypes.WorkloadTypeBatch, ApprovalStatus: hpctypes.WorkloadApprovalApproved},
	}

	var buf bytes.Buffer
	require.NoError(t, renderTemplateTable(&buf, templates))

	output := buf.String()
	require.Contains(t, output, "ID")
	require.Contains(t, output, "NAME")
	require.Contains(t, output, "VERSION")
	require.Contains(t, output, "t1")
}

func TestRenderTemplateValidation(t *testing.T) {
	template := &hpctypes.WorkloadTemplate{TemplateID: "test", Version: "1.0.0"}
	result := &hpc_workload_library.ValidationResult{
		Valid: false,
		Errors: []hpc_workload_library.ValidationError{
			{Field: "runtime", Message: "invalid", Code: "TEST"},
		},
	}

	var buf bytes.Buffer
	renderTemplateValidation(&buf, template, result)

	output := buf.String()
	require.Contains(t, output, "Template: test")
	require.Contains(t, output, "Errors:")
	require.Contains(t, output, "runtime")
}

func TestNormalizeTemplateOutput(t *testing.T) {
	format, err := normalizeTemplateOutput("json")
	require.NoError(t, err)
	require.Equal(t, "json", format)

	_, err = normalizeTemplateOutput("invalid")
	require.Error(t, err)
}
