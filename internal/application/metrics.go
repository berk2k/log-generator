package application

import (
	"fmt"
	"log-generator/internal/domain"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Produced    uint64
	Processed   uint64
	Channel     <-chan domain.LogMessage
	WorkerCount int
}

func NewMetrics(ch <-chan domain.LogMessage, workerCount int) *Metrics {
	return &Metrics{
		Channel:     ch,
		WorkerCount: workerCount,
	}
}

func (m *Metrics) IncProduced() {
	atomic.AddUint64(&m.Produced, 1)
}

func (m *Metrics) IncProcessed() {
	atomic.AddUint64(&m.Processed, 1)
}

func (m *Metrics) Start() {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			produced := atomic.SwapUint64(&m.Produced, 0)
			processed := atomic.SwapUint64(&m.Processed, 0)
			queueLen := len(m.Channel)
			queueCap := cap(m.Channel)

			fmt.Printf("[METRICS] Produced=%d/s | Processed=%d/s | Queue=%d/%d | Workers=%d\n",
				produced, processed, queueLen, queueCap, m.WorkerCount)
		}
	}()
}
