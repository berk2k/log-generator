package application

import (
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
)

type Worker struct {
	ID     int
	InChan <-chan domain.LogMessage //recieve-only
	Logger *infrastructure.ConsoleLogger
	Count  int
}

func NewWorker(id int, in <-chan domain.LogMessage, logger *infrastructure.ConsoleLogger) *Worker {
	return &Worker{
		ID:     id,
		InChan: in,
		Logger: logger,
	}
}

func (w *Worker) Start() {
	go func() {
		for msg := range w.InChan {
			_ = w.Logger.Write(msg)
			w.Count++
		}
	}()
}
