package application

import (
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
	"sync"
	"time"
)

type WorkerPool struct {
	Workers      []*Worker
	NextID       int
	InChan       <-chan domain.LogMessage
	Logger       *infrastructure.ConsoleLogger
	Metrics      *Metrics
	Wg           *sync.WaitGroup
	Mu           sync.Mutex
	BatchSize    int
	BatchTimeout time.Duration
}

func NewWorkerPool(
	in <-chan domain.LogMessage,
	logger *infrastructure.ConsoleLogger,
	metrics *Metrics,
	wg *sync.WaitGroup,
	batchSize int,
	batchTimeout time.Duration,
) *WorkerPool {

	return &WorkerPool{
		Workers:      []*Worker{},
		NextID:       1,
		InChan:       in,
		Logger:       logger,
		Metrics:      metrics,
		Wg:           wg,
		BatchSize:    batchSize,
		BatchTimeout: batchTimeout,
	}
}

func (p *WorkerPool) AddWorker() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	worker := NewWorker(
		p.NextID,
		p.InChan,
		p.Logger,
		p.Metrics,
		p.Wg,
	)

	p.Workers = append(p.Workers, worker)
	p.Wg.Add(1)

	worker.Start(p.BatchSize, p.BatchTimeout)

	p.NextID++

	if p.Metrics != nil {
		p.Metrics.SetWorkers(len(p.Workers))
	}

	println("worker added:", worker.ID)
}

func (p *WorkerPool) RemoveWorker() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if len(p.Workers) == 0 {
		return
	}

	// stop last worker
	worker := p.Workers[len(p.Workers)-1]
	close(worker.StopChan) // stop signal

	//remove from slice
	p.Workers = p.Workers[:len(p.Workers)-1]

	if p.Metrics != nil {
		p.Metrics.SetWorkers(len(p.Workers))
	}
	println("Worker removed:", worker.ID)
}

func (p *WorkerPool) WorkerCount() int {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	return len(p.Workers)
}
