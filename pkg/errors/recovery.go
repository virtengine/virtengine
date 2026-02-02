package errors

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// PanicHandler is a function that handles recovered panics.
type PanicHandler func(recovered interface{}, stack []byte)

var (
	defaultPanicHandler     PanicHandler
	defaultPanicHandlerOnce sync.Once
)

// SetDefaultPanicHandler sets the default panic handler for all recovery operations.
func SetDefaultPanicHandler(handler PanicHandler) {
	defaultPanicHandlerOnce.Do(func() {
		defaultPanicHandler = handler
	})
}

// GetDefaultPanicHandler returns the default panic handler.
func GetDefaultPanicHandler() PanicHandler {
	if defaultPanicHandler != nil {
		return defaultPanicHandler
	}
	// Default handler: just print to stderr
	return func(recovered interface{}, stack []byte) {
		fmt.Printf("PANIC recovered: %v\nStack trace:\n%s\n", recovered, stack)
	}
}

// RecoverAndLog recovers from panics and logs them.
// Should be used with defer at the beginning of goroutines.
//
// Example:
//
//	go func() {
//	    defer errors.RecoverAndLog("worker goroutine")
//	    // ... work ...
//	}()
func RecoverAndLog(context string) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		handler := GetDefaultPanicHandler()
		if context != "" {
			handler(fmt.Sprintf("%s: %v", context, r), stack)
		} else {
			handler(r, stack)
		}
	}
}

// RecoverToError recovers from panics and converts them to errors.
// Returns nil if no panic occurred.
//
// Example:
//
//	func DoWork() (err error) {
//	    defer func() {
//	        if recErr := errors.RecoverToError("DoWork"); recErr != nil {
//	            err = recErr
//	        }
//	    }()
//	    // ... work that might panic ...
//	    return nil
//	}
func RecoverToError(context string) error {
	if r := recover(); r != nil {
		stack := debug.Stack()
		handler := GetDefaultPanicHandler()
		handler(r, stack)

		var msg string
		if context != "" {
			msg = fmt.Sprintf("panic in %s: %v", context, r)
		} else {
			msg = fmt.Sprintf("panic: %v", r)
		}
		return NewInternalError("", 0, context, msg)
	}
	return nil
}

// SafeGo runs a function in a goroutine with panic recovery.
// Panics are logged but do not crash the program.
//
// Example:
//
//	errors.SafeGo("background worker", func() {
//	    // ... work ...
//	})
func SafeGo(context string, fn func()) {
	go func() {
		defer RecoverAndLog(context)
		fn()
	}()
}

// SafeGoWithError runs a function in a goroutine with panic recovery.
// Errors (including panics) are sent to the error channel.
//
// Example:
//
//	errCh := make(chan error, 1)
//	errors.SafeGoWithError("worker", errCh, func() error {
//	    // ... work ...
//	    return nil
//	})
//	if err := <-errCh; err != nil {
//	    log.Error("worker failed", "error", err)
//	}
func SafeGoWithError(context string, errCh chan<- error, fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				handler := GetDefaultPanicHandler()
				handler(r, stack)

				var msg string
				if context != "" {
					msg = fmt.Sprintf("panic in %s: %v", context, r)
				} else {
					msg = fmt.Sprintf("panic: %v", r)
				}
				errCh <- NewInternalError("", 0, context, msg)
			}
		}()

		if err := fn(); err != nil {
			errCh <- err
		}
	}()
}

// RecoverWithCleanup recovers from panics and runs cleanup functions.
// Cleanup functions are always executed, even if no panic occurred.
//
// Example:
//
//	defer errors.RecoverWithCleanup("worker", func() {
//	    // cleanup resources
//	})
func RecoverWithCleanup(context string, cleanup func()) {
	cleanup()
	RecoverAndLog(context)
}

// MustNotPanic wraps a function and panics with a custom message if the function panics.
// This is useful for testing or situations where panics should never occur.
//
// Example:
//
//	errors.MustNotPanic("critical operation", func() {
//	    // ... work that must not panic ...
//	})
func MustNotPanic(context string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			panic(fmt.Sprintf("MUST NOT PANIC: %s panicked with: %v\nStack:\n%s", context, r, stack))
		}
	}()
	fn()
}

// TryWithPanicRecovery attempts to run a function and returns any error or recovered panic.
//
// Example:
//
//	err := errors.TryWithPanicRecovery("risky operation", func() error {
//	    // ... work that might panic or error ...
//	    return nil
//	})
func TryWithPanicRecovery(context string, fn func() error) (err error) {
	defer func() {
		if recErr := RecoverToError(context); recErr != nil {
			err = recErr
		}
	}()
	return fn()
}

// SafeClose safely closes a channel with panic recovery.
// Returns true if channel was closed, false if already closed or panic occurred.
func SafeClose(ch chan struct{}) bool {
	defer func() {
		_ = recover()
	}()
	close(ch)
	return true
}

// PanicGroup manages multiple goroutines with panic recovery.
// Similar to sync.WaitGroup but with panic handling.
type PanicGroup struct {
	wg      sync.WaitGroup
	errCh   chan error
	errOnce sync.Once
	err     error
}

// NewPanicGroup creates a new PanicGroup.
func NewPanicGroup() *PanicGroup {
	return &PanicGroup{
		errCh: make(chan error, 1),
	}
}

// Go runs a function in a goroutine with panic recovery.
func (pg *PanicGroup) Go(context string, fn func() error) {
	pg.wg.Add(1)
	go func() {
		defer pg.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				handler := GetDefaultPanicHandler()
				handler(r, stack)

				err := fmt.Errorf("panic in %s: %v", context, r)
				pg.errOnce.Do(func() {
					pg.err = err
					pg.errCh <- err
				})
			}
		}()

		if err := fn(); err != nil {
			pg.errOnce.Do(func() {
				pg.err = err
				pg.errCh <- err
			})
		}
	}()
}

// Wait waits for all goroutines to complete and returns the first error (if any).
func (pg *PanicGroup) Wait() error {
	pg.wg.Wait()
	close(pg.errCh)
	return pg.err
}
