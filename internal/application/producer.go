package application

import (
	"context"
	"log-generator/internal/domain"
	"time"
)

type Producer struct {
	OutChan chan<- domain.LogMessage // send-only
	Rate    time.Duration
	Metrics *Metrics
	Ctx     context.Context
}

func NewProducer(out chan<- domain.LogMessage, rate time.Duration, metrics *Metrics, ctx context.Context) *Producer {
	return &Producer{
		OutChan: out,
		Rate:    rate,
		Metrics: metrics,
		Ctx:     ctx,
	}
}

func (p *Producer) Start() {
	go func() { //goroutine
		ticker := time.NewTicker(p.Rate)
		defer ticker.Stop()

		for {
			select {

			// shutdown signal
			case <-p.Ctx.Done():
				// producer stop
				return

			case <-ticker.C:
				// 1) calculate queue usage
				queueLen := len(p.OutChan)
				queueCap := cap(p.OutChan)
				usage := float64(queueLen) / float64(queueCap)

				// 2) adaptive pacing (flow control)

				// heavy backpressure -> slow down more
				if usage > 0.9 {
					time.Sleep(2 * time.Millisecond)
				} else if usage > 0.8 {
					time.Sleep(1 * time.Millisecond)
				} else if usage > 0.5 {
					time.Sleep(500 * time.Microsecond)
				}

				// 3) produce log
				msg := domain.LogMessage{
					Msg:   "hello from producer",
					Level: "INFO",
					Ts:    time.Now().UnixNano(),
				}

				p.OutChan <- msg
				if p.Metrics != nil {
					p.Metrics.IncProduced()
				}
			}
		}

	}()
}
