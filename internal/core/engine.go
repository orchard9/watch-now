package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/orchard9/watch-now/internal/config"
	"github.com/orchard9/watch-now/internal/monitors"
)

type Engine struct {
	config    *config.Config
	monitors  []monitors.Monitor
	state     *StateStore
	scheduler *Scheduler
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		config: cfg,
		state:  NewStateStore(),
	}
}

func (e *Engine) Initialize() error {
	// Create service monitors
	for _, serviceCfg := range e.config.Services {
		switch serviceCfg.Type {
		case "rest":
			monitor := monitors.NewRESTMonitor(serviceCfg)
			e.monitors = append(e.monitors, monitor)
		case "grpc":
			// TODO: Implement gRPC monitor
			fmt.Printf("Warning: gRPC monitor not yet implemented for %s\n", serviceCfg.Name)
		default:
			fmt.Printf("Warning: unknown service type %s for %s\n", serviceCfg.Type, serviceCfg.Name)
		}
	}

	// Create quality monitors from checks
	for _, checkCfg := range e.config.Checks {
		monitor := monitors.NewQualityMonitor(checkCfg)
		e.monitors = append(e.monitors, monitor)
	}

	// Create scheduler
	e.scheduler = NewScheduler(e.config.Interval, e.monitors, e.state)

	return nil
}

func (e *Engine) Start(ctx context.Context) error {
	// Start scheduler
	return e.scheduler.Start(ctx)
}

func (e *Engine) State() *StateStore {
	return e.state
}

func (e *Engine) MonitorCount() int {
	return len(e.monitors)
}

type Scheduler struct {
	interval time.Duration
	monitors []monitors.Monitor
	state    *StateStore
}

func NewScheduler(interval time.Duration, monitors []monitors.Monitor, state *StateStore) *Scheduler {
	return &Scheduler{
		interval: interval,
		monitors: monitors,
		state:    state,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	// Run initial check
	s.runChecks(ctx)

	// Set up ticker for periodic checks
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.runChecks(ctx)
		}
	}
}

func (s *Scheduler) runChecks(ctx context.Context) {
	var wg sync.WaitGroup

	// Run all monitors concurrently
	for _, monitor := range s.monitors {
		wg.Add(1)
		go func(m monitors.Monitor) {
			defer wg.Done()

			result, err := m.Check(ctx)
			if err != nil {
				// Create error result
				result = &monitors.Result{
					Name:      m.Name(),
					Type:      m.Type(),
					Status:    monitors.StatusFail,
					Message:   fmt.Sprintf("Monitor error: %v", err),
					Timestamp: time.Now(),
				}
			}

			// Update state
			s.state.Update(result)
		}(monitor)
	}

	wg.Wait()
}
