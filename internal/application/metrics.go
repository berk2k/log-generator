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
	WorkerCount int64
}

func NewMetrics(ch <-chan domain.LogMessage) *Metrics {
	return &Metrics{
		Channel: ch,
	}
}

/* ---------- PRODUCED ---------- */

func (m *Metrics) IncProduced() {
	atomic.AddUint64(&m.Produced, 1)
}

/* ---------- PROCESSED ---------- */

// batch-aware processed increment
func (m *Metrics) IncProcessedBy(n int) {
	atomic.AddUint64(&m.Processed, uint64(n))
}

/* ---------- WORKERS ---------- */

func (m *Metrics) SetWorkers(n int) {
	atomic.StoreInt64(&m.WorkerCount, int64(n))
}

/* ---------- START ---------- */

func (m *Metrics) Start() {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			produced := atomic.SwapUint64(&m.Produced, 0)
			processed := atomic.SwapUint64(&m.Processed, 0)
			workers := atomic.LoadInt64(&m.WorkerCount)

			queueLen := len(m.Channel)
			queueCap := cap(m.Channel)

			fmt.Printf(
				"[METRICS] Produced=%d/s | Processed=%d/s | Queue=%d/%d | Workers=%d\n",
				produced,
				processed,
				queueLen,
				queueCap,
				workers,
			)
		}
	}()
}
