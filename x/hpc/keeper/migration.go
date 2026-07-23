package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// MigrateReservationLinks quarantines no state itself; legacy executable jobs
// without reservation lineage are counted so resources migration/operator
// reconciliation can resolve them without synthetic capacity.
func (k Keeper) MigrateReservationLinks(ctx sdk.Context) (jobs, linked, executableOrphans, terminalPreserved uint64) {
	values := make([]types.HPCJob, 0)
	k.WithJobs(ctx, func(job types.HPCJob) bool { values = append(values, job); return false })
	sort.Slice(values, func(i, j int) bool { return values[i].JobID < values[j].JobID })
	for _, job := range values {
		jobs++
		if job.ReservationID != "" {
			linked++
			continue
		}
		if types.IsTerminalJobState(job.State) {
			terminalPreserved++
			continue
		}
		if job.State == types.JobStateQueued || job.State == types.JobStateRunning {
			executableOrphans++
		}
	}
	return
}

// SetLegacyReservationLink records migration quarantine lineage without
// otherwise changing the legacy job lifecycle state.
func (k Keeper) SetLegacyReservationLink(ctx sdk.Context, jobID, reservationID string) error {
	job, found := k.GetJob(ctx, jobID)
	if !found {
		return types.ErrJobNotFound
	}
	job.ReservationID = reservationID
	job.AllocationID = reservationID
	return k.SetJob(ctx, job)
}
