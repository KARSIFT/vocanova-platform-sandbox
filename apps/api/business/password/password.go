// Package password implements bounded Argon2id password handling. Raw
// passwords never leave this package except transiently in caller memory.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	minChars           = 15
	maxChars           = 128
	maxBytes           = 512
	memoryKiB   uint32 = 19 * 1024
	iterations  uint32 = 2
	parallelism uint8  = 1
	saltBytes          = 16
	keyBytes           = 32
)

var ErrInvalidPassword = errors.New("password must be 15 to 128 characters")
var ErrMalformedHash = errors.New("malformed password hash")

// argonWork bounds concurrent 19MiB Argon2 allocations. Waiting callers do
// not allocate password-work memory until a slot is available.
var argonWork = make(chan struct{}, 4)

func withArgon[T any](fn func() T) T {
	argonWork <- struct{}{}
	defer func() { <-argonWork }()
	return fn()
}

func Validate(value string) error {
	if len(value) > maxBytes || !utf8.ValidString(value) || utf8.RuneCountInString(value) < minChars || utf8.RuneCountInString(value) > maxChars {
		return ErrInvalidPassword
	}
	return nil
}

// Hash returns a portable, parameterized Argon2id encoding.
func Hash(value string) (string, error) {
	if err := Validate(value); err != nil {
		return "", err
	}
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}
	key := withArgon(func() []byte {
		return argon2.IDKey([]byte(value), salt, iterations, memoryKiB, parallelism, keyBytes)
	})
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memoryKiB, iterations, parallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func Verify(encoded, value string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memoryKiB, iterations, parallelism) {
		return false, ErrMalformedHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != saltBytes {
		return false, ErrMalformedHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) != keyBytes {
		return false, ErrMalformedHash
	}
	actual := withArgon(func() []byte {
		return argon2.IDKey([]byte(value), salt, iterations, memoryKiB, parallelism, keyBytes)
	})
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}
