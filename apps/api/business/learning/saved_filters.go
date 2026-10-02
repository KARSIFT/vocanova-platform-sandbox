package learning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func prepareSavedWords(req ListSavedWordsRequest) (ListSavedWordsRequest, savedCursor, error) {
	var cursor savedCursor
	if !utf8.ValidString(req.Query) || strings.ContainsRune(req.Query, 0) {
		return req, cursor, ErrInvalidSavedFilter
	}
	req.Query = strings.ToLower(strings.TrimSpace(req.Query))
	if utf8.RuneCountInString(req.Query) > 100 {
		return req, cursor, ErrInvalidSavedFilter
	}
	switch req.Stage {
	case "", "new", "learning", "reviewing", "mastered", "ignored", "archived":
	default:
		return req, cursor, ErrInvalidSavedFilter
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	if req.AfterCursor != "" {
		if len(req.AfterCursor) > 2048 {
			return req, cursor, ErrInvalidCursor
		}
		var err error
		cursor, err = decodeSavedCursor(req.AfterCursor)
		// Pre-filter cursors have no requester/filter binding and must be restarted.
		if err != nil || cursor.Version != 1 || cursor.Filter != savedFilter(req) {
			return req, cursor, ErrInvalidCursor
		}
	}
	return req, cursor, nil
}

func savedFilter(req ListSavedWordsRequest) string {
	b, _ := json.Marshal(struct {
		User, Query, Stage string
		Due                bool
	}{req.UserID.String(), req.Query, req.Stage, req.DueOnly})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func nextSavedCursor(req ListSavedWordsRequest, item SavedMeaning) string {
	return encodeSavedCursor(savedCursor{Version: 1, Filter: savedFilter(req), AddedAt: item.AddedAt, ID: item.UserWordID})
}
