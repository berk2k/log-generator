package application

import (
	"context"
	"log-generator/internal/domain"
	"time"
)

type AutoScaler struct {
	Pool               *WorkerPool
	InChan             chan domain.LogMessage
	Interval           time.Duration
	MinWorkers         int
	MaxWorkers         int
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
	Ctx                context.Context
}

func NewAutoScaler(
	pool *WorkerPool,
	in chan domain.LogMessage,
	interval time.Duration,
	min int,
	max int,
	ctx context.Context,
) *AutoScaler {
	return &AutoScaler{
		Pool:               pool,
		InChan:             in,
		Interval:           interval,
		MinWorkers:         min,
		MaxWorkers:         max,
		ScaleUpThreshold:   0.8, // %80 full
		ScaleDownThreshold: 0.2, // %20 empty
		Ctx:                ctx,
	}
}

func (a *AutoScaler) Start() {
	go func() {
		ticker := time.NewTicker(a.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-a.Ctx.Done():
				// shutdown -> stop autoscaler
				return

			case <-ticker.C:
				a.check()
			}
		}
	}()
}

func (a *AutoScaler) check() {
	queueSize := len(a.InChan)
	queueCap := cap(a.InChan)

	usage := float64(queueSize) / float64(queueCap)

	workers := a.Pool.WorkerCount()

	//scale up
	if usage > a.ScaleUpThreshold && workers < a.MaxWorkers {
		a.Pool.AddWorker()
		println("Scaler: scalling UP, workers = ", workers+1)
		return
	}

	//scale down
	if usage < a.ScaleDownThreshold && workers > a.MinWorkers {
		a.Pool.RemoveWorker()
		println("Scaler: scalling Down, workers = ", workers-1)
		return
	}
}
