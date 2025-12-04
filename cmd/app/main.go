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
	workers := []*application.Worker{}

	for i := 1; i <= 5; i++ {
		worker := application.NewWorker(i, logChan, logger)
		workers = append(workers, worker)
		worker.Start()
	}

	for i := 0; i < 2; i++ {
		time.Sleep(1 * time.Second)
		fmt.Print("Stats: ")

		for _, w := range workers {
			fmt.Printf("W%d=%d ", w.ID, w.Count)
		}

		fmt.Println()
	}

	fmt.Println("Program finished")

	//select {} // endless loop
}
