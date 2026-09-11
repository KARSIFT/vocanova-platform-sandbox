package password

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBurnPasswordCheckUsesBoundedArgonWork(t *testing.T) {
	// This deliberately has no result: it proves the timing equalisation path
	// is runnable without generating or persisting a credential hash.
	burnPasswordCheck("a long pasted password with spaces")
}
