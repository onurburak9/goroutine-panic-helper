package goroutine_panic_helper

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type GoroutineGroup struct {
	wg      sync.WaitGroup
	ctx     context.Context
	handler PanicHandler
	errOnce sync.Once
	err     error
}

// PanicHandler is a function type that defines how panics should be handled
type PanicHandler func(interface{}, []byte)

func NewGoroutineGroup(ctx context.Context, handler PanicHandler) *GoroutineGroup {
	gg := &GoroutineGroup{ctx: ctx}
	if handler == nil {
		handler = DefaultPanicHandler
	}
	gg.handler = handler
	return gg
}

func (gg *GoroutineGroup) Go(fn func(context.Context)) {
	gg.wg.Add(1)
	go func() {
		defer gg.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				gg.handler(r, stack)
				err := recoveryToError(r)
				
				gg.errOnce.Do(func() {
					gg.err = err
				})
			}
		}()
		fn(gg.ctx)
	}()
}

func (gg *GoroutineGroup) Wait() error {
	gg.wg.Wait()
	return gg.err
}

func DefaultPanicHandler(panic interface{}, stack []byte) {
	fmt.Printf("Panic: %v\nStack: %s\n", panic, string(stack))
}

func recoveryToError(recovery any) error {
	switch value := recovery.(type) {
	case string:
		return fmt.Errorf("panic recovery: %s", value)
	case error:
		return fmt.Errorf("panic recovery: %w", value)
	default:
		return fmt.Errorf("panic recovery: %v", value)
	}
}
