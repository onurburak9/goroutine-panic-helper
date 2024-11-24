package goroutine_panic_helper

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"math/rand"
)

func TestGroup_Go(t *testing.T) {
	ctx := context.Background()
	executed := false

	group := NewGoroutineGroup(ctx, nil)
	group.Go(func(ctx context.Context) {
		executed = true
	})

	err := group.Wait()

	if err != nil || !executed {
		t.Error("Function was not executed")
	}
}

func TestGroup_GoWithPanic(t *testing.T) {
	ctx := context.Background()
	panicCaught := false

	customHandler := func(r interface{}, stack []byte) {
		panicCaught = true
		if r.(string) != "test panic" {
			t.Errorf("Expected 'test panic', got %v", r)
		}
	}

	group := NewGoroutineGroup(ctx, customHandler)
	group.Go(func(ctx context.Context) {
		panic("test panic")
	})

	group.Wait()

	if !panicCaught {
		t.Error("Panic was not caught by custom handler")
	}
}

func TestGroup_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	group := NewGoroutineGroup(ctx, nil)
	executed := false

	group.Go(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			// Context cancelled as expected
			executed = true
		case <-time.After(100 * time.Millisecond):
			t.Error("Context cancellation not detected")
		}
	})

	group.Wait()

	if !executed {
		t.Error("Context cancellation handler was not executed")
	}
}

func TestGroup_Error(t *testing.T) {
	ctx := context.Background()
	group := NewGoroutineGroup(ctx, nil)
	group.Go(func(ctx context.Context) {
		panic("test panic")
	})
	err := group.Wait()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGroup_MultipleGoroutines_Success(t *testing.T) {
	ctx := context.Background()
	group := NewGoroutineGroup(ctx, nil)

	completed := make([]bool, 5)
	for i := 0; i < 5; i++ {
		i := i // capture loop variable
		group.Go(func(ctx context.Context) {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			completed[i] = true
		})
	}

	err := group.Wait()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	for i, done := range completed {
		if !done {
			t.Errorf("Goroutine %d did not complete", i)
		}
	}
}

func TestGroup_MultipleGoroutines_Panic(t *testing.T) {
	ctx := context.Background()
	panicCaught := false

	handler := func(r interface{}, stack []byte) {
		panicCaught = true
		if r.(string) != "intentional panic" {
			t.Errorf("Expected 'intentional panic', got %v", r)
		}
	}

	group := NewGoroutineGroup(ctx, handler)

	// Launch multiple goroutines, one will panic
	for i := 0; i < 5; i++ {
		i := i
		group.Go(func(ctx context.Context) {
			if i == 2 {
				panic("intentional panic")
			}
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		})
	}

	err := group.Wait()
	if err == nil {
		t.Error("Expected an error from Wait(), got nil")
	}
	if !panicCaught {
		t.Error("Panic was not caught by custom handler")
	}
	if !strings.Contains(err.Error(), "intentional panic") {
		t.Errorf("Unexpected error message: %v", err)
	}

}

func TestGroup_MultipleGoroutines_MultiplePanics(t *testing.T) {
	ctx := context.Background()
	panicCount := int32(0)

	handler := func(r interface{}, stack []byte) {
		atomic.AddInt32(&panicCount, 1)
	}

	group := NewGoroutineGroup(ctx, handler)

	// Launch goroutines that all panic
	for i := 0; i < 3; i++ {
		i := i
		group.Go(func(ctx context.Context) {
			panic(fmt.Sprintf("panic %d", i))
		})
	}

	err := group.Wait()
	if err == nil {
		t.Error("Expected an error from Wait(), got nil")
	}

	// Only the first panic should be returned as error
	if !strings.Contains(err.Error(), "panic 0") &&
		!strings.Contains(err.Error(), "panic 1") &&
		!strings.Contains(err.Error(), "panic 2") {
		t.Errorf("Unexpected error message: %v", err)
	}

	if count := atomic.LoadInt32(&panicCount); count != 3 {
		t.Errorf("Expected 3 panics to be caught, got %d", count)
	}
}

func TestGroup_MultipleGoroutines_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	group := NewGoroutineGroup(ctx, nil)

	// Start some long-running goroutines
	for i := 0; i < 3; i++ {
		group.Go(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				t.Error("Goroutine should have been cancelled")
			}
		})
	}

	// Cancel context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := group.Wait()
	if err != nil {
		t.Errorf("Expected no error from cancelled goroutines, got: %v", err)
	}
}
