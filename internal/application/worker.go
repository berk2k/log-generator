package application

import (
	"fmt"
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
	"sync"
)

type Worker struct {
	ID      int
	InChan  <-chan domain.LogMessage //recieve-only
	Logger  *infrastructure.ConsoleLogger
	Metrics *Metrics
	Wg      *sync.WaitGroup
}

func NewWorker(id int, in <-chan domain.LogMessage, logger *infrastructure.ConsoleLogger, metrics *Metrics, wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:      id,
		InChan:  in,
		Logger:  logger,
		Metrics: metrics,
		Wg:      wg,
	}
}

func (w *Worker) Start(batchSize int) {
	go func() {
		defer w.Wg.Done()

		// create batch slice
		batch := make([]domain.LogMessage, 0, batchSize)

		for msg := range w.InChan {

			batch = append(batch, msg)

			// if batch full --> FLUSH
			if len(batch) >= batchSize {
				w.processBatch(batch)

				// Batch reset (no allocation!)
				batch = batch[:0]
			}

			if w.Metrics != nil {
				w.Metrics.IncProcessed()
			}
		}

		// Worker is going down -> flush rest of the batch
		if len(batch) > 0 {
			w.processBatch(batch)
		}
	}()
}

func (w *Worker) processBatch(batch []domain.LogMessage) {
	fmt.Printf("Worker %d flushing batch of %d logs\n", w.ID, len(batch))
}
