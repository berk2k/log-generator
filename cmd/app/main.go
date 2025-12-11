package main

import (
	"context"
	"fmt"
	"log-generator/internal/application"
	"log-generator/internal/domain"
	"log-generator/internal/infrastructure"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	// cancel context
	ctx, cancel := context.WithCancel(context.Background())

	// catch SIGINT(Ctrl + C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logChan := make(chan domain.LogMessage, 10000)

	workerCount := 5
	metrics := application.NewMetrics(logChan, workerCount)
	metrics.Start()

	// logger
	logger := infrastructure.NewConsoleLogger()

	// producer: for all 1ms 1 log = 1000 log/s
	producer := application.NewProducer(logChan, time.Millisecond, metrics, ctx)
	producer.Start()

	// workers will be using WaitGroup
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// worker pool

	for i := 1; i <= workerCount; i++ {
		worker := application.NewWorker(i, logChan, logger, metrics, &wg)
		worker.Start()
	}

	// wait for shutdown signal
	<-sigChan
	println("Shutdown signal received!")

	// stop producer
	cancel()

	time.Sleep(200 * time.Millisecond) //wait for all producers to close before shutdown

	close(logChan)

	wg.Wait()

	fmt.Println("Program finished")

	//select {} // endless loop
}
