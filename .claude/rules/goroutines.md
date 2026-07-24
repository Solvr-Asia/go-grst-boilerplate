# Goroutine Best Practices

Reference: https://github.com/superbolang/golang-goroutines_problem

## 1. Goroutine Leaks (CRITICAL)

Goroutines that never terminate consume resources indefinitely.

```go
// ✗ Bad: Goroutine leak - no way to stop
func startWorker() {
    go func() {
        for {
            doWork()
        }
    }()
}

// ✓ Good: Use context for cancellation
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // Clean exit
            default:
                doWork()
            }
        }
    }()
}

// ✓ Good: Use done channel
func startWorker(done <-chan struct{}) {
    go func() {
        for {
            select {
            case <-done:
                return
            default:
                doWork()
            }
        }
    }()
}
```

## 2. Race Conditions (CRITICAL)

Multiple goroutines accessing shared memory without synchronization.

```go
// ✗ Bad: Race condition
var counter int

func increment() {
    go func() { counter++ }()
    go func() { counter++ }()
}

// ✓ Good: Use mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// ✓ Good: Use atomic operations
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

// ✓ Good: Use channels for communication
func increment(counterCh chan<- int) {
    counterCh <- 1
}
```

**Always run tests with race detection:**

```bash
go test -race ./...
```

## 3. Deadlocks

Goroutines blocked indefinitely, waiting on resources held by others.

```go
// ✗ Bad: Potential deadlock with unbuffered channel
func process() {
    ch := make(chan int)
    ch <- 1  // Blocks forever - no receiver
    <-ch
}

// ✓ Good: Use buffered channel or goroutine
func process() {
    ch := make(chan int, 1)  // Buffered
    ch <- 1
    <-ch
}

// ✓ Good: Use select with timeout
func process() {
    ch := make(chan int)
    select {
    case ch <- 1:
        // sent
    case <-time.After(5 * time.Second):
        // timeout
    }
}
```

## 4. Resource Exhaustion

Creating excessive goroutines that exhaust system resources.

```go
// ✗ Bad: Unbounded goroutine creation
func processItems(items []Item) {
    for _, item := range items {
        go process(item)  // Could create millions of goroutines
    }
}

// ✓ Good: Worker pool pattern
func processItems(ctx context.Context, items []Item) {
    const numWorkers = 10
    jobs := make(chan Item, len(items))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case item, ok := <-jobs:
                    if !ok {
                        return
                    }
                    process(item)
                }
            }
        }()
    }

    // Send jobs
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    wg.Wait()
}

// ✓ Good: Use semaphore pattern
func processItems(items []Item) {
    sem := make(chan struct{}, 10)  // Limit to 10 concurrent
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        sem <- struct{}{}  // Acquire
        go func(item Item) {
            defer wg.Done()
            defer func() { <-sem }()  // Release
            process(item)
        }(item)
    }

    wg.Wait()
}
```

## 5. Channel Blocking

Operations on channels causing unexpected blocking.

```go
// ✗ Bad: Blocking forever
func fetch(url string) string {
    ch := make(chan string)
    go func() {
        // If this panics, main goroutine blocks forever
        ch <- httpGet(url)
    }()
    return <-ch
}

// ✓ Good: Select with timeout
func fetch(ctx context.Context, url string) (string, error) {
    ch := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := httpGet(url)
        if err != nil {
            errCh <- err
            return
        }
        ch <- result
    }()

    select {
    case result := <-ch:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-ctx.Done():
        return "", ctx.Err()
    case <-time.After(30 * time.Second):
        return "", errors.New("request timeout")
    }
}
```

## 6. Unhandled Panics

Panic in a goroutine crashes only that goroutine, leaving application in inconsistent state.

```go
// ✗ Bad: Panic crashes silently
func processAsync(data string) {
    go func() {
        // If this panics, main app continues but this goroutine dies
        process(data)
    }()
}

// ✓ Good: Recover from panics
func processAsync(data string) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Recovered from panic: %v\nStack: %s", r, debug.Stack())
                // Report to monitoring system
            }
        }()
        process(data)
    }()
}

// ✓ Good: Use error channel for goroutine errors
func processAsync(ctx context.Context, data string) <-chan error {
    errCh := make(chan error, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                errCh <- fmt.Errorf("panic: %v", r)
            }
        }()
        if err := process(data); err != nil {
            errCh <- err
            return
        }
        errCh <- nil
    }()
    return errCh
}
```

## 7. Goroutine Monitoring

```go
// Monitor goroutine count in production
import "runtime"

func monitorGoroutines() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        count := runtime.NumGoroutine()
        log.Printf("Active goroutines: %d", count)
        if count > 10000 {
            log.Warn("High goroutine count detected!")
        }
    }
}
```
