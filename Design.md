# High-Throughput Log Pipeline – Design & Trade-offs

This document explains the internal design decisions, trade-offs, and alternative approaches considered while building the pipeline.

The goal of this project is not to build a production Logstash clone, but to deeply understand the mechanics behind high-throughput log ingestion systems.

---

# 1: Producer Design

## Why a Ticker-Based Producer?

The producer uses `time.Ticker` to emit logs at a configurable interval rather than generating as fast as possible.

### Why?
- Simulates realistic, rate-controlled log sources
- Prevents immediate queue saturation at startup
- Allows backpressure behavior to be observed clearly

### Trade-off:
- Fixed-interval production doesn't model bursty real-world sources
- A burst model (e.g., Poisson arrival) would be more realistic but adds complexity

---

## Why JSON Format?

```json
{"msg": "hello from producer", "level": "INFO", "ts": 1712345678123456}
```

### Why?
- Machine-readable from the start
- Matches real log shipper formats (Filebeat, Fluent Bit)
- Enables field-level processing downstream

### Trade-off:
- JSON serialization has overhead compared to plain strings
- At extreme throughput, this cost is measurable

---

# 2: Buffered Channel as the Queue

## Why a Buffered Channel?

The pipeline uses a single `chan []byte` with a fixed capacity as the queue between producer and workers.

### Why?
- Native Go primitive — no external dependency
- Blocking semantics built-in (producer blocks when full)
- Provides natural backpressure without extra coordination

### Trade-off:
- In-memory only — no persistence across restarts
- Single channel = single partition (no ordering guarantees across workers)
- Channel closure is the shutdown signal, which requires careful coordination

### Alternative considered:
- Ring buffer / circular queue for lower allocation overhead
- Rejected: channel semantics are clearer and sufficient for this scope

---

# 3: Backpressure Handling

## Why Adaptive Producer Slowdown?

When queue utilization crosses thresholds, the producer sleeps:

| Queue Usage | Producer Sleep |
|-------------|----------------|
| > 95%       | 2ms            |
| > 90%       | 1ms            |
| > 80%       | 500µs          |
| Normal      | 0              |

### Why?
- Prevents the producer from spinning on a full channel
- Reduces CPU thrashing under saturation
- Self-stabilizing: as workers drain the queue, producer speeds up automatically

### Trade-off:
- Sleep-based throttling is coarse — not a precise rate limiter
- A token bucket or leaky bucket algorithm would be more precise
- Added complexity not warranted at this scale

### Alternative considered:
- Drop logs when queue is full (rejected — loses data)
- Block silently via channel send (rejected — no visibility into pressure)

---

# 4: Worker Pool & Batch Processing

## Why Batch Instead of Per-Log Processing?

Each worker buffers logs internally and flushes when:
- `batchSize` is reached, or
- `batchTimeout` expires

### Why?
- Dramatically reduces per-log overhead (context switching, I/O calls)
- Models real-world pipelines (Kafka consumers, Logstash output plugins)
- Timeout flush prevents stale logs accumulating in low-throughput periods

### Trade-off:
- Batch introduces latency — a log may wait up to `batchTimeout` before processing
- Larger batches = higher throughput but worse tail latency

Tuning is workload-dependent.

---

## Why Per-Worker StopChan?

Each worker has its own `StopChan chan struct{}` for individual termination.

### Why?
- Allows the autoscaler to stop specific workers during scale-down
- Decouples worker lifecycle from the global context

### Trade-off:
- Requires mutex-protected worker tracking
- Slightly more complex than a simple global stop signal

### Alternative considered:
- Cancel individual worker contexts
- Equivalent complexity, StopChan is more explicit

---

# 5: Auto-Scaling

## Why Threshold-Based Scaling?

The autoscaler polls queue utilization at a fixed interval:

```
queue_usage = len(queue) / cap(queue)
```

- Scale up when usage > 80%
- Scale down when usage < 20%

### Why?
- Simple to reason about
- Mirrors Kubernetes HPA behavior
- Avoids oscillation: the gap between thresholds creates a hysteresis band

### Trade-off:
- Poll-based — reacts with up to one `AutoScaleInterval` delay
- Threshold tuning is workload-sensitive
- Can oscillate if thresholds are too close together

### Alternative considered:
- Event-driven scaling triggered by channel pressure — more responsive but harder to bound
- Rejected for simplicity

---

## Why Mutex-Protected Worker Pool?

The worker slice is protected by a `sync.Mutex` during add/remove operations.

### Why?
- Autoscaler and shutdown path both modify worker state concurrently
- Mutex prevents race conditions on the worker list

### Trade-off:
- Lock contention is negligible at this worker count
- At high worker counts (100+), a lock-free structure would matter

---

# 6: Graceful Shutdown

## Why Context Cancellation + Channel Close?

Shutdown uses two signals:
1. Context cancellation — stops producer and autoscaler
2. Channel close — signals workers that no more logs will arrive

### Why?
- Producer and autoscaler are context-aware components
- Workers naturally exit their `range queue` loop when the channel is closed
- WaitGroup ensures all workers flush before the program exits

### Trade-off:
- Channel close is a one-way, irreversible signal
- Any producer writing after close would panic — requires careful ordering

Shutdown order is critical:
1. Cancel context (stop producer, autoscaler)
2. Close channel (signal workers)
3. Wait for WaitGroup

---

## Why Flush-Before-Exit in Workers?

Each worker flushes its remaining batch before returning.

### Why?
- Zero log loss guarantee
- Without this, partial batches would be discarded on shutdown

### Trade-off:
- Shutdown time is bounded by the largest in-flight batch flush
- Acceptable given the small batch sizes in this implementation

---

# 7: Metrics Design

## Why Inline Metrics Printing?

Metrics are printed to stdout at a fixed interval:

```
[METRICS] Produced=1477/s | Processed=150/s | Queue=1277/1500 | Workers=2
```

### Why?
- Zero dependency observability
- Immediately visible during development and testing
- Shows backpressure dynamics in real time

### Trade-off:
- Not queryable or persistent
- Not suitable for production alerting

### Future improvement:
- Expose `/metrics` in Prometheus exposition format
- Use atomic counters to avoid lock contention on metric reads

---

# 8: Concurrency Model

## Why Goroutines + Channels Instead of Thread Pool?

Go's goroutine scheduler handles multiplexing onto OS threads automatically.

### Why?
- Goroutines are cheap (2KB stack, grows as needed)
- Channel-based communication avoids shared memory bugs
- Select-based event loops allow clean multi-condition waiting

### Trade-off:
- Goroutine leaks are silent — requires careful WaitGroup discipline
- Channel direction types (`chan<-`, `<-chan`) should be used at boundaries for safety

---

# 9: Architecture Layout

```
domain/          → log types, interfaces
application/     → producer, worker, autoscaler, metrics
infrastructure/  → channel queue, configuration
cmd/             → entry point
```

### Why this layout?
- Separates concerns clearly
- Domain logic is independent of infrastructure
- Mirrors clean architecture / hexagonal patterns

### Trade-off:
- Overhead for a project this size
- Justified as a learning exercise in production-style structure

---

# 10: Known Limitations

- In-memory only — no persistence
- Single channel = no partitioning
- No message ordering guarantees across workers
- No retry or dead letter handling for failed log processing
- Metrics are ephemeral (stdout only)

These are intentional omissions to keep focus on concurrency and backpressure mechanics.

---

# 11: Future Improvements

- Persistent queue (WAL or disk-backed ring buffer)
- Partitioned channels with per-partition workers
- Prometheus `/metrics` endpoint with histogram buckets
- Token bucket rate limiter for the producer
- Per-worker processing latency tracking
- Dead letter handling for unprocessable logs

---

# Systems Learning Outcomes

This project demonstrates understanding of:

- Go concurrency primitives (goroutines, channels, select, context, WaitGroup)
- Producer-consumer pipeline design
- Backpressure detection and adaptive throttling
- Dynamic worker scaling with hysteresis
- Batch processing trade-offs (throughput vs latency)
- Zero data loss shutdown sequencing
- Clean architecture in Go

The implementation prioritizes clarity of mechanics over feature completeness.

---

# Closing Note

This pipeline is intentionally minimal.

Its purpose is to expose the internal mechanics behind:

- Logstash / Fluent Bit worker models
- Kafka consumer group scaling
- Kubernetes HPA behavior

By building these primitives manually, the underlying system behavior becomes explicit instead of hidden behind abstractions.
