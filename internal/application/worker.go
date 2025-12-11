package application

import (
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

func (w *Worker) Start() {
	go func() {
		defer w.Wg.Done() // notify main

		for msg := range w.InChan {
			_ = w.Logger.Write(msg)

			if w.Metrics != nil {
				w.Metrics.IncProcessed()
			}
		}
	}()
}
