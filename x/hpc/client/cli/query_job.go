package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"

	hpctypes "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// NewCmdQueryJob queries a job by ID.
func NewCmdQueryJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [job-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query an HPC job by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			resp, err := queryClient.Job(cmd.Context(), &hpctypes.QueryJobRequest{
				JobId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryJobs lists jobs with optional filters.
func NewCmdQueryJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Args:  cobra.NoArgs,
		Short: "List HPC jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			owner, err := cmd.Flags().GetString(flagOwnerAddr)
			if err != nil {
				return err
			}
			customer, err := cmd.Flags().GetString(flagSubmitter)
			if err != nil {
				return err
			}
			provider, err := cmd.Flags().GetString(flagProvider)
			if err != nil {
				return err
			}
			clusterID, err := cmd.Flags().GetString(flagClusterID)
			if err != nil {
				return err
			}
			queueName, err := cmd.Flags().GetString(flagQueueName)
			if err != nil {
				return err
			}
			status, err := cmd.Flags().GetString(flagStatus)
			if err != nil {
				return err
			}

			ownerFilter, err := resolveOwnerFilter(owner, customer)
			if err != nil {
				return err
			}

			var stateFilter *hpctypes.JobState
			if status != "" {
				parsedState, ok := parseJobStateFilter(status)
				if !ok {
					return fmt.Errorf("invalid job status: %s", status)
				}
				stateFilter = &parsedState
			}

			requiresLocalFilter := ownerFilter != "" || provider != "" || queueName != ""
			queryClient := hpctypes.NewQueryClient(clientCtx)

			if !requiresLocalFilter {
				resp, err := queryClient.Jobs(cmd.Context(), &hpctypes.QueryJobsRequest{
					State:      stateFilterValue(stateFilter),
					ClusterId:  strings.TrimSpace(clusterID),
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(resp)
			}

			if len(pageReq.Key) > 0 {
				return fmt.Errorf("page-key pagination is not supported with filters")
			}

			var jobs []hpctypes.HPCJob
			switch {
			case ownerFilter != "":
				resp, err := queryClient.JobsByCustomer(cmd.Context(), &hpctypes.QueryJobsByCustomerRequest{
					CustomerAddress: ownerFilter,
					Pagination:      nil,
				})
				if err != nil {
					return err
				}
				jobs = resp.Jobs
			case provider != "":
				resp, err := queryClient.JobsByProvider(cmd.Context(), &hpctypes.QueryJobsByProviderRequest{
					ProviderAddress: provider,
					Pagination:      nil,
				})
				if err != nil {
					return err
				}
				jobs = resp.Jobs
			default:
				resp, err := queryClient.Jobs(cmd.Context(), &hpctypes.QueryJobsRequest{
					State:      stateFilterValue(stateFilter),
					ClusterId:  strings.TrimSpace(clusterID),
					Pagination: nil,
				})
				if err != nil {
					return err
				}
				jobs = resp.Jobs
			}

			filtered := filterJobsByFields(jobs, jobFilterOptions{
				state:     stateFilter,
				provider:  provider,
				clusterID: clusterID,
				queueName: queueName,
				owner:     ownerFilter,
			})
			pageJobs, pageResp := paginateSlice(filtered, pageReq)

			return clientCtx.PrintProto(&hpctypes.QueryJobsResponse{
				Jobs:       pageJobs,
				Pagination: pageResp,
			})
		},
	}

	cmd.Flags().String(flagOwnerAddr, "", "Filter jobs by owner (customer) address")
	cmd.Flags().String(flagSubmitter, "", "Filter jobs by customer address")
	cmd.Flags().String(flagProvider, "", "Filter jobs by provider address")
	cmd.Flags().String(flagClusterID, "", "Filter jobs by cluster ID")
	cmd.Flags().String(flagQueueName, "", "Filter jobs by queue name")
	cmd.Flags().String(flagStatus, "", "Filter jobs by status")
	flags.AddPaginationFlagsToCmd(cmd, "jobs")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryJobLogs returns job timeline information.
func NewCmdQueryJobLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job-logs [job-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query job status timeline and messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			resp, err := queryClient.Job(cmd.Context(), &hpctypes.QueryJobRequest{
				JobId: args[0],
			})
			if err != nil {
				return err
			}

			logs := buildJobLogEntries(resp.Job)
			payload := JobLogsResponse{
				JobId:         resp.Job.JobId,
				State:         resp.Job.State,
				StatusMessage: resp.Job.StatusMessage,
				ExitCode:      resp.Job.ExitCode,
				Events:        logs,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryJobResult returns job outputs and accounting info.
func NewCmdQueryJobResult() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job-result [job-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query job outputs and accounting details",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			jobResp, err := queryClient.Job(cmd.Context(), &hpctypes.QueryJobRequest{
				JobId: args[0],
			})
			if err != nil {
				return err
			}

			var accounting *hpctypes.JobAccounting
			acctResp, err := queryClient.JobAccounting(cmd.Context(), &hpctypes.QueryJobAccountingRequest{
				JobId: args[0],
			})
			if err == nil {
				accounting = &acctResp.Accounting
			} else if !isNotFoundError(err) {
				return err
			}

			payload := JobResultResponse{
				Job:             jobResp.Job,
				Accounting:      accounting,
				OutputPointer:   strings.TrimSpace(jobResp.Job.EncryptedOutputsPointer),
				OutputLocation:  parseOutputLocation(jobResp.Job.StatusMessage),
				StatusMessage:   jobResp.Job.StatusMessage,
				ExitCode:        jobResp.Job.ExitCode,
				CompletionState: jobResp.Job.State,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewCmdQueryUsage returns usage history for a customer address.
func NewCmdQueryUsage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query HPC usage history for a customer address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			status, err := cmd.Flags().GetString(flagStatus)
			if err != nil {
				return err
			}
			clusterID, err := cmd.Flags().GetString(flagClusterID)
			if err != nil {
				return err
			}
			queueName, err := cmd.Flags().GetString(flagQueueName)
			if err != nil {
				return err
			}

			var stateFilter *hpctypes.JobState
			if status != "" {
				parsedState, ok := parseJobStateFilter(status)
				if !ok {
					return fmt.Errorf("invalid job status: %s", status)
				}
				stateFilter = &parsedState
			}

			queryClient := hpctypes.NewQueryClient(clientCtx)
			requiresLocalFilter := clusterID != "" || queueName != "" || stateFilter != nil

			var jobs []hpctypes.HPCJob
			var pagination *query.PageResponse
			if !requiresLocalFilter {
				resp, err := queryClient.JobsByCustomer(cmd.Context(), &hpctypes.QueryJobsByCustomerRequest{
					CustomerAddress: args[0],
					Pagination:      pageReq,
				})
				if err != nil {
					return err
				}
				jobs = resp.Jobs
				pagination = resp.Pagination
			} else {
				if len(pageReq.Key) > 0 {
					return fmt.Errorf("page-key pagination is not supported with filters")
				}
				resp, err := queryClient.JobsByCustomer(cmd.Context(), &hpctypes.QueryJobsByCustomerRequest{
					CustomerAddress: args[0],
					Pagination:      nil,
				})
				if err != nil {
					return err
				}
				filtered := filterJobsByFields(resp.Jobs, jobFilterOptions{
					state:     stateFilter,
					clusterID: clusterID,
					queueName: queueName,
					owner:     args[0],
				})
				pageJobs, pageResp := paginateSlice(filtered, pageReq)
				jobs = pageJobs
				pagination = pageResp
			}

			entries := make([]UsageEntry, 0, len(jobs))
			for _, job := range jobs {
				entry := UsageEntry{
					JobId:           job.JobId,
					ClusterId:       job.ClusterId,
					ProviderAddress: job.ProviderAddress,
					CustomerAddress: job.CustomerAddress,
					State:           job.State,
					QueueName:       job.QueueName,
					CreatedAt:       job.CreatedAt,
				}
				if job.CompletedAt != nil {
					entry.CompletedAt = job.CompletedAt
				}

				acctResp, err := queryClient.JobAccounting(cmd.Context(), &hpctypes.QueryJobAccountingRequest{
					JobId: job.JobId,
				})
				if err == nil {
					entry.UsageMetrics = &acctResp.Accounting.UsageMetrics
					entry.TotalCost = acctResp.Accounting.TotalCost
					entry.ProviderReward = acctResp.Accounting.ProviderReward
					entry.PlatformFee = acctResp.Accounting.PlatformFee
					entry.SignedUsageRecordIds = acctResp.Accounting.SignedUsageRecordIds
					entry.SettlementStatus = acctResp.Accounting.SettlementStatus
					entry.SettlementId = acctResp.Accounting.SettlementId
					entry.FinalizedAt = acctResp.Accounting.FinalizedAt
				} else if !isNotFoundError(err) {
					return err
				}
				entries = append(entries, entry)
			}

			payload := UsageResponse{
				Address:    args[0],
				Entries:    entries,
				Pagination: pagination,
			}

			return clientCtx.PrintObjectLegacy(payload)
		},
	}

	cmd.Flags().String(flagStatus, "", "Filter usage by job status")
	cmd.Flags().String(flagClusterID, "", "Filter usage by cluster ID")
	cmd.Flags().String(flagQueueName, "", "Filter usage by queue name")
	flags.AddPaginationFlagsToCmd(cmd, "usage")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

type jobFilterOptions struct {
	state     *hpctypes.JobState
	provider  string
	clusterID string
	queueName string
	owner     string
}

func filterJobsByFields(jobs []hpctypes.HPCJob, opts jobFilterOptions) []hpctypes.HPCJob {
	if opts.state == nil && opts.provider == "" && opts.clusterID == "" && opts.queueName == "" && opts.owner == "" {
		return jobs
	}

	filtered := make([]hpctypes.HPCJob, 0, len(jobs))
	for _, job := range jobs {
		if opts.state != nil && job.State != *opts.state {
			continue
		}
		if opts.provider != "" && job.ProviderAddress != opts.provider {
			continue
		}
		if opts.clusterID != "" && job.ClusterId != opts.clusterID {
			continue
		}
		if opts.queueName != "" && !strings.EqualFold(job.QueueName, opts.queueName) {
			continue
		}
		if opts.owner != "" && job.CustomerAddress != opts.owner {
			continue
		}
		filtered = append(filtered, job)
	}
	return filtered
}

func resolveOwnerFilter(owner, customer string) (string, error) {
	owner = strings.TrimSpace(owner)
	customer = strings.TrimSpace(customer)
	if owner == "" {
		return customer, nil
	}
	if customer == "" {
		return owner, nil
	}
	if owner != customer {
		return "", fmt.Errorf("--%s and --%s must match when both are set", flagOwnerAddr, flagSubmitter)
	}
	return owner, nil
}

type JobLogEntry struct {
	Event     string    `json:"event" yaml:"event"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	Detail    string    `json:"detail,omitempty" yaml:"detail,omitempty"`
}

type JobLogsResponse struct {
	JobId         string            `json:"job_id" yaml:"job_id"`
	State         hpctypes.JobState `json:"state" yaml:"state"`
	StatusMessage string            `json:"status_message,omitempty" yaml:"status_message,omitempty"`
	ExitCode      int32             `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	Events        []JobLogEntry     `json:"events" yaml:"events"`
}

type JobResultResponse struct {
	Job             hpctypes.HPCJob         `json:"job" yaml:"job"`
	Accounting      *hpctypes.JobAccounting `json:"accounting,omitempty" yaml:"accounting,omitempty"`
	OutputPointer   string                  `json:"output_pointer,omitempty" yaml:"output_pointer,omitempty"`
	OutputLocation  string                  `json:"output_location,omitempty" yaml:"output_location,omitempty"`
	StatusMessage   string                  `json:"status_message,omitempty" yaml:"status_message,omitempty"`
	ExitCode        int32                   `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	CompletionState hpctypes.JobState       `json:"completion_state" yaml:"completion_state"`
}

type UsageEntry struct {
	JobId                string                    `json:"job_id" yaml:"job_id"`
	ClusterId            string                    `json:"cluster_id" yaml:"cluster_id"`
	ProviderAddress      string                    `json:"provider_address" yaml:"provider_address"`
	CustomerAddress      string                    `json:"customer_address" yaml:"customer_address"`
	State                hpctypes.JobState         `json:"state" yaml:"state"`
	QueueName            string                    `json:"queue_name,omitempty" yaml:"queue_name,omitempty"`
	UsageMetrics         *hpctypes.HPCUsageMetrics `json:"usage_metrics,omitempty" yaml:"usage_metrics,omitempty"`
	TotalCost            sdk.Coins                 `json:"total_cost,omitempty" yaml:"total_cost,omitempty"`
	ProviderReward       sdk.Coins                 `json:"provider_reward,omitempty" yaml:"provider_reward,omitempty"`
	PlatformFee          sdk.Coins                 `json:"platform_fee,omitempty" yaml:"platform_fee,omitempty"`
	SignedUsageRecordIds []string                  `json:"signed_usage_record_ids,omitempty" yaml:"signed_usage_record_ids,omitempty"`
	SettlementStatus     string                    `json:"settlement_status,omitempty" yaml:"settlement_status,omitempty"`
	SettlementId         string                    `json:"settlement_id,omitempty" yaml:"settlement_id,omitempty"`
	CreatedAt            time.Time                 `json:"created_at" yaml:"created_at"`
	CompletedAt          *time.Time                `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	FinalizedAt          *time.Time                `json:"finalized_at,omitempty" yaml:"finalized_at,omitempty"`
}

type UsageResponse struct {
	Address    string              `json:"address" yaml:"address"`
	Entries    []UsageEntry        `json:"entries" yaml:"entries"`
	Pagination *query.PageResponse `json:"pagination,omitempty" yaml:"pagination,omitempty"`
}

func buildJobLogEntries(job hpctypes.HPCJob) []JobLogEntry {
	entries := make([]JobLogEntry, 0, 4)
	if !job.CreatedAt.IsZero() {
		entries = append(entries, JobLogEntry{Event: "created", Timestamp: job.CreatedAt})
	}
	if job.QueuedAt != nil && !job.QueuedAt.IsZero() {
		entries = append(entries, JobLogEntry{Event: "queued", Timestamp: *job.QueuedAt})
	}
	if job.StartedAt != nil && !job.StartedAt.IsZero() {
		entries = append(entries, JobLogEntry{Event: "started", Timestamp: *job.StartedAt})
	}
	if job.CompletedAt != nil && !job.CompletedAt.IsZero() {
		entries = append(entries, JobLogEntry{Event: "completed", Timestamp: *job.CompletedAt})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})
	return entries
}

func parseOutputLocation(statusMessage string) string {
	for _, field := range strings.Fields(statusMessage) {
		if strings.HasPrefix(field, "output=") {
			return strings.TrimPrefix(field, "output=")
		}
	}
	return ""
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.NotFound
}

func stateFilterValue(state *hpctypes.JobState) hpctypes.JobState {
	if state == nil {
		return hpctypes.JobStateUnspecified
	}
	return *state
}
