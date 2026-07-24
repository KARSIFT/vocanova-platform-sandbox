// Package email provides the outbound email boundary used by authentication.
package email

import "context"

// Address is a recipient address.
type Address struct {
	Email string
	Name  string
}

// Message is a single outbound message.
type Message struct {
	To       []Address
	Subject  string
	BodyText string
	BodyHTML string
}

// Sender delivers messages. Implementations must not log or return sensitive
// bearer tokens or secrets.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Fake records every sent message for tests. It is safe for concurrent use.
type Fake struct {
	Sent []Message
}

// Send appends the message to the in-memory list.
func (f *Fake) Send(ctx context.Context, msg Message) error {
	f.Sent = append(f.Sent, msg)
	return nil
}

// Last returns the most recently sent message, if any.
func (f *Fake) Last() (Message, bool) {
	if len(f.Sent) == 0 {
		return Message{}, false
	}
	return f.Sent[len(f.Sent)-1], true
}

// Reset clears the recorded messages.
func (f *Fake) Reset() { f.Sent = f.Sent[:0] }
