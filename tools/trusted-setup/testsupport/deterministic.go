package testsupport

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"sync"
	"time"
)

var randMu sync.Mutex

type deterministicReader struct {
	seed    []byte
	counter uint64
	buffer  []byte
	offset  int
}

func (r *deterministicReader) Read(p []byte) (int, error) {
	written := 0
	for written < len(p) {
		if r.offset >= len(r.buffer) {
			block := sha256.Sum256(append(append([]byte{}, r.seed...), counterBytes(r.counter)...))
			r.buffer = block[:]
			r.offset = 0
			r.counter++
		}
		n := copy(p[written:], r.buffer[r.offset:])
		written += n
		r.offset += n
	}
	return written, nil
}

func WithDeterministicRand(seed string, fn func() error) error {
	randMu.Lock()
	defer randMu.Unlock()

	previous := cryptorand.Reader
	cryptorand.Reader = io.Reader(&deterministicReader{seed: []byte(seed)})
	defer func() {
		cryptorand.Reader = previous
	}()

	return fn()
}

type StepClock struct {
	mu      sync.Mutex
	current time.Time
	step    time.Duration
}

func NewStepClock(start time.Time, step time.Duration) *StepClock {
	return &StepClock{
		current: start.UTC(),
		step:    step,
	}
}

func (c *StepClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.current
	c.current = c.current.Add(c.step)
	return now
}

func counterBytes(value uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)
	return buf
}
