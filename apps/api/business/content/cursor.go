package content

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type situationCursor struct {
	DisplayOrder int       `json:"d"`
	ID           uuid.UUID `json:"i"`
}

func encodeSituationCursor(c situationCursor) string {
	b, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(b)
}

func decodeSituationCursor(s string) (situationCursor, error) {
	var c situationCursor
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
