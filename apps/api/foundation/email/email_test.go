package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeSenderRecordsMessages(t *testing.T) {
	fake := &Fake{}
	msg := Message{
		To:       []Address{{Email: "a@example.com"}},
		Subject:  "hello",
		BodyText: "world",
	}
	require.NoError(t, fake.Send(context.Background(), msg))
	assert.Len(t, fake.Sent, 1)
	last, ok := fake.Last()
	require.True(t, ok)
	assert.Equal(t, "hello", last.Subject)
	fake.Reset()
	assert.Len(t, fake.Sent, 0)
	_, ok = fake.Last()
	assert.False(t, ok)
}
