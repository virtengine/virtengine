package v2beta3

import (
	"fmt"
	"sort"
	"strings"

	k8svalidation "k8s.io/apimachinery/pkg/util/validation"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
)

func (s *Service) validate(helper *validateManifestGroupsHelper) error {
	if len(s.Name) == 0 {
		return fmt.Errorf("%w: service name is empty", ErrInvalidManifest)
	}

	serviceNameValid := serviceNameValidationRegex.MatchString(s.Name)
	if !serviceNameValid {
		return fmt.Errorf("%w: service %q name is invalid", ErrInvalidManifest, s.Name)
	}

	if len(s.Image) == 0 {
		return fmt.Errorf("%w: service %q has empty image name", ErrInvalidManifest, s.Name)
	}

	if err := s.Resources.Validate(); err != nil {
		return err
	}

	for _, envVar := range s.Env {
		tokens := strings.SplitN(envVar, "=", 2)
		if len(tokens) == 0 {
			return fmt.Errorf("%w: service %q defines an env. var. with an empty name", ErrInvalidManifest, s.Name)
		}

		envVarName := tokens[0]

		if len(k8svalidation.IsEnvVarName(envVarName)) != 0 {
			return fmt.Errorf("%w: service %q defines an env. var. with an invalid name %q", ErrInvalidManifest, s.Name, envVarName)
		}
	}

	if !sort.IsSorted(s.Expose) {
		return fmt.Errorf("%w: service %q: expose is not sorted", ErrInvalidManifest, s.Name)
	}

	for _, serviceExpose := range s.Expose {
		if err := serviceExpose.validate(helper); err != nil {
			return fmt.Errorf("%w: service %q: %w", ErrInvalidManifest, s.Name, err)
		}
	}

	return nil
}

func (s *Service) checkAgainstGSpec(gspec *groupSpec) error {
	// find resource units by id
	var gRes *dtypes.ResourceUnit

	for idx := range gspec.gs.Resources {
		if s.Resources.ID == gspec.gs.Resources[idx].ID {
			gRes = &gspec.gs.Resources[idx]
			break
		}
	}

	if gRes == nil {
		return fmt.Errorf("service %q: not found deployment group resources with ID = %d", s.Name, s.Resources.ID)
	}

	if s.Count > gRes.Count {
		return fmt.Errorf("service %q: over-utilized replicas (%d) > group spec resources count (%d)",
			s.Name, s.Count, gRes.Count)
	}

	// do not compare resources directly
	if !s.Resources.CPU.Equal(gRes.CPU) {
		return fmt.Errorf("service %q: CPU resources mismatch for ID %d", s.Name, s.Resources.ID)
	}

	if !s.Resources.GPU.Equal(gRes.GPU) {
		return fmt.Errorf("service %q: GPU resources mismatch for ID %d", s.Name, s.Resources.ID)
	}

	if !s.Resources.Memory.Equal(gRes.Memory) {
		return fmt.Errorf("service %q: Memory resources mismatch for ID %d", s.Name, s.Resources.ID)
	}

	if !s.Resources.Storage.Equal(gRes.Storage) {
		return fmt.Errorf("service %q: Storage resources mismatch for ID %d", s.Name, s.Resources.ID)
	}

	for _, expose := range s.Expose {
		if err := expose.checkAgainstResources(gRes, gspec.endpoints); err != nil {
			return fmt.Errorf("service %q: resource ID %d: %w", s.Name, gRes.ID, err)
		}
	}

	gRes.Count -= s.Count

	return nil
}
