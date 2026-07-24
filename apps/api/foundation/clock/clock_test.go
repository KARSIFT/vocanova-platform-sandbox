package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRealClockReturnsUTCTime(t *testing.T) {
	before := time.Now().UTC()
	got := Real{}.Now()
	after := time.Now().UTC()
	assert.True(t, got.Equal(got.UTC()))
	assert.False(t, got.Before(before))
	assert.False(t, got.After(after))
}

func TestFixedClockReturnsConfiguredTime(t *testing.T) {
	fixed := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	c := Fixed{T: fixed}
	assert.Equal(t, fixed, c.Now())
	c.Advance(time.Minute)
	assert.Equal(t, fixed.Add(time.Minute), c.Now())
}
