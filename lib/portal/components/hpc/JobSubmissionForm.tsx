// @ts-nocheck
/**
 * Job Submission Form Component
 * VE-705: HPC job submission UI
 */
import * as React from 'react';
import { useHPC } from '../../hooks/useHPC';
import { formatTokenAmount, formatDuration } from '../../utils/format';
import { sanitizePlainText, sanitizeJsonInput } from '../../utils/security';
import type { WorkloadTemplate, JobManifest, JobPriceQuote, JobResources } from '../../types/hpc';

/**
 * Job submission form props
 */
export interface JobSubmissionFormProps {
  /**
   * Pre-selected workload template
   */
  template?: WorkloadTemplate;

  /**
   * Callback when job is submitted
   */
  onSubmit?: (jobId: string) => void;

  /**
   * Callback when submission is cancelled
   */
  onCancel?: () => void;

  /**
   * Custom CSS class
   */
  className?: string;
}

/**
 * Form step
 */
type FormStep = 'template' | 'config' | 'review' | 'submitting';

/**
 * Job submission form component
 */
export function JobSubmissionForm({
  template: initialTemplate,
  onSubmit,
  onCancel,
  className = '',
}: JobSubmissionFormProps): JSX.Element {
  const { state, actions } = useHPC();

  const [step, setStep] = React.useState<FormStep>(initialTemplate ? 'config' : 'template');
  const [selectedTemplate, setSelectedTemplate] = React.useState<WorkloadTemplate | null>(
    initialTemplate || null
  );
  const [quote, setQuote] = React.useState<JobPriceQuote | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  // Form state
  const [name, setName] = React.useState('');
  const [nodes, setNodes] = React.useState(1);
  const [cpusPerNode, setCpusPerNode] = React.useState(8);
  const [memoryGBPerNode, setMemoryGBPerNode] = React.useState(32);
  const [gpusPerNode, setGpusPerNode] = React.useState(0);
  const [maxRuntimeSeconds, setMaxRuntimeSeconds] = React.useState(3600); // 1 hour
  const [storageGB, setStorageGB] = React.useState(10);
  const [encryptedInputs, setEncryptedInputs] = React.useState('');

  // Load templates on mount
  React.useEffect(() => {
    if (!initialTemplate) {
      actions.getWorkloadTemplates();
    }
  }, [initialTemplate, actions]);

  // Get templates from state
  const templates = state.workloadTemplates;

  /**
   * Handle template selection
   */
  const handleSelectTemplate = (template: WorkloadTemplate) => {
    setSelectedTemplate(template);
    setName(`${template.name} Job`);
    setNodes(template.defaultResources.nodes);
    setCpusPerNode(template.defaultResources.cpusPerNode);
    setMemoryGBPerNode(template.defaultResources.memoryGBPerNode);
    setGpusPerNode(template.defaultResources.gpusPerNode || 0);
    setMaxRuntimeSeconds(template.defaultResources.maxRuntimeSeconds);
    setStorageGB(template.defaultResources.storageGB);
    setStep('config');
  };

  /**
   * Handle configuration complete
   */
  const handleConfigComplete = async () => {
    if (!selectedTemplate) return;

    setError(null);

    try {
      const sanitizedName = sanitizePlainText(name, { maxLength: 120 });
      if (!sanitizedName) {
        setError('Job name is required');
        return;
      }

      let sanitizedEncryptedInputs: Record<string, unknown> | undefined = undefined;
      if (encryptedInputs && encryptedInputs.trim()) {
        try {
          sanitizedEncryptedInputs = sanitizeJsonInput(encryptedInputs, {
            maxDepth: 4,
            maxKeyLength: 64,
            maxStringLength: 4096,
            escapeHtmlStrings: false,
          });
        } catch {
          setError('Encrypted inputs must be valid JSON');
          return;
        }
      }

      const resources: JobResources = {
        nodes,
        cpusPerNode,
        memoryGBPerNode,
        gpusPerNode: gpusPerNode > 0 ? gpusPerNode : undefined,
        maxRuntimeSeconds,
        storageGB,
      };

      const manifest: Partial<JobManifest> = {
        version: '1.0',
        templateId: selectedTemplate.id,
        name: sanitizedName,
        resources,
        parameters: {},
        encryptedInputs: sanitizedEncryptedInputs,
      };

      // Update manifest in state
      actions.updateJobManifest(manifest);

      // Get price quote
      const priceQuote = await actions.getQuote();
      setQuote(priceQuote);
      setStep('review');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get quote');
    }
  };

  /**
   * Handle job submission
   */
  const handleSubmit = async () => {
    setStep('submitting');
    setError(null);

    try {
      const result = await actions.submitJob();
      onSubmit?.(result.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit job');
      setStep('review');
    }
  };

  return (
    <div className={`job-form ${className}`}>
      {/* Progress */}
      <div className="job-form__progress">
        <div className={`job-form__progress-step ${step === 'template' ? 'job-form__progress-step--active' : step !== 'template' ? 'job-form__progress-step--done' : ''}`}>
          1. Template
        </div>
        <div className={`job-form__progress-step ${step === 'config' ? 'job-form__progress-step--active' : ['review', 'submitting'].includes(step) ? 'job-form__progress-step--done' : ''}`}>
          2. Configure
        </div>
        <div className={`job-form__progress-step ${step === 'review' || step === 'submitting' ? 'job-form__progress-step--active' : ''}`}>
          3. Review
        </div>
      </div>

      {/* Content */}
      <div className="job-form__content">
        {step === 'template' && (
          <TemplateSelector
            templates={templates}
            onSelect={handleSelectTemplate}
            isLoading={state.isLoading}
          />
        )}

        {step === 'config' && selectedTemplate && (
          <ConfigurationForm
            template={selectedTemplate}
            name={name}
            onNameChange={setName}
            nodes={nodes}
            onNodesChange={setNodes}
            cpusPerNode={cpusPerNode}
            onCpusPerNodeChange={setCpusPerNode}
            memoryGBPerNode={memoryGBPerNode}
            onMemoryGBPerNodeChange={setMemoryGBPerNode}
            gpusPerNode={gpusPerNode}
            onGpusPerNodeChange={setGpusPerNode}
            maxRuntimeSeconds={maxRuntimeSeconds}
            onMaxRuntimeSecondsChange={setMaxRuntimeSeconds}
            storageGB={storageGB}
            onStorageGBChange={setStorageGB}
            encryptedInputs={encryptedInputs}
            onEncryptedInputsChange={setEncryptedInputs}
            onContinue={handleConfigComplete}
            onBack={() => setStep('template')}
            isLoading={state.isLoading}
            error={error}
          />
        )}

        {step === 'review' && quote && state.submission?.manifest && (
          <ReviewStep
            manifest={state.submission.manifest}
            quote={quote}
            onSubmit={handleSubmit}
            onBack={() => setStep('config')}
            isLoading={false}
            error={error}
          />
        )}

        {step === 'submitting' && (
          <div className="job-form__submitting">
            <div className="job-form__spinner" />
            <p>Submitting job to the network...</p>
          </div>
        )}
      </div>

      {/* Cancel button */}
      {step !== 'submitting' && (
        <div className="job-form__actions">
          <button
            className="job-form__button job-form__button--secondary"
            onClick={onCancel}
          >
            Cancel
          </button>
        </div>
      )}

      <style>{formStyles}</style>
    </div>
  );
}

/**
 * Template selector step
 */
interface TemplateSelectorProps {
  templates: WorkloadTemplate[];
  onSelect: (template: WorkloadTemplate) => void;
  isLoading: boolean;
}

function TemplateSelector({ templates, onSelect, isLoading }: TemplateSelectorProps): JSX.Element {
  if (isLoading) {
    return (
      <div className="job-form__loading">Loading templates...</div>
    );
  }

  const getCategoryIcon = (category: string): string => {
    switch (category) {
      case 'ml_training':
      case 'ml_inference':
        return '🧠';
      case 'data_processing':
        return '📊';
      case 'rendering':
        return '🎨';
      case 'simulation':
        return '🔬';
      case 'scientific':
        return '🔭';
      default:
        return '🖥️';
    }
  };

  return (
    <div className="job-form__templates">
      <h3 className="job-form__title">Select a Workload Template</h3>
      <div className="job-form__template-grid">
        {templates.map((template) => (
          <button
            key={template.id}
            className="job-form__template"
            onClick={() => onSelect(template)}
          >
            <span className="job-form__template-icon">{getCategoryIcon(template.category)}</span>
            <span className="job-form__template-name">{template.name}</span>
            <span className="job-form__template-desc">{template.description}</span>
          </button>
        ))}
      </div>
    </div>
  );
}

/**
 * Configuration form step
 */
interface ConfigurationFormProps {
  template: WorkloadTemplate;
  name: string;
  onNameChange: (name: string) => void;
  nodes: number;
  onNodesChange: (nodes: number) => void;
  cpusPerNode: number;
  onCpusPerNodeChange: (cpus: number) => void;
  memoryGBPerNode: number;
  onMemoryGBPerNodeChange: (memory: number) => void;
  gpusPerNode: number;
  onGpusPerNodeChange: (gpus: number) => void;
  maxRuntimeSeconds: number;
  onMaxRuntimeSecondsChange: (seconds: number) => void;
  storageGB: number;
  onStorageGBChange: (storage: number) => void;
  encryptedInputs: string;
  onEncryptedInputsChange: (inputs: string) => void;
  onContinue: () => void;
  onBack: () => void;
  isLoading: boolean;
  error: string | null;
}

function ConfigurationForm({
  template,
  name,
  onNameChange,
  nodes,
  onNodesChange,
  cpusPerNode,
  onCpusPerNodeChange,
  memoryGBPerNode,
  onMemoryGBPerNodeChange,
  gpusPerNode,
  onGpusPerNodeChange,
  maxRuntimeSeconds,
  onMaxRuntimeSecondsChange,
  storageGB,
  onStorageGBChange,
  encryptedInputs,
  onEncryptedInputsChange,
  onContinue,
  onBack,
  isLoading,
  error,
}: ConfigurationFormProps): JSX.Element {
  return (
    <div className="job-form__config" role="form" aria-label={`Configure ${template.name} job`}>
      <h3 className="job-form__title" id="config-title">Configure {template.name}</h3>

      <div className="job-form__field">
        <label htmlFor="job-name" className="job-form__label">
          Job Name <span aria-hidden="true">*</span>
        </label>
        <input
          id="job-name"
          type="text"
          className="job-form__input"
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
          placeholder="Enter job name"
          aria-required="true"
          aria-describedby={error && !name ? 'job-name-error' : undefined}
        />
      </div>

      <div className="job-form__field-row">
        <div className="job-form__field">
          <label htmlFor="nodes" className="job-form__label">Nodes</label>
          <input
            id="nodes"
            type="number"
            className="job-form__input"
            value={nodes}
            onChange={(e) => onNodesChange(parseInt(e.target.value, 10))}
            min={1}
          />
        </div>
        <div className="job-form__field">
          <label htmlFor="cpus-per-node" className="job-form__label">CPUs per Node</label>
          <input
            id="cpus-per-node"
            type="number"
            className="job-form__input"
            value={cpusPerNode}
            onChange={(e) => onCpusPerNodeChange(parseInt(e.target.value, 10))}
            min={template.defaultResources.cpusPerNode}
            aria-describedby="cpus-hint"
          />
          <span id="cpus-hint" className="job-form__hint">
            Minimum: {template.defaultResources.cpusPerNode}
          </span>
        </div>
      </div>

      <div className="job-form__field-row">
        <div className="job-form__field">
          <label htmlFor="memory-per-node" className="job-form__label">Memory per Node (GB)</label>
          <input
            id="memory-per-node"
            type="number"
            className="job-form__input"
            value={memoryGBPerNode}
            onChange={(e) => onMemoryGBPerNodeChange(parseInt(e.target.value, 10))}
            min={template.defaultResources.memoryGBPerNode}
            aria-describedby="memory-hint"
          />
          <span id="memory-hint" className="job-form__hint">
            Minimum: {template.defaultResources.memoryGBPerNode} GB
          </span>
        </div>
        <div className="job-form__field">
          <label htmlFor="gpus-per-node" className="job-form__label">GPUs per Node</label>
          <input
            id="gpus-per-node"
            type="number"
            className="job-form__input"
            value={gpusPerNode}
            onChange={(e) => onGpusPerNodeChange(parseInt(e.target.value, 10))}
            min={0}
          />
        </div>
      </div>

      <div className="job-form__field-row">
        <div className="job-form__field">
          <label htmlFor="storage" className="job-form__label">Storage (GB)</label>
          <input
            id="storage"
            type="number"
            className="job-form__input"
            value={storageGB}
            onChange={(e) => onStorageGBChange(parseInt(e.target.value, 10))}
            min={1}
          />
        </div>
        <div className="job-form__field">
          <label htmlFor="max-runtime" className="job-form__label">Max Runtime</label>
          <select
            id="max-runtime"
            className="job-form__select"
            value={maxRuntimeSeconds}
            onChange={(e) => onMaxRuntimeSecondsChange(parseInt(e.target.value, 10))}
          >
            <option value={1800}>30 minutes</option>
            <option value={3600}>1 hour</option>
            <option value={7200}>2 hours</option>
            <option value={14400}>4 hours</option>
            <option value={28800}>8 hours</option>
            <option value={86400}>24 hours</option>
          </select>
        </div>
      </div>

      <div className="job-form__field">
        <label htmlFor="encrypted-inputs" className="job-form__label">
          Encrypted Inputs (JSON, optional)
        </label>
        <textarea
          id="encrypted-inputs"
          className="job-form__textarea"
          value={encryptedInputs}
          onChange={(e) => onEncryptedInputsChange(e.target.value)}
          placeholder='{"key": "encrypted_value"}'
          rows={4}
          aria-describedby="inputs-hint"
        />
        <span id="inputs-hint" className="job-form__hint">
          Optional JSON object with encrypted input parameters
        </span>
      </div>

      {error && (
        <p id="job-name-error" className="job-form__error" role="alert" aria-live="assertive">
          <span aria-hidden="true">⚠ </span>
          {error}
        </p>
      )}

      <div className="job-form__buttons">
        <button
          className="job-form__button job-form__button--secondary"
          onClick={onBack}
          type="button"
        >
          Back
        </button>
        <button
          className="job-form__button job-form__button--primary"
          onClick={onContinue}
          disabled={!name || isLoading}
          aria-busy={isLoading}
          type="button"
        >
          {isLoading ? 'Getting Quote...' : 'Continue'}
          {isLoading && <span className="sr-only">Please wait</span>}
        </button>
      </div>
    </div>
  );
}

/**
 * Review step
 */
interface ReviewStepProps {
  manifest: Partial<JobManifest>;
  quote: JobPriceQuote;
  onSubmit: () => void;
  onBack: () => void;
  isLoading: boolean;
  error: string | null;
}

function ReviewStep({
  manifest,
  quote,
  onSubmit,
  onBack,
  isLoading,
  error,
}: ReviewStepProps): JSX.Element {
  const resources = manifest.resources;

  return (
    <div className="job-form__review">
      <h3 className="job-form__title">Review Job</h3>

      <div className="job-form__summary">
        <div className="job-form__summary-row">
          <span className="job-form__summary-label">Job Name</span>
          <span className="job-form__summary-value">{manifest.name}</span>
        </div>
        <div className="job-form__summary-row">
          <span className="job-form__summary-label">Template</span>
          <span className="job-form__summary-value">{manifest.templateId}</span>
        </div>
        {resources && (
          <>
            <div className="job-form__summary-row">
              <span className="job-form__summary-label">Nodes</span>
              <span className="job-form__summary-value">{resources.nodes}</span>
            </div>
            <div className="job-form__summary-row">
              <span className="job-form__summary-label">Resources per Node</span>
              <span className="job-form__summary-value">
                {resources.cpusPerNode} CPUs, {resources.memoryGBPerNode}GB RAM
                {resources.gpusPerNode ? `, ${resources.gpusPerNode} GPUs` : ''}
              </span>
            </div>
            <div className="job-form__summary-row">
              <span className="job-form__summary-label">Max Runtime</span>
              <span className="job-form__summary-value">
                {formatDuration(resources.maxRuntimeSeconds)}
              </span>
            </div>
            <div className="job-form__summary-row">
              <span className="job-form__summary-label">Storage</span>
              <span className="job-form__summary-value">{resources.storageGB} GB</span>
            </div>
          </>
        )}
      </div>

      <div className="job-form__quote">
        <h4 className="job-form__quote-title">Price Quote</h4>
        <div className="job-form__quote-row">
          <span>Estimated Cost</span>
          <span className="job-form__quote-amount">
            {formatTokenAmount(quote.estimatedCost)}
          </span>
        </div>
        <div className="job-form__quote-row">
          <span>Required Deposit</span>
          <span className="job-form__quote-amount">
            {formatTokenAmount(quote.requiredDeposit)}
          </span>
        </div>
        <p className="job-form__quote-note">
          Valid until: {new Date(quote.validUntil * 1000).toLocaleString()}
        </p>
      </div>

      {error && <p className="job-form__error">{error}</p>}

      <div className="job-form__buttons">
        <button
          className="job-form__button job-form__button--secondary"
          onClick={onBack}
        >
          Back
        </button>
        <button
          className="job-form__button job-form__button--primary"
          onClick={onSubmit}
          disabled={isLoading}
        >
          Submit Job
        </button>
      </div>
    </div>
  );
}

const formStyles = `
  .job-form {
    background: white;
    border-radius: 12px;
    padding: 24px;
    max-width: 600px;
    margin: 0 auto;
  }

  .job-form__progress {
    display: flex;
    justify-content: space-between;
    margin-bottom: 32px;
  }

  .job-form__progress-step {
    font-size: 0.75rem;
    color: #9ca3af;
    padding: 8px 16px;
    border-radius: 4px;
    background: #f3f4f6;
  }

  .job-form__progress-step--active {
    background: #dbeafe;
    color: #2563eb;
    font-weight: 600;
  }

  .job-form__progress-step--done {
    background: #dcfce7;
    color: #166534;
  }

  .job-form__title {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 24px;
  }

  .job-form__template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 12px;
  }

  .job-form__template {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    background: white;
    cursor: pointer;
    text-align: center;
    transition: all 0.2s;
  }

  .job-form__template:hover {
    border-color: #3b82f6;
    background: #f8fafc;
  }

  .job-form__template-icon {
    font-size: 2rem;
  }

  .job-form__template-name {
    font-weight: 600;
    color: #111827;
  }

  .job-form__template-desc {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .job-form__template-tee {
    font-size: 0.625rem;
    color: #16a34a;
    background: #dcfce7;
    padding: 2px 8px;
    border-radius: 9999px;
  }

  .job-form__field {
    margin-bottom: 16px;
  }

  .job-form__field-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .job-form__label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    margin-bottom: 4px;
  }

  .job-form__hint {
    display: block;
    font-size: 0.75rem;
    color: #6b7280;
    margin-top: 4px;
  }

  .job-form__input,
  .job-form__select,
  .job-form__textarea {
    width: 100%;
    padding: 10px 12px;
    border: 2px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.875rem;
    min-height: 44px; /* WCAG 2.5.5 target size */
  }

  .job-form__input:focus,
  .job-form__select:focus,
  .job-form__textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  .job-form__checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    color: #374151;
    cursor: pointer;
  }

  .job-form__checkbox-label input[type="checkbox"] {
    width: 20px;
    height: 20px;
    accent-color: #3b82f6;
  }

  .job-form__checkbox-label input[type="checkbox"]:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  /* Screen reader only utility */
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .job-form__summary {
    background: #f9fafb;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .job-form__summary-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid #e5e7eb;
  }

  .job-form__summary-row:last-child {
    border-bottom: none;
  }

  .job-form__summary-label {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .job-form__summary-value {
    color: #111827;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .job-form__quote {
    background: #f0f9ff;
    border: 1px solid #bae6fd;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .job-form__quote-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: #0369a1;
    margin: 0 0 12px;
  }

  .job-form__quote-row {
    display: flex;
    justify-content: space-between;
    padding: 4px 0;
    font-size: 0.875rem;
  }

  .job-form__quote-amount {
    font-weight: 600;
    color: #111827;
  }

  .job-form__quote-note {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 8px 0 0;
  }

  .job-form__error {
    color: #991b1b;
    font-size: 0.875rem;
    margin: 16px 0;
    padding: 12px;
    background: #fee2e2;
    border-radius: 6px;
    border-left: 4px solid #dc2626;
  }

  .job-form__buttons {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
  }

  .job-form__button {
    padding: 10px 20px;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
    min-height: 44px; /* WCAG 2.5.5 target size */
  }

  .job-form__button:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  .job-form__button--primary {
    background: #3b82f6;
    color: white;
  }

  .job-form__button--primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .job-form__button--primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .job-form__button--secondary {
    background: white;
    border: 2px solid #d1d5db;
    color: #374151;
  }

  .job-form__button--secondary:hover {
    background: #f9fafb;
  }

  .job-form__submitting {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 48px 0;
  }

  .job-form__spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .job-form__loading {
    text-align: center;
    padding: 48px 0;
    color: #6b7280;
  }

  .job-form__actions {
    display: flex;
    justify-content: center;
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid #e5e7eb;
  }
`;
