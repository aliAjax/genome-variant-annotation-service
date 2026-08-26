package platform

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"time"
)

var (
	ErrInvalid   = errors.New("invalid input")
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrCancelled = errors.New("cancelled")
	ErrBudget    = errors.New("resource budget exceeded")
)

type Clock interface{ Now() time.Time }
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
func NewID(prefix string) string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return prefix + "-fallback"
	}
	return prefix + "-" + hex.EncodeToString(b)
}

type Config struct {
	HTTPAddr        string
	AuthToken       string
	WorkerCount     int
	MaximumVCFBytes int64
}

func LoadConfig() Config {
	c := Config{HTTPAddr: ":8080", AuthToken: "dev-token", WorkerCount: 4, MaximumVCFBytes: 16 << 20}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("AUTH_TOKEN"); v != "" {
		c.AuthToken = v
	}
	if v, err := strconv.Atoi(os.Getenv("WORKER_COUNT")); err == nil && v > 0 {
		c.WorkerCount = v
	}
	return c
}
