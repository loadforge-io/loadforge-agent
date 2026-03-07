package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"loadforge-agent/internal/scenario"
)

type Runner struct {
	scenario *scenario.Scenario
}

func New(sc *scenario.Scenario) *Runner {
	return &Runner{scenario: sc}
}

func (r *Runner) Run(ctx context.Context) error {
	sc := r.scenario
	if sc == nil {
		return fmt.Errorf("scenario is nil")
	}

	duration := time.Duration(sc.Duration) * time.Second
	numVUs := int(sc.VirtualUsers)

	rampDuration := duration / 10
	if rampDuration < time.Millisecond {
		rampDuration = 0
	}
	var rampInterval time.Duration
	if numVUs > 1 && rampDuration > 0 {
		rampInterval = rampDuration / time.Duration(numVUs)
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	var wg sync.WaitGroup

	for i := range numVUs {
		if i > 0 && rampInterval > 0 {
			select {
			case <-ctx.Done():
				break
			case <-time.After(rampInterval):
			}
		}

		if ctx.Err() != nil {
			break
		}

		vu, err := newVirtualUser(uint64(i+1), sc)
		if err != nil {
			return fmt.Errorf("failed to create vu %d: %w", i+1, err)
		}

		wg.Add(1)
		go func(vu *VirtualUser) {
			defer wg.Done()
			runVU(ctx, vu)
		}(vu)
	}

	wg.Wait()
	signal.Stop(sigCh)
	return nil
}

func runVU(ctx context.Context, vu *VirtualUser) {
	for {
		if ctx.Err() != nil {
			return
		}
		_ = vu.runScenario(ctx)
	}
}
