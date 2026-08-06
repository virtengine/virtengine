// @ts-nocheck
/**
 * useHPC Hook
 * VE-705: Supercomputer/HPC UI (job submission, library workloads, outputs)
 */

import {
  useState,
  useCallback,
  useEffect,
  useContext,
  createContext,
  useRef,
} from "react";
import type { ReactNode } from "react";
import type {
  HPCState,
  WorkloadTemplate,
  JobManifest,
  JobSubmission,
  Job,
  JobStatus,
  JobOutput,
  JobOutputReference,
  JobPriceQuote,
  JobValidationError,
} from "../types/hpc";
import { initialHPCState } from "../types/hpc";
import type { QueryClient, ChainEvent } from "../types/chain";
import { sanitizePlainText, sanitizeObject } from "../utils/security";
import {
  HPCClientUnavailableError,
  assertValidSubmitJobParams,
  assertCommittedJobMutation,
  requireHPCSigner,
  type CommittedJobMutation,
  type HPCSignerAdapter,
  type SubmitJobParams,
} from "../components/hpc/hpc-mutation";
import {
  requireHPCOutputAdapter,
  validateHPCOutputReferences,
  validateResolvedHPCOutput,
  type HPCOutputAdapter,
} from "../components/hpc/hpc-output";
import {
  HPCQueryValidationError,
  requireHPCQueryAdapter,
  validateHPCJob,
  validateHPCJobPriceQuote,
  validateHPCQuoteRequest,
  validateHPCJobSubscriptionEvent,
  validateHPCJobs,
  validateHPCWorkloadTemplates,
  type HPCQueryAdapter,
  type HPCQuoteRequest,
} from "../components/hpc/hpc-query";

interface HPCContextValue {
  state: HPCState;
  actions: HPCActions;
}

interface HPCActions {
  refresh: () => Promise<void>;
  getWorkloadTemplates: () => Promise<void>;
  startJobSubmission: (templateId?: string) => void;
  updateJobManifest: (manifest: Partial<JobManifest>) => void;
  selectOffering: (offeringId: string) => void;
  getQuote: (
    request?: JobManifest["resources"] | HPCQuoteRequest,
  ) => Promise<JobPriceQuote>;
  validateJob: () => JobValidationError[];
  submitJob: () => Promise<CommittedJobMutation>;
  cancelSubmission: () => void;
  getJobs: () => Promise<void>;
  getJob: (jobId: string) => Promise<Job>;
  cancelJob: (jobId: string) => Promise<CommittedJobMutation>;
  getOutputs: (jobId: string) => Promise<readonly JobOutputReference[]>;
  decryptOutput: (
    jobId: string,
    outputRef: JobOutputReference,
  ) => Promise<JobOutput>;
  subscribeToJob: (
    jobId: string,
    callback: (event: ChainEvent) => void,
  ) => () => void;
  clearError: () => void;
}

const sameResources = (
  left: JobManifest["resources"],
  right: JobManifest["resources"],
): boolean =>
  left.nodes === right.nodes &&
  left.cpusPerNode === right.cpusPerNode &&
  left.memoryGBPerNode === right.memoryGBPerNode &&
  left.gpusPerNode === right.gpusPerNode &&
  left.gpuType === right.gpuType &&
  left.maxRuntimeSeconds === right.maxRuntimeSeconds &&
  left.storageGB === right.storageGB;

const HPCContext = createContext<HPCContextValue | null>(null);

export interface HPCProviderProps {
  children: ReactNode;
  queryClient: QueryClient;
  chainId: string;
  accountAddress: string | null;
  getAuthHeader?: () => Promise<string>;
  mutationAdapter?: HPCSignerAdapter;
  outputAdapter?: HPCOutputAdapter;
  queryAdapter?: HPCQueryAdapter;
}

export function HPCProvider({
  children,
  queryClient,
  chainId,
  accountAddress,
  getAuthHeader,
  mutationAdapter,
  outputAdapter,
  queryAdapter,
}: HPCProviderProps) {
  const [state, setState] = useState<HPCState>(initialHPCState);
  const submissionToken = useRef(0);
  const submissionInFlight = useRef(false);
  const mutationGeneration = useRef(0);
  const mutationAuthority = useRef({
    mutationAdapter,
    chainId,
    accountAddress,
  });
  const cancellationsInFlight = useRef(new Set<string>());
  const outputGeneration = useRef(0);
  const outputAuthority = useRef({ outputAdapter, chainId, accountAddress });
  const queryGeneration = useRef(0);
  const queryAuthority = useRef({ queryAdapter, chainId, accountAddress });
  const queryStateGeneration = useRef(0);
  const queryResetPending = useRef(false);
  const templateRequest = useRef(0);
  const jobsRequest = useRef(0);
  const jobRequest = useRef(0);
  const quoteRequest = useRef(0);
  const quotedSubmission = useRef<{
    submissionId: number;
    offeringId: string;
    resources: JobManifest["resources"];
    quote: JobPriceQuote;
  } | null>(null);
  const activeQuerySubscriptions = useRef(new Set<() => void>());
  const activeQueryOperations = useRef(new Set<string>());
  const queryErrors = useRef(new Map<string, HPCState["error"]>());

  if (
    mutationAuthority.current.mutationAdapter !== mutationAdapter ||
    mutationAuthority.current.chainId !== chainId ||
    mutationAuthority.current.accountAddress !== accountAddress
  ) {
    mutationAuthority.current = { mutationAdapter, chainId, accountAddress };
    mutationGeneration.current += 1;
    submissionToken.current += 1;
    submissionInFlight.current = false;
    cancellationsInFlight.current.clear();
  }
  const renderMutationGeneration = mutationGeneration.current;

  if (
    outputAuthority.current.outputAdapter !== outputAdapter ||
    outputAuthority.current.chainId !== chainId ||
    outputAuthority.current.accountAddress !== accountAddress
  ) {
    outputAuthority.current = { outputAdapter, chainId, accountAddress };
    outputGeneration.current += 1;
  }
  const renderOutputGeneration = outputGeneration.current;

  if (
    queryAuthority.current.queryAdapter !== queryAdapter ||
    queryAuthority.current.chainId !== chainId ||
    queryAuthority.current.accountAddress !== accountAddress
  ) {
    queryAuthority.current = { queryAdapter, chainId, accountAddress };
    queryGeneration.current += 1;
    templateRequest.current += 1;
    jobsRequest.current += 1;
    jobRequest.current += 1;
    quoteRequest.current += 1;
    quotedSubmission.current = null;
    queryResetPending.current = true;
    activeQueryOperations.current.clear();
    queryErrors.current.clear();
  }
  const renderQueryGeneration = queryGeneration.current;
  const effectiveState: HPCState =
    queryStateGeneration.current === renderQueryGeneration
      ? state
      : {
          ...state,
          workloadTemplates: [],
          jobs: [],
          selectedJob: null,
          submission: null,
        };

  const beginQueryOperation = useCallback((operation: string) => {
    activeQueryOperations.current.add(operation);
    queryErrors.current.delete(operation);
    setState((prev) => ({
      ...prev,
      isLoading: true,
      error: Array.from(queryErrors.current.values()).at(-1) ?? null,
    }));
  }, []);

  const finishQueryOperation = useCallback(
    (operation: string, error?: unknown) => {
      activeQueryOperations.current.delete(operation);
      if (error !== undefined) {
        queryErrors.current.set(operation, {
          code: "network_error",
          message:
            error instanceof Error
              ? error.message
              : `HPC ${operation} query failed`,
        });
      }
      setState((prev) => ({
        ...prev,
        isLoading: activeQueryOperations.current.size > 0,
        error: Array.from(queryErrors.current.values()).at(-1) ?? null,
      }));
    },
    [renderQueryGeneration],
  );

  useEffect(() => {
    if (!queryResetPending.current) return;
    queryResetPending.current = false;
    queryStateGeneration.current = renderQueryGeneration;
    setState((prev) => ({
      ...prev,
      workloadTemplates: [],
      jobs: [],
      selectedJob: null,
      submission: null,
      isLoading: false,
      error: null,
    }));
  }, [renderQueryGeneration]);

  useEffect(
    () => () => {
      for (const unsubscribe of activeQuerySubscriptions.current) unsubscribe();
      activeQuerySubscriptions.current.clear();
    },
    [renderQueryGeneration],
  );

  const getWorkloadTemplates = useCallback(async () => {
    const generation = renderQueryGeneration;
    if (!accountAddress || generation !== queryGeneration.current) {
      throw new HPCClientUnavailableError("query");
    }
    const requestId = ++templateRequest.current;
    const binding = { chainId, accountAddress };
    const adapter = requireHPCQueryAdapter(queryAdapter, binding);
    beginQueryOperation("templates");

    try {
      const evidence = await adapter.getWorkloadTemplates();
      if (
        generation !== queryGeneration.current ||
        requestId !== templateRequest.current
      ) {
        throw new HPCClientUnavailableError("query");
      }
      const templates = validateHPCWorkloadTemplates(evidence, binding);

      setState((prev) => ({
        ...prev,
        workloadTemplates: templates,
      }));
      finishQueryOperation("templates");
    } catch (error) {
      if (
        generation !== queryGeneration.current ||
        requestId !== templateRequest.current
      ) {
        throw error;
      }
      finishQueryOperation("templates", error);
      throw error;
    }
  }, [
    accountAddress,
    beginQueryOperation,
    chainId,
    finishQueryOperation,
    queryAdapter,
    renderQueryGeneration,
  ]);

  const refresh = useCallback(async () => {
    await Promise.all([getWorkloadTemplates(), getJobs()]);
  }, [getWorkloadTemplates]);

  const startJobSubmission = useCallback(
    (templateId?: string) => {
      if (renderQueryGeneration !== queryGeneration.current) {
        throw new HPCClientUnavailableError("query");
      }
      submissionToken.current += 1;
      quoteRequest.current += 1;
      quotedSubmission.current = null;
      submissionInFlight.current = false;
      const template = templateId
        ? effectiveState.workloadTemplates.find((t) => t.id === templateId)
        : null;

      const defaultParams: Record<string, string | number | boolean> = {};
      if (template) {
        for (const [key, param] of Object.entries(template.defaultParameters)) {
          if (param.defaultValue !== undefined) {
            defaultParams[key] = param.defaultValue;
          }
        }
      }

      setState((prev) => ({
        ...prev,
        submission: {
          step: templateId ? "configure" : "select_template",
          manifest: template
            ? {
                version: "1.0.0",
                name: "",
                templateId: template.id,
                resources: template.defaultResources,
                parameters: defaultParams,
              }
            : {
                version: "1.0.0",
                name: "",
                resources: {
                  nodes: 1,
                  cpusPerNode: 4,
                  memoryGBPerNode: 16,
                  maxRuntimeSeconds: 3600,
                  storageGB: 50,
                },
                parameters: {},
              },
          selectedTemplate: template,
          selectedOffering: null,
          priceQuote: null,
          validationErrors: [],
          error: null,
        },
      }));
    },
    [effectiveState.workloadTemplates, renderQueryGeneration],
  );

  const updateJobManifest = useCallback(
    (manifestUpdate: Partial<JobManifest>) => {
      if (renderQueryGeneration !== queryGeneration.current) {
        throw new HPCClientUnavailableError("query");
      }
      const sanitizedUpdate: Partial<JobManifest> = { ...manifestUpdate };

      if (typeof manifestUpdate.name === "string") {
        sanitizedUpdate.name = sanitizePlainText(manifestUpdate.name, {
          maxLength: 120,
        });
      }

      if (typeof manifestUpdate.description === "string") {
        sanitizedUpdate.description = sanitizePlainText(
          manifestUpdate.description,
          {
            maxLength: 500,
          },
        );
      }

      if (typeof manifestUpdate.command === "string") {
        sanitizedUpdate.command = sanitizePlainText(manifestUpdate.command, {
          maxLength: 300,
        });
      }

      if (typeof manifestUpdate.image === "string") {
        sanitizedUpdate.image = sanitizePlainText(manifestUpdate.image, {
          maxLength: 200,
        });
      }

      if (manifestUpdate.environment) {
        sanitizedUpdate.environment = sanitizeObject(
          manifestUpdate.environment,
          {
            maxDepth: 2,
            maxKeyLength: 64,
            maxStringLength: 256,
            escapeHtmlStrings: false,
          },
        ) as Record<string, string>;
      }

      if (manifestUpdate.parameters) {
        sanitizedUpdate.parameters = sanitizeObject(manifestUpdate.parameters, {
          maxDepth: 2,
          maxKeyLength: 64,
          maxStringLength: 256,
          escapeHtmlStrings: false,
        }) as Record<string, string | number | boolean>;
      }

      if (manifestUpdate.resources) {
        sanitizedUpdate.resources = Object.freeze({
          ...manifestUpdate.resources,
        });
      }

      quoteRequest.current += 1;
      quotedSubmission.current = null;

      setState((prev) => ({
        ...prev,
        submission: prev.submission
          ? {
              ...prev.submission,
              manifest: { ...prev.submission.manifest, ...sanitizedUpdate },
              priceQuote: null,
            }
          : null,
      }));
    },
    [],
  );

  const selectOffering = useCallback(
    (offeringId: string) => {
      if (renderQueryGeneration !== queryGeneration.current) {
        throw new HPCClientUnavailableError("query");
      }
      quoteRequest.current += 1;
      quotedSubmission.current = null;
      setState((prev) => ({
        ...prev,
        submission: prev.submission
          ? {
              ...prev.submission,
              selectedOffering: offeringId,
              step: "review",
              priceQuote: null,
            }
          : null,
      }));
    },
    [renderQueryGeneration],
  );

  const getQuote = useCallback(
    async (
      requestSnapshot?: JobManifest["resources"] | HPCQuoteRequest,
    ): Promise<JobPriceQuote> => {
      const explicitRequest =
        requestSnapshot && "offeringId" in requestSnapshot
          ? requestSnapshot
          : undefined;
      const resources =
        explicitRequest?.resources ??
        requestSnapshot ??
        effectiveState.submission?.manifest.resources;
      const offeringId =
        explicitRequest?.offeringId ??
        effectiveState.submission?.selectedOffering;
      if (!resources || !offeringId) {
        throw new Error("No job configured");
      }
      if (
        !accountAddress ||
        renderQueryGeneration !== queryGeneration.current
      ) {
        throw new HPCClientUnavailableError("query");
      }
      const binding = { chainId, accountAddress };
      const generation = renderQueryGeneration;
      const submissionId = submissionToken.current;
      const requestId = ++quoteRequest.current;
      const request = validateHPCQuoteRequest({
        offeringId,
        resources,
      });
      const adapter = requireHPCQueryAdapter(queryAdapter, binding);
      beginQueryOperation("quote");
      let quote: JobPriceQuote;
      try {
        const evidence = await adapter.getQuote(request);
        if (
          generation !== queryGeneration.current ||
          requestId !== quoteRequest.current ||
          submissionId !== submissionToken.current
        ) {
          throw new HPCClientUnavailableError("query");
        }
        quote = validateHPCJobPriceQuote(evidence, binding, request);
        quotedSubmission.current = Object.freeze({
          submissionId,
          offeringId,
          resources: request.resources,
          quote,
        });
        finishQueryOperation("quote");
      } catch (error) {
        if (
          generation === queryGeneration.current &&
          requestId === quoteRequest.current &&
          submissionId === submissionToken.current
        ) {
          finishQueryOperation("quote", error);
        }
        throw error;
      }

      setState((prev) => ({
        ...prev,
        submission:
          prev.submission?.selectedOffering === offeringId &&
          prev.submission.manifest.resources &&
          JSON.stringify(prev.submission.manifest.resources) ===
            JSON.stringify(resources)
            ? { ...prev.submission, priceQuote: quote }
            : null,
      }));

      return quote;
    },
    [
      accountAddress,
      beginQueryOperation,
      chainId,
      finishQueryOperation,
      queryAdapter,
      renderQueryGeneration,
      effectiveState.submission?.manifest.resources,
      effectiveState.submission?.selectedOffering,
    ],
  );

  const validateJob = useCallback((): JobValidationError[] => {
    const errors: JobValidationError[] = [];
    const manifest = state.submission?.manifest;

    if (!manifest?.name) {
      errors.push({ field: "name", message: "Job name is required" });
    }

    if (
      !manifest?.resources.maxRuntimeSeconds ||
      manifest.resources.maxRuntimeSeconds < 60
    ) {
      errors.push({
        field: "resources.maxRuntimeSeconds",
        message: "Runtime must be at least 60 seconds",
      });
    }

    setState((prev) => ({
      ...prev,
      submission: prev.submission
        ? { ...prev.submission, validationErrors: errors }
        : null,
    }));

    return errors;
  }, [state.submission?.manifest]);

  const submitJob = useCallback(async (): Promise<CommittedJobMutation> => {
    if (submissionInFlight.current) throw new Error("submission_in_progress");
    if (
      !accountAddress ||
      renderMutationGeneration !== mutationGeneration.current ||
      mutationAuthority.current.mutationAdapter !== mutationAdapter ||
      mutationAuthority.current.chainId !== chainId ||
      mutationAuthority.current.accountAddress !== accountAddress
    ) {
      throw new HPCClientUnavailableError("signer");
    }
    const errors = validateJob();
    if (errors.length > 0) {
      throw new Error("Validation failed");
    }

    const submission = state.submission;
    const quoteBinding = quotedSubmission.current;
    if (
      !submission?.selectedOffering ||
      !submission.manifest.resources ||
      !submission.priceQuote ||
      !quoteBinding ||
      quoteBinding.submissionId !== submissionToken.current ||
      quoteBinding.offeringId !== submission.selectedOffering ||
      quoteBinding.quote !== submission.priceQuote ||
      !sameResources(quoteBinding.resources, submission.manifest.resources)
    ) {
      throw new Error("feature_unavailable");
    }
    const params: SubmitJobParams = Object.freeze({
      offeringId: submission.selectedOffering,
      name: submission.manifest.name || "",
      description: submission.manifest.description,
      templateId: submission.manifest.templateId,
      resources: Object.freeze({ ...submission.manifest.resources }),
      command: submission.manifest.command,
      containerImage: submission.manifest.image,
      environment: submission.manifest.environment
        ? Object.freeze({ ...submission.manifest.environment })
        : undefined,
      parameters: submission.manifest.parameters
        ? Object.freeze({ ...submission.manifest.parameters })
        : undefined,
      encryptedInputs: submission.manifest.encryptedInputs
        ? Object.freeze({ ...submission.manifest.encryptedInputs })
        : undefined,
      inputRefs: submission.manifest.inputRefs
        ? Object.freeze([...submission.manifest.inputRefs])
        : undefined,
      quote: Object.freeze({
        estimatedTotal: submission.priceQuote.estimatedTotal,
        depositRequired: submission.priceQuote.depositRequired,
        pricePerHour: submission.priceQuote.pricePerHour,
        maxHours: submission.priceQuote.maxHours,
        denom: submission.priceQuote.denom,
      }),
    });
    assertValidSubmitJobParams(params, true);
    submissionInFlight.current = true;
    const token = ++submissionToken.current;
    const generation = renderMutationGeneration;

    setState((prev) => ({
      ...prev,
      submission: prev.submission
        ? { ...prev.submission, step: "submit" }
        : null,
    }));

    try {
      if (
        !accountAddress ||
        generation !== mutationGeneration.current ||
        mutationAuthority.current.mutationAdapter !== mutationAdapter ||
        mutationAuthority.current.chainId !== chainId ||
        mutationAuthority.current.accountAddress !== accountAddress
      ) {
        throw new HPCClientUnavailableError("signer");
      }
      const result = await requireHPCSigner(mutationAdapter, {
        chainId,
        accountAddress,
      }).submitJob(params);
      assertCommittedJobMutation(result);
      if (
        submissionToken.current !== token ||
        mutationGeneration.current !== generation
      ) {
        throw new Error("submission_cancelled");
      }
      setState((prev) => ({
        ...prev,
        submission: prev.submission
          ? { ...prev.submission, step: "complete", error: null }
          : null,
      }));
      return Object.freeze({ ...result });
    } catch (error) {
      if (
        submissionToken.current !== token ||
        mutationGeneration.current !== generation
      ) {
        throw error;
      }
      setState((prev) => ({
        ...prev,
        submission: prev.submission
          ? {
              ...prev.submission,
              step: "review",
              error: {
                code: "network_error",
                message:
                  error instanceof Error
                    ? error.message
                    : "Job submission failed",
              },
            }
          : null,
      }));
      throw error;
    } finally {
      if (submissionToken.current === token) submissionInFlight.current = false;
    }
  }, [
    accountAddress,
    chainId,
    mutationAdapter,
    renderMutationGeneration,
    validateJob,
    state.submission,
  ]);

  const cancelSubmission = useCallback(() => {
    if (renderQueryGeneration !== queryGeneration.current) {
      throw new HPCClientUnavailableError("query");
    }
    submissionToken.current += 1;
    quoteRequest.current += 1;
    quotedSubmission.current = null;
    submissionInFlight.current = false;
    setState((prev) => ({ ...prev, submission: null }));
  }, [renderQueryGeneration]);

  const getJobs = useCallback(async () => {
    if (!accountAddress || renderQueryGeneration !== queryGeneration.current) {
      throw new HPCClientUnavailableError("query");
    }
    const binding = { chainId, accountAddress };
    const generation = renderQueryGeneration;
    const requestId = ++jobsRequest.current;
    const adapter = requireHPCQueryAdapter(queryAdapter, binding);
    beginQueryOperation("jobs");
    try {
      const evidence = await adapter.getJobs();
      if (
        generation !== queryGeneration.current ||
        requestId !== jobsRequest.current
      ) {
        throw new HPCClientUnavailableError("query");
      }
      const jobs = validateHPCJobs(evidence, binding);
      setState((prev) => ({ ...prev, jobs: jobs as Job[] }));
      finishQueryOperation("jobs");
    } catch (error) {
      if (
        generation === queryGeneration.current &&
        requestId === jobsRequest.current
      ) {
        finishQueryOperation("jobs", error);
      }
      throw error;
    }
  }, [
    accountAddress,
    beginQueryOperation,
    chainId,
    finishQueryOperation,
    queryAdapter,
    renderQueryGeneration,
  ]);

  const getJob = useCallback(
    async (jobId: string): Promise<Job> => {
      if (
        !accountAddress ||
        renderQueryGeneration !== queryGeneration.current
      ) {
        throw new HPCClientUnavailableError("query");
      }
      const binding = { chainId, accountAddress, jobId };
      const generation = renderQueryGeneration;
      const requestId = ++jobRequest.current;
      const adapter = requireHPCQueryAdapter(queryAdapter, binding);
      beginQueryOperation("job");
      try {
        const evidence = await adapter.getJob(jobId);
        if (
          generation !== queryGeneration.current ||
          requestId !== jobRequest.current
        ) {
          throw new HPCClientUnavailableError("query");
        }
        const job = validateHPCJob(evidence, binding);
        setState((prev) => ({ ...prev, selectedJob: job }));
        finishQueryOperation("job");
        return job;
      } catch (error) {
        if (
          generation === queryGeneration.current &&
          requestId === jobRequest.current
        ) {
          finishQueryOperation("job", error);
        }
        throw error;
      }
    },
    [
      accountAddress,
      beginQueryOperation,
      chainId,
      finishQueryOperation,
      queryAdapter,
      renderQueryGeneration,
    ],
  );

  const cancelJob = useCallback(
    async (jobId: string): Promise<CommittedJobMutation> => {
      if (
        !accountAddress ||
        renderMutationGeneration !== mutationGeneration.current ||
        mutationAuthority.current.mutationAdapter !== mutationAdapter ||
        mutationAuthority.current.chainId !== chainId ||
        mutationAuthority.current.accountAddress !== accountAddress
      ) {
        throw new HPCClientUnavailableError("signer");
      }
      if (cancellationsInFlight.current.has(jobId))
        throw new Error("cancellation_in_progress");
      cancellationsInFlight.current.add(jobId);
      const generation = renderMutationGeneration;
      try {
        if (
          !accountAddress ||
          generation !== mutationGeneration.current ||
          mutationAuthority.current.mutationAdapter !== mutationAdapter ||
          mutationAuthority.current.chainId !== chainId ||
          mutationAuthority.current.accountAddress !== accountAddress
        ) {
          throw new HPCClientUnavailableError("signer");
        }
        const result = await requireHPCSigner(mutationAdapter, {
          chainId,
          accountAddress,
        }).cancelJob(jobId);
        assertCommittedJobMutation(result, jobId);
        if (mutationGeneration.current !== generation)
          throw new Error("submission_cancelled");
        return Object.freeze({ ...result });
      } finally {
        if (mutationGeneration.current === generation)
          cancellationsInFlight.current.delete(jobId);
      }
    },
    [accountAddress, chainId, mutationAdapter, renderMutationGeneration],
  );

  const getOutputs = useCallback(
    async (jobId: string): Promise<readonly JobOutputReference[]> => {
      if (!accountAddress) throw new HPCClientUnavailableError("provider");
      if (renderOutputGeneration !== outputGeneration.current) {
        throw new HPCClientUnavailableError("provider");
      }
      const binding = { chainId, accountAddress, jobId };
      const generation = renderOutputGeneration;
      const result = await requireHPCOutputAdapter(
        outputAdapter,
        binding,
      ).getOutputs(jobId);
      if (generation !== outputGeneration.current) {
        throw new HPCClientUnavailableError("provider");
      }
      return validateHPCOutputReferences(result, binding);
    },
    [accountAddress, chainId, outputAdapter, renderOutputGeneration],
  );

  const decryptOutput = useCallback(
    async (
      jobId: string,
      outputRef: JobOutputReference,
    ): Promise<JobOutput> => {
      if (!accountAddress) throw new HPCClientUnavailableError("provider");
      if (renderOutputGeneration !== outputGeneration.current) {
        throw new HPCClientUnavailableError("provider");
      }
      const binding = { chainId, accountAddress, jobId };
      const reference = validateHPCOutputReferences(
        { ...binding, outputs: [outputRef] },
        binding,
      )[0];
      const generation = renderOutputGeneration;
      const result = await requireHPCOutputAdapter(
        outputAdapter,
        binding,
      ).resolveOutput(reference);
      if (generation !== outputGeneration.current) {
        throw new HPCClientUnavailableError("provider");
      }
      return validateResolvedHPCOutput(result, reference, binding);
    },
    [accountAddress, chainId, outputAdapter, renderOutputGeneration],
  );

  const subscribeToJob = useCallback(
    (jobId: string, callback: (event: ChainEvent) => void): (() => void) => {
      if (
        !accountAddress ||
        renderQueryGeneration !== queryGeneration.current
      ) {
        throw new HPCClientUnavailableError("query");
      }
      const adapter = requireHPCQueryAdapter(queryAdapter, {
        chainId,
        accountAddress,
      });
      if (!adapter.subscribeToJob) {
        throw new HPCClientUnavailableError("query");
      }
      const generation = renderQueryGeneration;
      let stopped = false;
      const unsubscribe = adapter.subscribeToJob(jobId, (event) => {
        if (generation !== queryGeneration.current) {
          return;
        }
        callback(validateHPCJobSubscriptionEvent(event, jobId));
      });
      if (typeof unsubscribe !== "function") {
        throw new HPCQueryValidationError();
      }
      const stop = () => {
        if (stopped) return;
        stopped = true;
        activeQuerySubscriptions.current.delete(stop);
        unsubscribe();
      };
      activeQuerySubscriptions.current.add(stop);
      return stop;
    },
    [accountAddress, chainId, queryAdapter, renderQueryGeneration],
  );

  const clearError = useCallback(() => {
    queryErrors.current.clear();
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  useEffect(() => {
    void getWorkloadTemplates().catch(() => undefined);
  }, [getWorkloadTemplates]);

  useEffect(() => {
    if (accountAddress) {
      void getJobs().catch(() => undefined);
    }
  }, [accountAddress, getJobs]);

  const actions: HPCActions = {
    refresh,
    getWorkloadTemplates,
    startJobSubmission,
    updateJobManifest,
    selectOffering,
    getQuote,
    validateJob,
    submitJob,
    cancelSubmission,
    getJobs,
    getJob,
    cancelJob,
    getOutputs,
    decryptOutput,
    subscribeToJob,
    clearError,
  };

  return (
    <HPCContext.Provider value={{ state: effectiveState, actions }}>
      {children}
    </HPCContext.Provider>
  );
}

export function useHPC(): HPCContextValue {
  const context = useContext(HPCContext);
  if (!context) {
    throw new Error("useHPC must be used within an HPCProvider");
  }
  return context;
}
