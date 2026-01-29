/**
 * Job Submission Form Component
 * VE-705: HPC job submission UI
 */
import * as React from 'react';
import { useHPC } from '../../hooks/useHPC';
import { formatTokenAmount, formatDuration } from '../../utils/format';
import { sanitizePlainText, sanitizeJsonInput } from '../../utils/security';
import type { WorkloadTemplate, JobManifest, JobPriceQuote } from '../../types/hpc';

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
  const {
    state,
    getWorkloadTemplates,
    startJobSubmission,
    getQuote,
    submitJob,
  } = useHPC();

  const [step, setStep] = React.useState<FormStep>(initialTemplate ? 'config' : 'template');
  const [selectedTemplate, setSelectedTemplate] = React.useState<WorkloadTemplate | null>(
    initialTemplate || null
  );
  const [templates, setTemplates] = React.useState<WorkloadTemplate[]>([]);
  const [quote, setQuote] = React.useState<JobPriceQuote | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  // Form state
  const [name, setName] = React.useState('');
  const [cpu, setCpu] = React.useState(4);
  const [memory, setMemory] = React.useState(8);
  const [gpu, setGpu] = React.useState(0);
  const [duration, setDuration] = React.useState(3600); // 1 hour
  const [requiresTee, setRequiresTee] = React.useState(true);
  const [encryptedInputs, setEncryptedInputs] = React.useState('');

  // Load templates
  React.useEffect(() => {
    if (!initialTemplate) {
      getWorkloadTemplates().then(setTemplates);
    }
  }, [initialTemplate, getWorkloadTemplates]);

  /**
   * Handle template selection
   */
  const handleSelectTemplate = (template: WorkloadTemplate) => {
    setSelectedTemplate(template);
    setName(`${template.name} Job`);
    setCpu(template.defaultResources.cpu);
    setMemory(template.defaultResources.memory);
    setGpu(template.defaultResources.gpu || 0);
    setRequiresTee(template.supportsTee);
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

      let sanitizedEncryptedInputs: Record<string, unknown> | null = null;
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

      const manifest: JobManifest = {
        templateId: selectedTemplate.id,
        name: sanitizedName,
        resources: {
          cpu,
          memory,
          gpu: gpu > 0 ? gpu : undefined,
        },
        maxDurationSeconds: duration,
        requiresTee,
        encryptedInputs: sanitizedEncryptedInputs || undefined,
      };

      // Start submission to get pending job
      await startJobSubmission(manifest);

      // Get price quote
      const priceQuote = await getQuote(manifest);
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
    if (!state.submission?.pendingJob) return;

    setStep('submitting');
    setError(null);

    try {
      const result = await submitJob(state.submission.pendingJob);
      onSubmit?.(result.jobId);
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
            cpu={cpu}
            onCpuChange={setCpu}
            memory={memory}
            onMemoryChange={setMemory}
            gpu={gpu}
            onGpuChange={setGpu}
            duration={duration}
            onDurationChange={setDuration}
            requiresTee={requiresTee}
            onRequiresTeeChange={setRequiresTee}
            encryptedInputs={encryptedInputs}
            onEncryptedInputsChange={setEncryptedInputs}
            onContinue={handleConfigComplete}
            onBack={() => setStep('template')}
            isLoading={state.isLoading}
            error={error}
          />
        )}

        {step === 'review' && quote && (
          <ReviewStep
            manifest={state.submission?.pendingJob?.manifest}
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
            <span className="job-form__template-icon">{template.category === 'ml-training' ? '🧠' : template.category === 'data-processing' ? '📊' : '🖥️'}</span>
            <span className="job-form__template-name">{template.name}</span>
            <span className="job-form__template-desc">{template.description}</span>
            {template.supportsTee && (
              <span className="job-form__template-tee">🔒 TEE</span>
            )}
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
  cpu: number;
  onCpuChange: (cpu: number) => void;
  memory: number;
  onMemoryChange: (memory: number) => void;
  gpu: number;
  onGpuChange: (gpu: number) => void;
  duration: number;
  onDurationChange: (duration: number) => void;
  requiresTee: boolean;
  onRequiresTeeChange: (requiresTee: boolean) => void;
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
  cpu,
  onCpuChange,
  memory,
  onMemoryChange,
  gpu,
  onGpuChange,
  duration,
  onDurationChange,
  requiresTee,
  onRequiresTeeChange,
  encryptedInputs,
  onEncryptedInputsChange,
  onContinue,
  onBack,
  isLoading,
  error,
}: ConfigurationFormProps): JSX.Element {
  return (
    <div className="job-form__config">
      <h3 className="job-form__title">Configure {template.name}</h3>

      <div className="job-form__field">
        <label className="job-form__label">Job Name</label>
        <input
          type="text"
          className="job-form__input"
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
          placeholder="Enter job name"
        />
      </div>

      <div className="job-form__field-row">
        <div className="job-form__field">
          <label className="job-form__label">CPU Cores</label>
          <input
            type="number"
            className="job-form__input"
            value={cpu}
            onChange={(e) => onCpuChange(parseInt(e.target.value, 10))}
            min={template.defaultResources.cpu}
          />
        </div>
        <div className="job-form__field">
          <label className="job-form__label">Memory (GB)</label>
          <input
            type="number"
            className="job-form__input"
            value={memory}
            onChange={(e) => onMemoryChange(parseInt(e.target.value, 10))}
            min={template.defaultResources.memory}
          />
        </div>
      </div>

      <div className="job-form__field-row">
        <div className="job-form__field">
          <label className="job-form__label">GPU Count</label>
          <input
            type="number"
            className="job-form__input"
            value={gpu}
            onChange={(e) => onGpuChange(parseInt(e.target.value, 10))}
            min={0}
          />
        </div>
        <div className="job-form__field">
          <label className="job-form__label">Max Duration</label>
          <select
            className="job-form__select"
            value={duration}
            onChange={(e) => onDurationChange(parseInt(e.target.value, 10))}
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

      {template.supportsTee && (
        <div className="job-form__field">
          <label className="job-form__checkbox-label">
            <input
              type="checkbox"
              checked={requiresTee}
              onChange={(e) => onRequiresTeeChange(e.target.checked)}
            />
            <span>Require TEE (Trusted Execution Environment)</span>
          </label>
        </div>
      )}

      <div className="job-form__field">
        <label className="job-form__label">Encrypted Inputs (JSON, optional)</label>
        <textarea
          className="job-form__textarea"
          value={encryptedInputs}
          onChange={(e) => onEncryptedInputsChange(e.target.value)}
          placeholder='{"key": "encrypted_value"}'
          rows={4}
        />
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
          onClick={onContinue}
          disabled={!name || isLoading}
        >
          {isLoading ? 'Getting Quote...' : 'Continue'}
        </button>
      </div>
    </div>
  );
}

/**
 * Review step
 */
interface ReviewStepProps {
  manifest?: JobManifest;
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
  if (!manifest) return <div>No manifest</div>;

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
        <div className="job-form__summary-row">
          <span className="job-form__summary-label">Resources</span>
          <span className="job-form__summary-value">
            {manifest.resources.cpu} CPU, {manifest.resources.memory}GB RAM
            {manifest.resources.gpu ? `, ${manifest.resources.gpu} GPU` : ''}
          </span>
        </div>
        <div className="job-form__summary-row">
          <span className="job-form__summary-label">Max Duration</span>
          <span className="job-form__summary-value">
            {formatDuration(manifest.maxDurationSeconds)}
          </span>
        </div>
        <div className="job-form__summary-row">
          <span className="job-form__summary-label">TEE Required</span>
          <span className="job-form__summary-value">
            {manifest.requiresTee ? 'Yes' : 'No'}
          </span>
        </div>
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

  .job-form__input,
  .job-form__select,
  .job-form__textarea {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-size: 0.875rem;
  }

  .job-form__input:focus,
  .job-form__select:focus,
  .job-form__textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .job-form__checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    color: #374151;
    cursor: pointer;
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
    color: #dc2626;
    font-size: 0.875rem;
    margin: 16px 0;
    padding: 12px;
    background: #fef2f2;
    border-radius: 6px;
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
    border: 1px solid #e5e7eb;
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
