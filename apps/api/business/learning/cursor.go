package learning

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type savedCursor struct {
	AddedAt time.Time `json:"a"`
	ID      uuid.UUID `json:"i"`
}

func encodeSavedCursor(c savedCursor) string {
	b, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(b)
}

func decodeSavedCursor(s string) (savedCursor, error) {
	var c savedCursor
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return c, fmt.Errorf("decode cursor: %w", err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, fmt.Errorf("parse cursor: %w", err)
	}
	if c.ID == uuid.Nil {
		return c, fmt.Errorf("cursor missing id")
	}
	return c, nil
}
