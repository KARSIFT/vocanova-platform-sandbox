package password

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	h, err := Hash("a long pasted password with spaces")
	require.NoError(t, err)
	ok, err := Verify(h, "a long pasted password with spaces")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = Verify(h, "different long pasted password")
	require.NoError(t, err)
	assert.False(t, ok)
}
func TestPasswordBoundsAndMalformedHash(t *testing.T) {
	assert.ErrorIs(t, Validate("short"), ErrInvalidPassword)
	assert.ErrorIs(t, Validate(strings.Repeat("a", 129)), ErrInvalidPassword)
	_, err := Verify("bad", "anything sufficiently long")
	assert.ErrorIs(t, err, ErrMalformedHash)
}
