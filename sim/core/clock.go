package core

import "time"

// Clock tracks simulation time progression.
type Clock struct {
	current time.Time
	step    time.Duration
}

// NewClock creates a new simulation clock.
func NewClock(start time.Time, step time.Duration) *Clock {
	return &Clock{current: start, step: step}
}

// Now returns the current simulated time.
func (c *Clock) Now() time.Time {
	return c.current
}

// Step advances the clock and returns the new time.
func (c *Clock) Step() time.Time {
	c.current = c.current.Add(c.step)
	return c.current
}
