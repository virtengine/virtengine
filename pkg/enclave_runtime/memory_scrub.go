package enclave_runtime

import (
	"runtime"
	"unsafe"
)

// MemoryScrubber provides utilities for securely scrubbing memory
// after processing sensitive data inside the enclave.
//
// Security Notes:
// - These functions attempt to zero memory before it's garbage collected
// - Go's garbage collector may move memory, so these are best-effort
// - For maximum security, use stack-allocated fixed-size arrays
// - Production enclaves should use platform-specific secure memory APIs

// ScrubBytes overwrites a byte slice with zeros
// WARNING: Go GC may have already copied the data elsewhere
func ScrubBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	for i := range data {
		data[i] = 0
	}
	// Attempt to prevent compiler optimization from removing the scrub
	runtime.KeepAlive(data)
}

// ScrubString creates a copy of the string, scrubs the copy's underlying bytes,
// and returns an empty string. Note: The original string is immutable in Go.
func ScrubString(s *string) {
	if s == nil || *s == "" {
		return
	}
	// In Go, strings are immutable so we can only clear the reference
	*s = ""
}

// ScrubFixedSize scrubs fixed-size arrays (recommended for sensitive data)
func ScrubFixedSize[T any](data *T) {
	if data == nil {
		return
	}
	//nolint:gosec // G103: Sizeof is only used to bound the in-place scrub length (no pointer arithmetic).
	size := unsafe.Sizeof(*data)
	if size == 0 {
		return
	}
	maxInt := uintptr(^uint(0) >> 1)
	if size > maxInt {
		return
	}
	//nolint:gosec // G103: pointer conversion required to zero fixed-size sensitive data in place (no aliasing outside bounds).
	ptr := unsafe.Pointer(data)

	// Zero the memory
	//nolint:gosec // G103: unsafe slice is intentional to zero fixed-size sensitive data without allocations; size is bounded.
	bytes := unsafe.Slice((*byte)(ptr), int(size))
	for i := range bytes {
		bytes[i] = 0
	}
	runtime.KeepAlive(data)
}

// SensitiveBuffer provides a buffer that tracks its sensitive content
// and provides a Destroy method for secure cleanup
type SensitiveBuffer struct {
	data      []byte
	destroyed bool
}

// NewSensitiveBuffer creates a new sensitive buffer with the given capacity
func NewSensitiveBuffer(capacity int) *SensitiveBuffer {
	return &SensitiveBuffer{
		data:      make([]byte, 0, capacity),
		destroyed: false,
	}
}

// Write appends data to the buffer
func (sb *SensitiveBuffer) Write(p []byte) (n int, err error) {
	if sb.destroyed {
		return 0, ErrEnclaveNotInitialized
	}
	sb.data = append(sb.data, p...)
	return len(p), nil
}

// Bytes returns the buffer contents
func (sb *SensitiveBuffer) Bytes() []byte {
	if sb.destroyed {
		return nil
	}
	return sb.data
}

// Len returns the length of the buffer
func (sb *SensitiveBuffer) Len() int {
	return len(sb.data)
}

// Destroy securely destroys the buffer contents
func (sb *SensitiveBuffer) Destroy() {
	if sb.destroyed {
		return
	}
	ScrubBytes(sb.data)
	sb.data = nil
	sb.destroyed = true
}

// IsDestroyed returns whether the buffer has been destroyed
func (sb *SensitiveBuffer) IsDestroyed() bool {
	return sb.destroyed
}

// SecureContext provides a context for processing sensitive data
// that automatically scrubs memory when done
type SecureContext struct {
	buffers []*SensitiveBuffer
	cleanup []func()
}

// NewSecureContext creates a new secure processing context
func NewSecureContext() *SecureContext {
	return &SecureContext{
		buffers: make([]*SensitiveBuffer, 0),
		cleanup: make([]func(), 0),
	}
}

// AllocateBuffer allocates a tracked sensitive buffer
func (sc *SecureContext) AllocateBuffer(capacity int) *SensitiveBuffer {
	buf := NewSensitiveBuffer(capacity)
	sc.buffers = append(sc.buffers, buf)
	return buf
}

// RegisterCleanup registers a cleanup function to be called on Destroy
func (sc *SecureContext) RegisterCleanup(fn func()) {
	sc.cleanup = append(sc.cleanup, fn)
}

// Destroy scrubs all allocated buffers and runs cleanup functions
func (sc *SecureContext) Destroy() {
	// Run cleanup functions in reverse order
	for i := len(sc.cleanup) - 1; i >= 0; i-- {
		sc.cleanup[i]()
	}

	// Destroy all buffers
	for _, buf := range sc.buffers {
		buf.Destroy()
	}

	sc.buffers = nil
	sc.cleanup = nil
}

// ProcessingScope represents a scope for processing sensitive data
// with automatic cleanup via defer
type ProcessingScope struct {
	ctx    *SecureContext
	onExit func()
}

// NewProcessingScope creates a new processing scope
// Usage:
//
//	scope := NewProcessingScope()
//	defer scope.Complete()
//	buf := scope.AllocateBuffer(1024)
//	// use buffer...
//	// buffer is automatically scrubbed when Complete() is called
func NewProcessingScope() *ProcessingScope {
	return &ProcessingScope{
		ctx: NewSecureContext(),
	}
}

// AllocateBuffer allocates a buffer within this scope
func (ps *ProcessingScope) AllocateBuffer(capacity int) *SensitiveBuffer {
	return ps.ctx.AllocateBuffer(capacity)
}

// OnExit registers a callback to run when the scope exits
func (ps *ProcessingScope) OnExit(fn func()) {
	ps.onExit = fn
}

// Complete completes the processing scope and scrubs all data
func (ps *ProcessingScope) Complete() {
	if ps.onExit != nil {
		ps.onExit()
	}
	ps.ctx.Destroy()
}

// FixedSizeKey represents a fixed-size cryptographic key
// Use this instead of slices for sensitive key material
type FixedSizeKey32 [32]byte

// Destroy scrubs the key
func (k *FixedSizeKey32) Destroy() {
	ScrubFixedSize(k)
}

// Bytes returns the key as a byte slice
func (k *FixedSizeKey32) Bytes() []byte {
	return k[:]
}

// FixedSizeKey64 represents a 64-byte key or hash
type FixedSizeKey64 [64]byte

// Destroy scrubs the key
func (k *FixedSizeKey64) Destroy() {
	ScrubFixedSize(k)
}

// Bytes returns the key as a byte slice
func (k *FixedSizeKey64) Bytes() []byte {
	return k[:]
}
