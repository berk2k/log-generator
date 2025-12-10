package application

import (
	"log-generator/internal/domain"
	"time"
)

type Producer struct {
	OutChan chan<- domain.LogMessage // send-only
	Rate    time.Duration
	Metrics *Metrics
}

func NewProducer(out chan<- domain.LogMessage, rate time.Duration, metrics *Metrics) *Producer {
	return &Producer{
		OutChan: out,
		Rate:    rate,
		Metrics: metrics,
	}
}

func (p *Producer) Start() {
	ticker := time.NewTicker(p.Rate)

	go func() { //goroutine
		for range ticker.C {
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
	}()
}
