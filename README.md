# Goroutine Panic Helper

A lightweight Go package that provides safe goroutine execution with panic recovery, error propagation, and context support. This package helps you manage groups of goroutines while properly handling panics and cancellation.

## Features

- ✨ Simple and clean API
- 🛡️ Automatic panic recovery
- 🔄 Context support for cancellation
- ⚡ Efficient error propagation
- 🎯 Custom panic handlers
- 🔍 Detailed stack traces

## Installation

```bash
go get github.com/onurburak9/goroutine-panic-helper
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "time"
    gh "github.com/onurburak9/goroutine-panic-helper"
)

func main() {
    // Create a context
    ctx := context.Background()
    
    // Create a new goroutine group with default panic handler
    group := gh.NewGoroutineGroup(ctx, nil)

    // Run goroutines
    group.Go(func(ctx context.Context) {
        // Your code here
        time.Sleep(100 * time.Millisecond)
    })

    // Wait for all goroutines to complete
    if err := group.Wait(); err != nil {
        // Handle any panics that were converted to errors
        log.Printf("Error occurred: %v", err)
    }
}
```

### Custom Panic Handler

```go
customHandler := func(r interface{}, stack []byte) {
    logger.WithFields(log.Fields{
        "panic": r,
        "stack": string(stack),
    }).Error("Goroutine panicked")
}

group := gh.NewGoroutineGroup(ctx, customHandler)
```

### With Context Cancellation

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

group := gh.NewGoroutineGroup(ctx, nil)

group.Go(func(ctx context.Context) {
    select {
    case <-ctx.Done():
        // Handle cancellation
        return
    case <-time.After(1 * time.Second):
        // Do work
    }
})
```

### Running Multiple Goroutines

```go
group := gh.NewGoroutineGroup(ctx, nil)

// Launch multiple goroutines
for i := 0; i < 5; i++ {
    i := i // capture loop variable
    group.Go(func(ctx context.Context) {
        // Each goroutine is protected from panics
        processItem(ctx, i)
    })
}

// Wait for all goroutines and check for errors
if err := group.Wait(); err != nil {
    log.Printf("At least one goroutine panicked: %v", err)
}
```

## Error Handling

The package converts panics to errors that can be handled normally:

- String panics: `"panic recovery: <string>"`
- Error panics: `"panic recovery: <error>"`
- Other panics: `"panic recovery: %v"`

Only the first panic in a group is returned as an error from `Wait()`, though all panics are passed to the panic handler if one is provided.

## Best Practices

1. Always check the error returned by `Wait()`
2. Use context for proper cancellation
3. Consider providing a custom panic handler for logging
4. Don't share variables between goroutines without proper synchronization
5. Remember that only the first panic will be returned as an error

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.