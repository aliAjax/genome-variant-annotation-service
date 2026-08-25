package reference

import "time"

type Status string

const (
	StatusDraft      Status = "draft"
	StatusValidating Status = "validating"
	StatusPublished  Status = "published"
	StatusRetired    Status = "retired"
)

type Dataset struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Build       string     `json:"build"`
	Version     int64      `json:"version"`
	Digest      string     `json:"digest"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Records     int        `json:"records"`
}
type Feature struct {
	DatasetID  string            `json:"dataset_id"`
	Chromosome string            `json:"chromosome"`
	Start      int64             `json:"start"`
	End        int64             `json:"end"`
	Kind       string            `json:"kind"`
	ID         string            `json:"id"`
	Gene       string            `json:"gene,omitempty"`
	Transcript string            `json:"transcript,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (f Feature) Overlaps(start, end int64) bool { return f.Start <= end && f.End >= start }
