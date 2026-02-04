// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Chain Subscriber - subscribes to on-chain HPC job events
package provider_daemon

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCEventType represents the type of HPC chain event
type HPCEventType string

const (
	// HPCEventTypeJobCreated is emitted when a new HPC job is created
	HPCEventTypeJobCreated HPCEventType = "hpc_job_created"

	// HPCEventTypeJobCancelled is emitted when an HPC job is cancelled
	HPCEventTypeJobCancelled HPCEventType = "hpc_job_cancelled"

	// HPCEventTypeJobUpdated is emitted when a job state changes
	HPCEventTypeJobUpdated HPCEventType = "hpc_job_updated"
)

// HPCChainEvent represents an HPC event from the chain
type HPCChainEvent struct {
	Type        HPCEventType     `json:"type"`
	JobID       string           `json:"job_id"`
	ClusterID   string           `json:"cluster_id"`
	BlockHeight int64            `json:"block_height"`
	Timestamp   time.Time        `json:"timestamp"`
	Job         *hpctypes.HPCJob `json:"job,omitempty"`
}

// SubscriberStats contains statistics for the subscriber
type SubscriberStats struct {
	JobsReceived      int64     `json:"jobs_received"`
	JobsProcessed     int64     `json:"jobs_processed"`
	CancelsReceived   int64     `json:"cancels_received"`
	CancelsProcessed  int64     `json:"cancels_processed"`
	ProcessingErrors  int64     `json:"processing_errors"`
	ReconnectCount    int64     `json:"reconnect_count"`
	LastEventTime     time.Time `json:"last_event_time"`
	LastErrorTime     time.Time `json:"last_error_time,omitempty"`
	LastError         string    `json:"last_error,omitempty"`
	StartTime         time.Time `json:"start_time"`
	Uptime            string    `json:"uptime"`
}

// HPCChainSubscriberWithStats extends HPCChainSubscriber with statistics tracking
// and enhanced error recovery capabilities
type HPCChainSubscriberWithStats struct {
	config        HPCChainSubscriberConfig
	clusterID     string
	providerAddr  string
	chainClient   HPCChainClient
	jobService    *HPCJobService
	healthStatus  HPCComponentHealth

	mu         sync.RWMutex
	running    bool
	connected  bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc

	// Statistics (use atomic for thread-safety)
	stats struct {
		jobsReceived     atomic.Int64
		jobsProcessed    atomic.Int64
		cancelsReceived  atomic.Int64
		cancelsProcessed atomic.Int64
		errors           atomic.Int64
		reconnectCount   atomic.Int64
		lastEventTime    atomic.Value // time.Time
		lastErrorTime    atomic.Value // time.Time
		lastError        atomic.Value // string
		startTime        time.Time
	}

	// Event channels
	jobEventCh    chan *hpctypes.HPCJob
	cancelEventCh chan string
}

// NewHPCChainSubscriberWithStats creates a new HPC chain subscriber with statistics tracking
func NewHPCChainSubscriberWithStats(
	config HPCChainSubscriberConfig,
	clusterID string,
	providerAddr string,
	chainClient HPCChainClient,
	jobService *HPCJobService,
) (*HPCChainSubscriberWithStats, error) {
	if jobService == nil {
		return nil, fmt.Errorf("job service is required")
	}

	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required")
	}

	bufferSize := config.SubscriptionBufferSize
	if bufferSize <= 0 {
		bufferSize = 100
	}

	sub := &HPCChainSubscriberWithStats{
		config:        config,
		clusterID:     clusterID,
		providerAddr:  providerAddr,
		chainClient:   chainClient,
		jobService:    jobService,
		stopCh:        make(chan struct{}),
		jobEventCh:    make(chan *hpctypes.HPCJob, bufferSize),
		cancelEventCh: make(chan string, bufferSize),
		healthStatus: HPCComponentHealth{
			Name:    "chain_subscriber_stats",
			Healthy: false,
			Message: "not started",
		},
	}

	sub.stats.lastEventTime.Store(time.Time{})
	sub.stats.lastErrorTime.Store(time.Time{})
	sub.stats.lastError.Store("")

	return sub, nil
}

// Start starts the HPC chain subscriber
func (s *HPCChainSubscriberWithStats) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.stats.startTime = time.Now()
	s.mu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// Start event processing workers
	s.wg.Add(2)
	go s.processJobEvents(ctx)
	go s.processCancelEvents(ctx)

	// Start subscription with reconnection
	s.wg.Add(1)
	go s.subscriptionLoop(ctx)

	s.mu.Lock()
	s.healthStatus = HPCComponentHealth{
		Name:      "chain_subscriber_stats",
		Healthy:   true,
		Message:   "running",
		LastCheck: time.Now(),
	}
	s.mu.Unlock()

	return nil
}

// Stop stops the HPC chain subscriber
func (s *HPCChainSubscriberWithStats) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.stopCh)

	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.mu.Unlock()

	// Wait for goroutines to finish
	s.wg.Wait()

	// Drain remaining events
	close(s.jobEventCh)
	close(s.cancelEventCh)

	s.mu.Lock()
	s.healthStatus = HPCComponentHealth{
		Name:      "chain_subscriber_stats",
		Healthy:   false,
		Message:   "stopped",
		LastCheck: time.Now(),
	}
	s.mu.Unlock()

	return nil
}

// IsRunning returns true if the subscriber is running
func (s *HPCChainSubscriberWithStats) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// IsConnected returns true if the subscriber is connected to the chain
func (s *HPCChainSubscriberWithStats) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// GetStats returns the subscriber statistics
func (s *HPCChainSubscriberWithStats) GetStats() *SubscriberStats {
	lastEventTime, _ := s.stats.lastEventTime.Load().(time.Time)
	lastErrorTime, _ := s.stats.lastErrorTime.Load().(time.Time)
	lastError, _ := s.stats.lastError.Load().(string)

	stats := &SubscriberStats{
		JobsReceived:     s.stats.jobsReceived.Load(),
		JobsProcessed:    s.stats.jobsProcessed.Load(),
		CancelsReceived:  s.stats.cancelsReceived.Load(),
		CancelsProcessed: s.stats.cancelsProcessed.Load(),
		ProcessingErrors: s.stats.errors.Load(),
		ReconnectCount:   s.stats.reconnectCount.Load(),
		LastEventTime:    lastEventTime,
		LastErrorTime:    lastErrorTime,
		LastError:        lastError,
		StartTime:        s.stats.startTime,
	}

	if !s.stats.startTime.IsZero() {
		stats.Uptime = time.Since(s.stats.startTime).Truncate(time.Second).String()
	}

	return stats
}

// GetHealth returns the health status
func (s *HPCChainSubscriberWithStats) GetHealth() HPCComponentHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health := s.healthStatus
	health.LastCheck = time.Now()

	stats := s.GetStats()
	health.Details = map[string]interface{}{
		"jobs_received":     stats.JobsReceived,
		"jobs_processed":    stats.JobsProcessed,
		"cancels_received":  stats.CancelsReceived,
		"cancels_processed": stats.CancelsProcessed,
		"processing_errors": stats.ProcessingErrors,
		"reconnect_count":   stats.ReconnectCount,
		"connected":         s.connected,
	}

	return health
}

// subscriptionLoop manages the chain subscription with reconnection
func (s *HPCChainSubscriberWithStats) subscriptionLoop(ctx context.Context) {
	defer s.wg.Done()

	reconnectDelay := s.config.ReconnectInterval
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
		}

		err := s.subscribe(ctx)
		if err != nil {
			s.recordError(err)
			s.stats.reconnectCount.Add(1)
			attempts++

			// Check max attempts
			if s.config.MaxReconnectAttempts > 0 && attempts >= s.config.MaxReconnectAttempts {
				s.mu.Lock()
				s.healthStatus.Healthy = false
				s.healthStatus.Message = fmt.Sprintf("max reconnect attempts (%d) reached", s.config.MaxReconnectAttempts)
				s.mu.Unlock()
				return
			}

			// Wait before reconnecting with exponential backoff
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-time.After(reconnectDelay):
			}

			// Increase backoff (double it up to 10 minutes max)
			reconnectDelay = time.Duration(float64(reconnectDelay) * 2)
			if reconnectDelay > 10*time.Minute {
				reconnectDelay = 10 * time.Minute
			}
		} else {
			// Reset on successful connection
			reconnectDelay = s.config.ReconnectInterval
			attempts = 0
		}
	}
}

// subscribe establishes the subscription to chain events
func (s *HPCChainSubscriberWithStats) subscribe(ctx context.Context) error {
	if s.chainClient == nil {
		return fmt.Errorf("chain client not configured")
	}

	s.setConnected(true)
	defer s.setConnected(false)

	errCh := make(chan error, 2)

	// Subscribe to job requests
	go func() {
		err := s.chainClient.SubscribeToJobRequests(ctx, s.clusterID, func(job *hpctypes.HPCJob) error {
			s.stats.jobsReceived.Add(1)
			s.stats.lastEventTime.Store(time.Now())

			select {
			case s.jobEventCh <- job:
			default:
				// Buffer full, record error
				s.recordError(fmt.Errorf("job event buffer full, dropping event for job %s", job.JobID))
			}
			return nil
		})
		if err != nil {
			errCh <- fmt.Errorf("job subscription error: %w", err)
		}
	}()

	// Subscribe to job cancellations
	go func() {
		err := s.chainClient.SubscribeToJobCancellations(ctx, s.clusterID, func(jobID string) error {
			s.stats.cancelsReceived.Add(1)
			s.stats.lastEventTime.Store(time.Now())

			select {
			case s.cancelEventCh <- jobID:
			default:
				// Buffer full, record error
				s.recordError(fmt.Errorf("cancel event buffer full, dropping event for job %s", jobID))
			}
			return nil
		})
		if err != nil {
			errCh <- fmt.Errorf("cancel subscription error: %w", err)
		}
	}()

	// Wait for error or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// processJobEvents processes incoming job creation events
func (s *HPCChainSubscriberWithStats) processJobEvents(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case job, ok := <-s.jobEventCh:
			if !ok {
				return
			}

			// Filter by provider address if configured
			if s.providerAddr != "" && job.ProviderAddress != s.providerAddr {
				continue
			}

			// Process the job request
			if err := s.jobService.HandleJobRequest(job); err != nil {
				s.recordError(fmt.Errorf("failed to handle job request %s: %w", job.JobID, err))
			} else {
				s.stats.jobsProcessed.Add(1)
			}
		}
	}
}

// processCancelEvents processes incoming job cancellation events
func (s *HPCChainSubscriberWithStats) processCancelEvents(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case jobID, ok := <-s.cancelEventCh:
			if !ok {
				return
			}

			// Process the cancellation
			if err := s.jobService.HandleJobCancellation(jobID); err != nil {
				s.recordError(fmt.Errorf("failed to handle job cancellation %s: %w", jobID, err))
			} else {
				s.stats.cancelsProcessed.Add(1)
			}
		}
	}
}

// setConnected sets the connected state
func (s *HPCChainSubscriberWithStats) setConnected(connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connected = connected

	if connected {
		s.healthStatus.Healthy = true
		s.healthStatus.Message = "connected"
	} else {
		s.healthStatus.Message = "disconnected"
	}
	s.healthStatus.LastCheck = time.Now()
}

// recordError records an error in the stats
func (s *HPCChainSubscriberWithStats) recordError(err error) {
	s.stats.errors.Add(1)
	s.stats.lastErrorTime.Store(time.Now())
	s.stats.lastError.Store(err.Error())

	s.mu.Lock()
	s.healthStatus.Details = map[string]interface{}{
		"last_error": err.Error(),
	}
	s.mu.Unlock()
}

// InjectJobEvent injects a job event for testing
func (s *HPCChainSubscriberWithStats) InjectJobEvent(job *hpctypes.HPCJob) error {
	if !s.IsRunning() {
		return fmt.Errorf("subscriber not running")
	}

	s.stats.jobsReceived.Add(1)
	s.stats.lastEventTime.Store(time.Now())

	select {
	case s.jobEventCh <- job:
		return nil
	default:
		return fmt.Errorf("event buffer full")
	}
}

// InjectCancelEvent injects a cancellation event for testing
func (s *HPCChainSubscriberWithStats) InjectCancelEvent(jobID string) error {
	if !s.IsRunning() {
		return fmt.Errorf("subscriber not running")
	}

	s.stats.cancelsReceived.Add(1)
	s.stats.lastEventTime.Store(time.Now())

	select {
	case s.cancelEventCh <- jobID:
		return nil
	default:
		return fmt.Errorf("event buffer full")
	}
}
