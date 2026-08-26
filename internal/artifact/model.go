package artifact

import "time"

type Status string

const (
	StatusQueued   Status = "queued"
	StatusBuilding Status = "building"
	StatusReady    Status = "ready"
	StatusFailed   Status = "failed"
)

type Bundle struct {
	ID        string    `json:"id"`
	DatasetID string    `json:"dataset_id"`
	Status    Status    `json:"status"`
	Objects   []string  `json:"objects"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
