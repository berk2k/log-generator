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

	// logger
	logger := infrastructure.NewConsoleLogger()

	// producer: for all 1ms 1 log = 1000 log/s
	producer := application.NewProducer(logChan, time.Millisecond)
	producer.Start()

	// worker pool
	for i := 1; i <= 5; i++ {
		worker := application.NewWorker(i, logChan, logger)
		worker.Start()
	}

	time.Sleep(5 * time.Second)
	fmt.Println("Program finished")
	//select {} // endless loop
}
