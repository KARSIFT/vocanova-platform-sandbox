// Package meaningidentity identifies the exact catalog meaning saved in a lesson.
package meaningidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Revision binds a saved exercise to its canonical spelling, use and definition.
// Older snapshots with incomplete metadata deliberately have no revision.
func Revision(id, word, pos, definition string) string {
	if id == "" || word == "" || pos == "" || definition == "" {
		return ""
	}
	data, _ := json.Marshal([4]string{id, word, pos, definition})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
