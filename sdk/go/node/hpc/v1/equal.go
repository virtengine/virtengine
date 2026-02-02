// Package v1 provides equality methods for generated HPC types.
package v1

// Equal compares two Params for equality
func (p *Params) Equal(other *Params) bool {
	if p == nil && other == nil {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	return p.MinDeposit == other.MinDeposit &&
		p.MaxJobDuration == other.MaxJobDuration &&
		p.PlatformFeeRate == other.PlatformFeeRate &&
		p.DisputeResolutionPeriod == other.DisputeResolutionPeriod
}

// Equal compares two Clusters for equality
func (c *Cluster) Equal(other *Cluster) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	return c.ClusterId == other.ClusterId &&
		c.Owner == other.Owner &&
		c.Name == other.Name &&
		c.ClusterType == other.ClusterType &&
		c.Region == other.Region &&
		c.Endpoint == other.Endpoint &&
		c.TotalNodes == other.TotalNodes &&
		c.TotalGpus == other.TotalGpus &&
		c.Active == other.Active &&
		c.RegisteredAt == other.RegisteredAt
}

// Equal compares two Offerings for equality
func (o *Offering) Equal(other *Offering) bool {
	if o == nil && other == nil {
		return true
	}
	if o == nil || other == nil {
		return false
	}
	return o.OfferingId == other.OfferingId &&
		o.ClusterId == other.ClusterId &&
		o.Provider == other.Provider &&
		o.Name == other.Name &&
		o.ResourceType == other.ResourceType &&
		o.PricePerHour == other.PricePerHour &&
		o.MinDuration == other.MinDuration &&
		o.MaxDuration == other.MaxDuration &&
		o.Active == other.Active
}

// Equal compares two Jobs for equality
func (j *Job) Equal(other *Job) bool {
	if j == nil && other == nil {
		return true
	}
	if j == nil || other == nil {
		return false
	}
	return j.JobId == other.JobId &&
		j.OfferingId == other.OfferingId &&
		j.Submitter == other.Submitter &&
		j.Status == other.Status &&
		j.SubmittedAt == other.SubmittedAt &&
		j.StartedAt == other.StartedAt &&
		j.CompletedAt == other.CompletedAt
}
