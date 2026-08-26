package job

import (
	"github.com/example/genome-variant-annotation/internal/annotation"
	"github.com/example/genome-variant-annotation/internal/variant"
	"time"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusPartial   Status = "partial"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Job struct {
	ID              string              `json:"id"`
	DatasetID       string              `json:"dataset_id"`
	Status          Status              `json:"status"`
	Input           []variant.Variant   `json:"-"`
	Results         []annotation.Result `json:"results,omitempty"`
	Failures        []Failure           `json:"failures,omitempty"`
	Progress        int                 `json:"progress"`
	CreatedAt       time.Time           `json:"created_at"`
	StartedAt       *time.Time          `json:"started_at,omitempty"`
	FinishedAt      *time.Time          `json:"finished_at,omitempty"`
	CancelRequested bool                `json:"cancel_requested"`
}

type Failure struct {
	Index      int    `json:"index"`
	VariantKey string `json:"variant_key"`
	Error      string `json:"error"`
}
