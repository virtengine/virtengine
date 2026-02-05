// Package keeper implements the HPC module keeper.
//
// VE-500 through VE-504: HPC module keeper
package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// IKeeper defines the interface for the HPC keeper
type IKeeper interface {
	// Cluster management
	RegisterCluster(ctx sdk.Context, cluster *types.HPCCluster) error
	UpdateCluster(ctx sdk.Context, cluster *types.HPCCluster) error
	DeregisterCluster(ctx sdk.Context, clusterID string, providerAddr sdk.AccAddress) error
	GetCluster(ctx sdk.Context, clusterID string) (types.HPCCluster, bool)
	GetClustersByProvider(ctx sdk.Context, providerAddr sdk.AccAddress) []types.HPCCluster
	SetCluster(ctx sdk.Context, cluster types.HPCCluster) error

	// Offering management
	CreateOffering(ctx sdk.Context, offering *types.HPCOffering) error
	UpdateOffering(ctx sdk.Context, offering *types.HPCOffering) error
	GetOffering(ctx sdk.Context, offeringID string) (types.HPCOffering, bool)
	GetOfferingsByCluster(ctx sdk.Context, clusterID string) []types.HPCOffering
	SetOffering(ctx sdk.Context, offering types.HPCOffering) error

	// Job management
	SubmitJob(ctx sdk.Context, job *types.HPCJob) error
	UpdateJobStatus(ctx sdk.Context, jobID string, state types.JobState, statusMessage string, exitCode int32, metrics *types.HPCUsageMetrics) error
	CancelJob(ctx sdk.Context, jobID string, requesterAddr sdk.AccAddress) error
	GetJob(ctx sdk.Context, jobID string) (types.HPCJob, bool)
	GetJobsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress) []types.HPCJob
	GetJobsByCluster(ctx sdk.Context, clusterID string) []types.HPCJob
	SetJob(ctx sdk.Context, job types.HPCJob) error

	// Job accounting
	CreateJobAccounting(ctx sdk.Context, accounting *types.JobAccounting) error
	FinalizeJobAccounting(ctx sdk.Context, jobID string) error
	GetJobAccounting(ctx sdk.Context, jobID string) (types.JobAccounting, bool)
	SetJobAccounting(ctx sdk.Context, accounting types.JobAccounting) error

	// Node metadata
	UpdateNodeMetadata(ctx sdk.Context, node *types.NodeMetadata) error
	GetNodeMetadata(ctx sdk.Context, nodeID string) (types.NodeMetadata, bool)
	GetNodesByCluster(ctx sdk.Context, clusterID string) []types.NodeMetadata
	SetNodeMetadata(ctx sdk.Context, node types.NodeMetadata) error

	// Scheduling
	ScheduleJob(ctx sdk.Context, job *types.HPCJob) (*types.SchedulingDecision, error)
	GetSchedulingDecision(ctx sdk.Context, decisionID string) (types.SchedulingDecision, bool)
	SetSchedulingDecision(ctx sdk.Context, decision types.SchedulingDecision) error

	// Rewards
	DistributeJobRewards(ctx sdk.Context, jobID string) (*types.HPCRewardRecord, error)
	GetHPCReward(ctx sdk.Context, rewardID string) (types.HPCRewardRecord, bool)
	GetRewardsByJob(ctx sdk.Context, jobID string) []types.HPCRewardRecord
	SetHPCReward(ctx sdk.Context, reward types.HPCRewardRecord) error

	// Disputes
	FlagDispute(ctx sdk.Context, dispute *types.HPCDispute) error
	ResolveDispute(ctx sdk.Context, disputeID string, status types.DisputeStatus, resolution string, resolverAddr sdk.AccAddress) error
	GetDispute(ctx sdk.Context, disputeID string) (types.HPCDispute, bool)
	SetDispute(ctx sdk.Context, dispute types.HPCDispute) error

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithClusters(ctx sdk.Context, fn func(types.HPCCluster) bool)
	WithOfferings(ctx sdk.Context, fn func(types.HPCOffering) bool)
	WithJobs(ctx sdk.Context, fn func(types.HPCJob) bool)
	WithNodeMetadatas(ctx sdk.Context, fn func(types.NodeMetadata) bool)
	WithSchedulingDecisions(ctx sdk.Context, fn func(types.SchedulingDecision) bool)
	WithHPCRewards(ctx sdk.Context, fn func(types.HPCRewardRecord) bool)
	WithDisputes(ctx sdk.Context, fn func(types.HPCDispute) bool)

	// Routing enforcement (VE-5B)
	CreateRoutingAuditRecord(ctx sdk.Context, record *types.RoutingAuditRecord) error
	GetRoutingAuditRecord(ctx sdk.Context, recordID string) (types.RoutingAuditRecord, bool)
	GetRoutingAuditRecordsByJob(ctx sdk.Context, jobID string) []types.RoutingAuditRecord
	CreateRoutingViolation(ctx sdk.Context, violation *types.RoutingViolation) error
	GetRoutingViolation(ctx sdk.Context, violationID string) (types.RoutingViolation, bool)
	GetViolationsByProvider(ctx sdk.Context, providerAddr string) []types.RoutingViolation
	ResolveRoutingViolation(ctx sdk.Context, violationID string, resolution string) error
	ValidateJobRouting(ctx sdk.Context, job *types.HPCJob, targetClusterID string) (*types.RoutingAuditRecord, error)
	RefreshSchedulingDecision(ctx sdk.Context, job *types.HPCJob) (*types.SchedulingDecision, error)

	// Genesis sequence setters
	SetNextClusterSequence(ctx sdk.Context, seq uint64)
	SetNextOfferingSequence(ctx sdk.Context, seq uint64)
	SetNextJobSequence(ctx sdk.Context, seq uint64)
	SetNextDecisionSequence(ctx sdk.Context, seq uint64)
	SetNextDisputeSequence(ctx sdk.Context, seq uint64)

	// Block hooks
	ProcessExpiredJobs(ctx sdk.Context) error
	CheckClusterHealth(ctx sdk.Context) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// Keeper implements the HPC module keeper
type Keeper struct {
	skey          storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    BankKeeper
	billingKeeper BillingKeeper
	authority     string
}

// NewKeeper creates and returns an instance for HPC keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, bankKeeper BankKeeper, authority string) Keeper {
	return Keeper{
		cdc:           cdc,
		skey:          skey,
		bankKeeper:    bankKeeper,
		billingKeeper: nil,
		authority:     authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Sequence Management
// ============================================================================

func (k Keeper) getNextSequence(ctx sdk.Context, key []byte) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(key)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setNextSequence(ctx sdk.Context, key []byte, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(key, bz)
}

func (k Keeper) incrementSequence(ctx sdk.Context, key []byte) uint64 {
	seq := k.getNextSequence(ctx, key)
	k.setNextSequence(ctx, key, seq+1)
	return seq
}

// GetNextClusterSequence gets and increments the next cluster sequence
func (k Keeper) GetNextClusterSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyCluster)
}

// GetNextOfferingSequence gets and increments the next offering sequence
func (k Keeper) GetNextOfferingSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyOffering)
}

// GetNextJobSequence gets and increments the next job sequence
func (k Keeper) GetNextJobSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyJob)
}

// GetNextDecisionSequence gets and increments the next decision sequence
func (k Keeper) GetNextDecisionSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyDecision)
}

// GetNextDisputeSequence gets and increments the next dispute sequence
func (k Keeper) GetNextDisputeSequence(ctx sdk.Context) uint64 {
	return k.incrementSequence(ctx, types.SequenceKeyDispute)
}

// SetNextClusterSequence sets the next cluster sequence
func (k Keeper) SetNextClusterSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyCluster, seq)
}

// SetNextOfferingSequence sets the next offering sequence
func (k Keeper) SetNextOfferingSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyOffering, seq)
}

// SetNextJobSequence sets the next job sequence
func (k Keeper) SetNextJobSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyJob, seq)
}

// SetNextDecisionSequence sets the next decision sequence
func (k Keeper) SetNextDecisionSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyDecision, seq)
}

// SetNextDisputeSequence sets the next dispute sequence
func (k Keeper) SetNextDisputeSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.SequenceKeyDispute, seq)
}

// ============================================================================
// Parameters
// ============================================================================

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// ============================================================================
// Cluster Management
// ============================================================================

// RegisterCluster registers a new HPC cluster
func (k Keeper) RegisterCluster(ctx sdk.Context, cluster *types.HPCCluster) error {
	if err := cluster.Validate(); err != nil {
		return err
	}

	// Generate cluster ID if not set
	if cluster.ClusterID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyCluster)
		cluster.ClusterID = fmt.Sprintf("hpc-cluster-%d", seq)
	}

	// Check for duplicate
	if _, exists := k.GetCluster(ctx, cluster.ClusterID); exists {
		return types.ErrClusterAlreadyExists
	}

	cluster.CreatedAt = ctx.BlockTime()
	cluster.UpdatedAt = ctx.BlockTime()
	cluster.BlockHeight = ctx.BlockHeight()
	cluster.State = types.ClusterStatePending

	return k.SetCluster(ctx, *cluster)
}

// UpdateCluster updates an existing cluster
func (k Keeper) UpdateCluster(ctx sdk.Context, cluster *types.HPCCluster) error {
	existing, exists := k.GetCluster(ctx, cluster.ClusterID)
	if !exists {
		return types.ErrClusterNotFound
	}

	// Verify ownership
	if existing.ProviderAddress != cluster.ProviderAddress {
		return types.ErrUnauthorized
	}

	cluster.CreatedAt = existing.CreatedAt
	cluster.UpdatedAt = ctx.BlockTime()

	if err := cluster.Validate(); err != nil {
		return err
	}

	return k.SetCluster(ctx, *cluster)
}

// DeregisterCluster deregisters a cluster
func (k Keeper) DeregisterCluster(ctx sdk.Context, clusterID string, providerAddr sdk.AccAddress) error {
	cluster, exists := k.GetCluster(ctx, clusterID)
	if !exists {
		return types.ErrClusterNotFound
	}

	if cluster.ProviderAddress != providerAddr.String() {
		return types.ErrUnauthorized
	}

	cluster.State = types.ClusterStateDeregistered
	cluster.UpdatedAt = ctx.BlockTime()

	return k.SetCluster(ctx, cluster)
}

// GetCluster retrieves a cluster by ID
func (k Keeper) GetCluster(ctx sdk.Context, clusterID string) (types.HPCCluster, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetClusterKey(clusterID))
	if bz == nil {
		return types.HPCCluster{}, false
	}

	var cluster types.HPCCluster
	if err := json.Unmarshal(bz, &cluster); err != nil {
		return types.HPCCluster{}, false
	}
	return cluster, true
}

// GetClustersByProvider retrieves clusters by provider
func (k Keeper) GetClustersByProvider(ctx sdk.Context, providerAddr sdk.AccAddress) []types.HPCCluster {
	var clusters []types.HPCCluster
	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		if cluster.ProviderAddress == providerAddr.String() {
			clusters = append(clusters, cluster)
		}
		return false
	})
	return clusters
}

// SetCluster stores a cluster
func (k Keeper) SetCluster(ctx sdk.Context, cluster types.HPCCluster) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	store.Set(types.GetClusterKey(cluster.ClusterID), bz)
	return nil
}

// WithClusters iterates over all clusters
func (k Keeper) WithClusters(ctx sdk.Context, fn func(types.HPCCluster) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ClusterPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var cluster types.HPCCluster
		if err := json.Unmarshal(iter.Value(), &cluster); err != nil {
			continue
		}
		if fn(cluster) {
			break
		}
	}
}

// ============================================================================
// Offering Management
// ============================================================================

// CreateOffering creates a new HPC offering
func (k Keeper) CreateOffering(ctx sdk.Context, offering *types.HPCOffering) error {
	// Verify cluster exists
	cluster, exists := k.GetCluster(ctx, offering.ClusterID)
	if !exists {
		return types.ErrClusterNotFound
	}

	// Verify ownership
	if cluster.ProviderAddress != offering.ProviderAddress {
		return types.ErrUnauthorized
	}

	// Generate offering ID if not set
	if offering.OfferingID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyOffering)
		offering.OfferingID = fmt.Sprintf("hpc-offering-%d", seq)
	}

	offering.Active = true
	offering.CreatedAt = ctx.BlockTime()
	offering.UpdatedAt = ctx.BlockTime()
	offering.BlockHeight = ctx.BlockHeight()

	if err := offering.Validate(); err != nil {
		return err
	}

	return k.SetOffering(ctx, *offering)
}

// UpdateOffering updates an existing offering
func (k Keeper) UpdateOffering(ctx sdk.Context, offering *types.HPCOffering) error {
	existing, exists := k.GetOffering(ctx, offering.OfferingID)
	if !exists {
		return types.ErrOfferingNotFound
	}

	if existing.ProviderAddress != offering.ProviderAddress {
		return types.ErrUnauthorized
	}

	offering.CreatedAt = existing.CreatedAt
	offering.UpdatedAt = ctx.BlockTime()

	if err := offering.Validate(); err != nil {
		return err
	}

	return k.SetOffering(ctx, *offering)
}

// GetOffering retrieves an offering by ID
func (k Keeper) GetOffering(ctx sdk.Context, offeringID string) (types.HPCOffering, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetOfferingKey(offeringID))
	if bz == nil {
		return types.HPCOffering{}, false
	}

	var offering types.HPCOffering
	if err := json.Unmarshal(bz, &offering); err != nil {
		return types.HPCOffering{}, false
	}
	return offering, true
}

// GetOfferingsByCluster retrieves offerings by cluster
func (k Keeper) GetOfferingsByCluster(ctx sdk.Context, clusterID string) []types.HPCOffering {
	var offerings []types.HPCOffering
	k.WithOfferings(ctx, func(offering types.HPCOffering) bool {
		if offering.ClusterID == clusterID {
			offerings = append(offerings, offering)
		}
		return false
	})
	return offerings
}

// SetOffering stores an offering
func (k Keeper) SetOffering(ctx sdk.Context, offering types.HPCOffering) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(offering)
	if err != nil {
		return err
	}
	store.Set(types.GetOfferingKey(offering.OfferingID), bz)
	return nil
}

// WithOfferings iterates over all offerings
func (k Keeper) WithOfferings(ctx sdk.Context, fn func(types.HPCOffering) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.OfferingPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var offering types.HPCOffering
		if err := json.Unmarshal(iter.Value(), &offering); err != nil {
			continue
		}
		if fn(offering) {
			break
		}
	}
}

// ============================================================================
// Job Management
// ============================================================================

// SubmitJob submits a new HPC job
func (k Keeper) SubmitJob(ctx sdk.Context, job *types.HPCJob) error {
	// Verify offering exists and is active
	offering, exists := k.GetOffering(ctx, job.OfferingID)
	if !exists {
		return types.ErrOfferingNotFound
	}
	if !offering.Active {
		return types.ErrOfferingNotFound.Wrap("offering is not active")
	}

	// Verify cluster exists and is available
	cluster, exists := k.GetCluster(ctx, offering.ClusterID)
	if !exists {
		return types.ErrClusterNotFound
	}
	if cluster.State != types.ClusterStateActive {
		return types.ErrClusterNotFound.Wrap("cluster is not active")
	}

	// Validate max runtime
	if job.MaxRuntimeSeconds > offering.MaxRuntimeSeconds {
		return types.ErrMaxRuntimeExceeded
	}

	// Generate job ID if not set
	if job.JobID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyJob)
		job.JobID = fmt.Sprintf("hpc-job-%d", seq)
	}

	job.ClusterID = offering.ClusterID
	job.ProviderAddress = offering.ProviderAddress
	job.State = types.JobStatePending
	job.CreatedAt = ctx.BlockTime()
	job.BlockHeight = ctx.BlockHeight()

	if err := job.Validate(); err != nil {
		return err
	}

	return k.SetJob(ctx, *job)
}

// UpdateJobStatus updates job status
func (k Keeper) UpdateJobStatus(ctx sdk.Context, jobID string, state types.JobState, statusMessage string, exitCode int32, metrics *types.HPCUsageMetrics) error {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return types.ErrJobNotFound
	}

	// Validate state transition
	if !isValidJobStateTransition(job.State, state) {
		return types.ErrInvalidJobState
	}

	job.State = state
	job.StatusMessage = statusMessage
	job.ExitCode = exitCode

	now := ctx.BlockTime()
	switch state {
	case types.JobStateQueued:
		job.QueuedAt = &now
	case types.JobStateRunning:
		job.StartedAt = &now
	case types.JobStateCompleted, types.JobStateFailed, types.JobStateCancelled, types.JobStateTimeout:
		job.CompletedAt = &now
	}

	return k.SetJob(ctx, job)
}

// CancelJob cancels a job
func (k Keeper) CancelJob(ctx sdk.Context, jobID string, requesterAddr sdk.AccAddress) error {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return types.ErrJobNotFound
	}

	// Only customer or provider can cancel
	if job.CustomerAddress != requesterAddr.String() && job.ProviderAddress != requesterAddr.String() {
		return types.ErrUnauthorized
	}

	// Cannot cancel terminal jobs
	if types.IsTerminalJobState(job.State) {
		return types.ErrInvalidJobState.Wrap("job is already in terminal state")
	}

	now := ctx.BlockTime()
	job.State = types.JobStateCancelled
	job.StatusMessage = "Cancelled by user"
	job.CompletedAt = &now

	return k.SetJob(ctx, job)
}

// GetJob retrieves a job by ID
func (k Keeper) GetJob(ctx sdk.Context, jobID string) (types.HPCJob, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetJobKey(jobID))
	if bz == nil {
		return types.HPCJob{}, false
	}

	var job types.HPCJob
	if err := json.Unmarshal(bz, &job); err != nil {
		return types.HPCJob{}, false
	}
	return job, true
}

// GetJobsByCustomer retrieves jobs by customer
func (k Keeper) GetJobsByCustomer(ctx sdk.Context, customerAddr sdk.AccAddress) []types.HPCJob {
	var jobs []types.HPCJob
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if job.CustomerAddress == customerAddr.String() {
			jobs = append(jobs, job)
		}
		return false
	})
	return jobs
}

// GetJobsByCluster retrieves jobs by cluster
func (k Keeper) GetJobsByCluster(ctx sdk.Context, clusterID string) []types.HPCJob {
	var jobs []types.HPCJob
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if job.ClusterID == clusterID {
			jobs = append(jobs, job)
		}
		return false
	})
	return jobs
}

// SetJob stores a job
func (k Keeper) SetJob(ctx sdk.Context, job types.HPCJob) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(job)
	if err != nil {
		return err
	}
	store.Set(types.GetJobKey(job.JobID), bz)
	return nil
}

// WithJobs iterates over all jobs
func (k Keeper) WithJobs(ctx sdk.Context, fn func(types.HPCJob) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.JobPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var job types.HPCJob
		if err := json.Unmarshal(iter.Value(), &job); err != nil {
			continue
		}
		if fn(job) {
			break
		}
	}
}

// isValidJobStateTransition checks if a job state transition is valid
func isValidJobStateTransition(from, to types.JobState) bool {
	validTransitions := map[types.JobState][]types.JobState{
		types.JobStatePending:   {types.JobStateQueued, types.JobStateFailed, types.JobStateCancelled},
		types.JobStateQueued:    {types.JobStateRunning, types.JobStateFailed, types.JobStateCancelled, types.JobStateTimeout},
		types.JobStateRunning:   {types.JobStateCompleted, types.JobStateFailed, types.JobStateCancelled, types.JobStateTimeout},
		types.JobStateCompleted: {},
		types.JobStateFailed:    {},
		types.JobStateCancelled: {},
		types.JobStateTimeout:   {},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ============================================================================
// Block Hooks
// ============================================================================

// ProcessExpiredJobs processes expired jobs
func (k Keeper) ProcessExpiredJobs(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	maxDuration := time.Duration(params.MaxJobDurationSeconds) * time.Second

	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if types.IsTerminalJobState(job.State) {
			return false
		}

		// Check if job has exceeded max runtime
		if job.StartedAt != nil {
			runtime := ctx.BlockTime().Sub(*job.StartedAt)
			if runtime > maxDuration || runtime > time.Duration(job.MaxRuntimeSeconds)*time.Second {
				now := ctx.BlockTime()
				job.State = types.JobStateTimeout
				job.StatusMessage = "Job exceeded maximum runtime"
				job.CompletedAt = &now
				if err := k.SetJob(ctx, job); err != nil {
					k.Logger(ctx).Error("failed to set job", "error", err)
				}
			}
		}

		return false
	})

	return nil
}

// CheckClusterHealth checks cluster health based on heartbeats
func (k Keeper) CheckClusterHealth(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	timeout := time.Duration(params.ClusterHeartbeatTimeout) * time.Second

	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		if cluster.State == types.ClusterStateDeregistered {
			return false
		}

		// Check if cluster has timed out
		if ctx.BlockTime().Sub(cluster.UpdatedAt) > timeout {
			if cluster.State == types.ClusterStateActive {
				cluster.State = types.ClusterStateOffline
				if err := k.SetCluster(ctx, cluster); err != nil {
					k.Logger(ctx).Error("failed to set cluster", "error", err)
				}
			}
		}

		return false
	})

	// Check for stale nodes and update cluster capacities
	if err := k.CheckStaleNodes(ctx); err != nil {
		k.Logger(ctx).Error("failed to check stale nodes", "error", err)
	}

	// Update cluster capacities based on active nodes
	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		if cluster.State != types.ClusterStateDeregistered {
			if err := k.UpdateClusterCapacity(ctx, cluster.ClusterID); err != nil {
				k.Logger(ctx).Error("failed to update cluster capacity",
					"cluster_id", cluster.ClusterID, "error", err)
			}
		}
		return false
	})

	return nil
}
