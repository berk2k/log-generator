package application

import (
	"fmt"
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
	"sync"
	"time"
)

type Worker struct {
	ID       int
	InChan   <-chan domain.LogMessage //recieve-only
	Logger   *infrastructure.ConsoleLogger
	Metrics  *Metrics
	Wg       *sync.WaitGroup
	StopChan chan struct{}
}

func NewWorker(id int, in <-chan domain.LogMessage, logger *infrastructure.ConsoleLogger, metrics *Metrics, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:       id,
		InChan:   in,
		Logger:   logger,
		Metrics:  metrics,
		Wg:       wg,
		StopChan: make(chan struct{}),
	}
}

func (w *Worker) Start(batchSize int, batchTimeout time.Duration) {
	go func() {
		defer w.Wg.Done()

		// Preallocate batch slice to avoid reallocation overhead
		batch := make([]domain.LogMessage, 0, batchSize)

		// Timeout ticker for flushing incomplete batches
		ticker := time.NewTicker(batchTimeout)
		defer ticker.Stop()

	Loop:
		for {
			select {

			// New log message received
			case msg, ok := <-w.InChan:
				if !ok {
					// Input channel is closed → shutdown signal for worker
					break Loop
				}

				// Add the log to the batch
				batch = append(batch, msg)

				// Batch is full → flush immediately
				if len(batch) >= batchSize {
					w.processBatch(batch)
					batch = batch[:0] // Reset slice without reallocating

				}

			// Timeout reached → flush partial batch
			case <-ticker.C:
				if len(batch) > 0 {
					w.processBatch(batch)
					batch = batch[:0]
				}

			case <-w.StopChan:
				// Graceful stop for this worker only
				break Loop
			}
		}

		// Worker is exiting → flush remaining logs
		if len(batch) > 0 {
			w.processBatch(batch)
		}
	}()
}

// Handles the actual batch processing logic (e.g., file write, DB insert, network send)
func (w *Worker) processBatch(batch []domain.LogMessage) {
	// Simulate slow IO (disk / network / external system)
	time.Sleep(300 * time.Millisecond)

	fmt.Printf("Worker %d flushed batch of %d logs\n", w.ID, len(batch))

	// Count processed messages in metrics

	if w.Metrics != nil {
		w.Metrics.IncProcessedBy(len(batch))
	}
}
