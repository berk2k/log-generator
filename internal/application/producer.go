package application

import (
	"log-generator/internal/domain"
	"time"
)

type Producer struct {
	OutChan chan<- domain.LogMessage // send-only
	Rate    time.Duration
}

func NewProducer(out chan<- domain.LogMessage, rate time.Duration) *Producer {
	return &Producer{
		OutChan: out,
		Rate:    rate,
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
			p.OutChan <- msg
		}
	}()
}
