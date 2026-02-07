import { describe, it, expect, beforeEach } from 'vitest';
import { useWizardStore } from '@/features/hpc/stores/wizardStore';
import { act } from '@testing-library/react';
import type { WorkloadTemplate } from '@/features/hpc';

const mockTemplate: WorkloadTemplate = {
  id: 'pytorch-training',
  name: 'PyTorch Training',
  description: 'Train deep learning models.',
  category: 'ml_training',
  defaultResources: {
    nodes: 1,
    cpusPerNode: 8,
    memoryGBPerNode: 64,
    gpusPerNode: 2,
    gpuType: 'nvidia-a100',
    maxRuntimeSeconds: 86400,
    storageGB: 100,
  },
  defaultParameters: {},
  requiredIdentityScore: 0,
  mfaRequired: false,
  estimatedCostPerHour: '5.50',
  version: '1.0.0',
};

describe('WizardStore', () => {
  beforeEach(() => {
    act(() => {
      useWizardStore.getState().reset();
    });
  });

  it('starts at template step', () => {
    const state = useWizardStore.getState();
    expect(state.currentStep).toBe('template');
  });

  it('advances through steps', () => {
    act(() => {
      useWizardStore.getState().nextStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('configure');

    act(() => {
      useWizardStore.getState().nextStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('resources');

    act(() => {
      useWizardStore.getState().nextStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('review');
  });

  it('does not advance past review', () => {
    act(() => {
      useWizardStore.getState().setStep('review');
      useWizardStore.getState().nextStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('review');
  });

  it('goes back through steps', () => {
    act(() => {
      useWizardStore.getState().setStep('resources');
      useWizardStore.getState().prevStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('configure');
  });

  it('does not go before template', () => {
    act(() => {
      useWizardStore.getState().prevStep();
    });
    expect(useWizardStore.getState().currentStep).toBe('template');
  });

  it('selects a template and updates manifest', () => {
    act(() => {
      useWizardStore.getState().selectTemplate(mockTemplate);
    });

    const state = useWizardStore.getState();
    expect(state.selectedTemplate).toEqual(mockTemplate);
    expect(state.manifest.templateId).toBe('pytorch-training');
    expect(state.manifest.resources).toEqual(mockTemplate.defaultResources);
  });

  it('clears template selection', () => {
    act(() => {
      useWizardStore.getState().selectTemplate(mockTemplate);
      useWizardStore.getState().selectTemplate(null);
    });

    const state = useWizardStore.getState();
    expect(state.selectedTemplate).toBeNull();
  });

  it('updates manifest fields', () => {
    act(() => {
      useWizardStore.getState().updateManifest({ name: 'My Job' });
      useWizardStore.getState().updateManifest({ description: 'Test' });
    });

    const state = useWizardStore.getState();
    expect(state.manifest.name).toBe('My Job');
    expect(state.manifest.description).toBe('Test');
  });

  it('sets offering', () => {
    act(() => {
      useWizardStore.getState().setOffering('offering-1');
    });
    expect(useWizardStore.getState().offeringId).toBe('offering-1');
  });

  it('sets estimated cost', () => {
    const cost = {
      total: '100.00',
      perHour: '5.00',
      breakdown: { compute: '3.00', storage: '1.00', network: '0.50', gpu: '0.50' },
      denom: 'uakt',
    };

    act(() => {
      useWizardStore.getState().setEstimatedCost(cost);
    });
    expect(useWizardStore.getState().estimatedCost).toEqual(cost);
  });

  it('manages validation errors', () => {
    act(() => {
      useWizardStore.getState().setError('name', 'Name is required');
    });
    expect(useWizardStore.getState().errors.name).toBe('Name is required');

    act(() => {
      useWizardStore.getState().clearError('name');
    });
    expect(useWizardStore.getState().errors.name).toBeUndefined();
  });

  it('clears all errors', () => {
    act(() => {
      useWizardStore.getState().setError('name', 'Required');
      useWizardStore.getState().setError('resources', 'Invalid');
      useWizardStore.getState().clearErrors();
    });
    expect(Object.keys(useWizardStore.getState().errors)).toHaveLength(0);
  });

  it('resets to initial state', () => {
    act(() => {
      useWizardStore.getState().selectTemplate(mockTemplate);
      useWizardStore.getState().setStep('review');
      useWizardStore.getState().updateManifest({ name: 'Test' });
      useWizardStore.getState().reset();
    });

    const state = useWizardStore.getState();
    expect(state.currentStep).toBe('template');
    expect(state.selectedTemplate).toBeNull();
    expect(state.manifest).toEqual({});
    expect(state.estimatedCost).toBeNull();
  });
});
