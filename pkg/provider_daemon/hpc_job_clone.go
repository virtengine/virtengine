// Package provider_daemon implements the VirtEngine provider daemon.
package provider_daemon

func cloneSchedulerMetrics(metrics *HPCSchedulerMetrics) *HPCSchedulerMetrics {
	if metrics == nil {
		return nil
	}

	clone := *metrics
	if metrics.SchedulerSpecific != nil {
		clone.SchedulerSpecific = make(map[string]interface{}, len(metrics.SchedulerSpecific))
		for key, value := range metrics.SchedulerSpecific {
			clone.SchedulerSpecific[key] = value
		}
	}
	return &clone
}

func cloneSchedulerJob(job *HPCSchedulerJob) *HPCSchedulerJob {
	if job == nil {
		return nil
	}

	clone := *job
	if job.NodeList != nil {
		clone.NodeList = append([]string(nil), job.NodeList...)
	}
	if job.StartTime != nil {
		start := *job.StartTime
		clone.StartTime = &start
	}
	if job.EndTime != nil {
		end := *job.EndTime
		clone.EndTime = &end
	}
	clone.Metrics = cloneSchedulerMetrics(job.Metrics)
	if job.OriginalJob != nil {
		orig := *job.OriginalJob
		clone.OriginalJob = &orig
	}
	return &clone
}
