package platform

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Cursor struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func EncodeCursor(cursor Cursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
func DecodeCursor(value string) (Cursor, error) {
	if value == "" {
		return Cursor{}, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Cursor{}, ErrInvalid
	}
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return Cursor{}, ErrInvalid
	}
	if cursor.ID == "" || cursor.Timestamp.IsZero() {
		return Cursor{}, ErrInvalid
	}
	return cursor, nil
}
func PageSize(value, defaultValue, maximum int) int {
	if value < 1 {
		return defaultValue
	}
	if value > maximum {
		return maximum
	}
	return value
}
