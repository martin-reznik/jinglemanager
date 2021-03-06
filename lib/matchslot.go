package lib

import (
	"time"
)

// MatchSlot - struct holds info about matches
type MatchSlot struct {
	StartsAt time.Time
	Duration time.Duration
	context  *Context
}

// NewMatchSlot - will create new match
func NewMatchSlot(startTime time.Time, duration time.Duration, ctx *Context) *MatchSlot {
	slot := &MatchSlot{
		StartsAt: startTime,
		Duration: duration,
		context:  ctx,
	}
	return slot
}

// Overlaps - return true if those two match slots are overlapsing
func (m *MatchSlot) Overlaps(m2 MatchSlot) bool {
	if m.StartsAt.After(m2.StartsAt) && m.StartsAt.Before(m2.StartsAt.Add(m2.Duration)) {
		return true
	}
	if m2.StartsAt.After(m.StartsAt) && m2.StartsAt.Before(m.StartsAt.Add(m.Duration)) {
		return true
	}
	if m.StartsAt.Equal(m2.StartsAt) {
		return true
	}
	return false
}
