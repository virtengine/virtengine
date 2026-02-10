package enclave_runtime

import (
	"testing"
)

func TestScrubBytes(t *testing.T) {
	data := []byte("sensitive data that must be scrubbed")
	original := make([]byte, len(data))
	copy(original, data)

	ScrubBytes(data)

	// Verify all bytes are zero
	for i, b := range data {
		if b != 0 {
			t.Errorf("byte at position %d should be 0, got %d", i, b)
		}
	}
}

func TestScrubBytes_Empty(t *testing.T) {
	// Should not panic on empty slice
	ScrubBytes([]byte{})
	ScrubBytes(nil)
}

func TestScrubString(t *testing.T) {
	s := "sensitive string"
	ScrubString(&s)

	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestScrubString_Nil(t *testing.T) {
	// Should not panic on nil
	ScrubString(nil)

	// Should not panic on empty string
	s := ""
	ScrubString(&s)
}

func TestScrubFixedSize(t *testing.T) {
	var key FixedSizeKey32
	copy(key[:], []byte("12345678901234567890123456789012"))

	ScrubFixedSize(&key)

	for i, b := range key {
		if b != 0 {
			t.Errorf("byte at position %d should be 0, got %d", i, b)
		}
	}
}

func TestSensitiveBuffer(t *testing.T) {
	buf := NewSensitiveBuffer(1024)

	// Write data
	n, err := buf.Write([]byte("secret data"))
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != 11 {
		t.Errorf("expected 11 bytes written, got %d", n)
	}

	// Read data
	if buf.Len() != 11 {
		t.Errorf("expected length 11, got %d", buf.Len())
	}

	data := buf.Bytes()
	if string(data) != "secret data" {
		t.Errorf("expected 'secret data', got %q", string(data))
	}

	// Destroy
	buf.Destroy()

	if !buf.IsDestroyed() {
		t.Error("expected buffer to be destroyed")
	}

	if buf.Bytes() != nil {
		t.Error("expected nil bytes after destroy")
	}
}

func TestSensitiveBuffer_WriteAfterDestroy(t *testing.T) {
	buf := NewSensitiveBuffer(100)
	buf.Destroy()

	_, err := buf.Write([]byte("test"))
	if err == nil {
		t.Error("expected error writing to destroyed buffer")
	}
}

func TestSecureContext(t *testing.T) {
	ctx := NewSecureContext()

	buf1 := ctx.AllocateBuffer(100)
	buf2 := ctx.AllocateBuffer(200)

	_, _ = buf1.Write([]byte("data1"))
	_, _ = buf2.Write([]byte("data2"))

	cleanupCalled := false
	ctx.RegisterCleanup(func() {
		cleanupCalled = true
	})

	ctx.Destroy()

	if !cleanupCalled {
		t.Error("expected cleanup to be called")
	}

	if !buf1.IsDestroyed() {
		t.Error("expected buf1 to be destroyed")
	}
	if !buf2.IsDestroyed() {
		t.Error("expected buf2 to be destroyed")
	}
}

func TestProcessingScope(t *testing.T) {
	exitCalled := false

	func() {
		scope := NewProcessingScope()
		defer scope.Complete()

		buf := scope.AllocateBuffer(100)
		_, _ = buf.Write([]byte("processing data"))

		scope.OnExit(func() {
			exitCalled = true
		})
	}()

	if !exitCalled {
		t.Error("expected OnExit to be called")
	}
}

func TestFixedSizeKey32(t *testing.T) {
	var key FixedSizeKey32
	copy(key[:], []byte("12345678901234567890123456789012"))

	bytes := key.Bytes()
	if len(bytes) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(bytes))
	}

	key.Destroy()

	for _, b := range key {
		if b != 0 {
			t.Error("expected key to be zeroed after Destroy")
			break
		}
	}
}

func TestFixedSizeKey64(t *testing.T) {
	var key FixedSizeKey64
	for i := range key {
		key[i] = byte(i)
	}

	bytes := key.Bytes()
	if len(bytes) != 64 {
		t.Errorf("expected 64 bytes, got %d", len(bytes))
	}

	key.Destroy()

	for _, b := range key {
		if b != 0 {
			t.Error("expected key to be zeroed after Destroy")
			break
		}
	}
}

func TestProcessingScope_MultipleBuffers(t *testing.T) {
	scope := NewProcessingScope()
	defer scope.Complete()

	buffers := make([]*SensitiveBuffer, 10)
	for i := range buffers {
		buffers[i] = scope.AllocateBuffer(100)
		_, _ = buffers[i].Write([]byte("data"))
	}

	// Verify all buffers work
	for i, buf := range buffers {
		if buf.Len() != 4 {
			t.Errorf("buffer %d: expected len 4, got %d", i, buf.Len())
		}
	}
}

func BenchmarkScrubBytes(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScrubBytes(data)
	}
}

func BenchmarkScrubFixedSize32(b *testing.B) {
	var key FixedSizeKey32
	for i := range key {
		key[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScrubFixedSize(&key)
	}
}
