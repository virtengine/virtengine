/**
 * Job Wizard Store
 *
 * Manages state for the job submission wizard using Zustand.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { JobManifest, WorkloadTemplate } from '../types';

export type WizardStep = 'template' | 'configure' | 'resources' | 'review';

interface WizardState {
  // Current wizard step
  currentStep: WizardStep;

  // Selected template (if using template)
  selectedTemplate: WorkloadTemplate | null;

  // Job manifest being built
  manifest: Partial<JobManifest>;

  // Offering ID selected
  offeringId: string | null;

  // Estimated cost
  estimatedCost: {
    total: string;
    perHour: string;
    breakdown: {
      compute: string;
      storage: string;
      network: string;
      gpu?: string;
    };
    denom: string;
  } | null;

  // Validation errors
  errors: Record<string, string>;

  // Actions
  setStep: (step: WizardStep) => void;
  nextStep: () => void;
  prevStep: () => void;
  selectTemplate: (template: WorkloadTemplate | null) => void;
  updateManifest: (updates: Partial<JobManifest>) => void;
  setOffering: (offeringId: string) => void;
  setEstimatedCost: (cost: NonNullable<WizardState['estimatedCost']>) => void;
  setError: (field: string, error: string) => void;
  clearError: (field: string) => void;
  clearErrors: () => void;
  reset: () => void;
}

const stepOrder: WizardStep[] = ['template', 'configure', 'resources', 'review'];

const initialState = {
  currentStep: 'template' as WizardStep,
  selectedTemplate: null,
  manifest: {},
  offeringId: null,
  estimatedCost: null,
  errors: {},
};

export const useWizardStore = create<WizardState>()(
  persist(
    (set, get) => ({
      ...initialState,

      setStep: (step) => set({ currentStep: step }),

      nextStep: () => {
        const { currentStep } = get();
        const currentIndex = stepOrder.indexOf(currentStep);
        if (currentIndex < stepOrder.length - 1) {
          set({ currentStep: stepOrder[currentIndex + 1] });
        }
      },

      prevStep: () => {
        const { currentStep } = get();
        const currentIndex = stepOrder.indexOf(currentStep);
        if (currentIndex > 0) {
          set({ currentStep: stepOrder[currentIndex - 1] });
        }
      },

      selectTemplate: (template) =>
        set((state) => ({
          selectedTemplate: template,
          manifest: template
            ? {
                ...state.manifest,
                templateId: template.id,
                resources: template.defaultResources,
              }
            : state.manifest,
        })),

      updateManifest: (updates) =>
        set((state) => ({
          manifest: { ...state.manifest, ...updates },
        })),

      setOffering: (offeringId) => set({ offeringId }),

      setEstimatedCost: (cost) => set({ estimatedCost: cost }),

      setError: (field, error) =>
        set((state) => ({
          errors: { ...state.errors, [field]: error },
        })),

      clearError: (field) =>
        set((state) => {
          const newErrors = { ...state.errors };
          delete newErrors[field];
          return { errors: newErrors };
        }),

      clearErrors: () => set({ errors: {} }),

      reset: () => set(initialState),
    }),
    {
      name: 'hpc-wizard-storage', // localStorage key
      partialize: (state) => ({
        // Only persist manifest and selected template (not current step)
        manifest: state.manifest,
        selectedTemplate: state.selectedTemplate,
      }),
    }
  )
);
