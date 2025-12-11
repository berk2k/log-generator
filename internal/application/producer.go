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
	ticker := time.NewTicker(p.Rate)

	go func() { //goroutine
		defer ticker.Stop()

		for {
			select {

			// shutdown signal
			case <-p.Ctx.Done():
				// producer stop
				return

			case <-ticker.C:
				msg := domain.LogMessage{
					Msg:   "hello from producer",
					Level: "INFO",
					Ts:    time.Now().UnixNano(),
				}

				if p.Metrics != nil {
					p.Metrics.IncProduced()
				}

				p.OutChan <- msg
			}
		}

	}()
}
