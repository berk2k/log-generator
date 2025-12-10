package main

import (
	"fmt"
	"log-generator/internal/application"
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
	"time"
)

func main() {
	logChan := make(chan domain.LogMessage, 10000)

	workerCount := 5
	metrics := application.NewMetrics(logChan, workerCount)
	metrics.Start()

	// logger
	logger := infrastructure.NewConsoleLogger()

	// producer: for all 1ms 1 log = 1000 log/s
	producer := application.NewProducer(logChan, time.Millisecond, metrics)
	producer.Start()

	// worker pool

	for i := 1; i <= workerCount; i++ {
		worker := application.NewWorker(i, logChan, logger, metrics)
		worker.Start()
	}

	time.Sleep(2 * time.Second) // run for 2 seconds

	fmt.Println("Program finished")

	//select {} // endless loop
}
