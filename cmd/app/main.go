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

	// cancel context for producer
	ctx, cancel := context.WithCancel(context.Background())

	// catch SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// main log channel
	logChan := make(chan domain.LogMessage, 1500)

	// METRICS
	metrics := application.NewMetrics(logChan)
	metrics.Start()

	// LOGGER
	logger := infrastructure.NewConsoleLogger()

	// WAITGROUP (all workers)
	var wg sync.WaitGroup

	// BATCH SETTINGS
	batchSize := 50
	batchTimeout := 500 * time.Millisecond

	// WORKER POOL
	pool := application.NewWorkerPool(
		logChan,
		logger,
		metrics,
		&wg,
		batchSize,
		batchTimeout,
	)

	// START WITH MIN WORKERS
	minWorkers := 1
	for i := 0; i < minWorkers; i++ {
		pool.AddWorker()
	}

	// AUTOSCALER
	scaler := application.NewAutoScaler(
		pool,
		logChan,
		1*time.Second, // check every 1s
		minWorkers,    // min workers
		10,            // max workers
	)
	scaler.Start()

	// PRODUCER
	producer := application.NewProducer(
		logChan,
		100*time.Microsecond,
		metrics,
		ctx,
	)
	producer.Start()

	// WAIT FOR SHUTDOWN SIGNAL
	<-sigChan
	println("Shutdown signal received!")

	// STOP PRODUCER
	cancel()

	// Give producer time to exit
	time.Sleep(200 * time.Millisecond)

	// CLOSE CHANNEL (workers will exit automatically)
	close(logChan)

	// WAIT FOR ALL WORKERS
	wg.Wait()

	fmt.Println("Program finished")
}
