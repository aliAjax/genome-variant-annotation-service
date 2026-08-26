package review

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusReviewing Status = "reviewing"
	StatusReleased  Status = "released"
	StatusRejected  Status = "rejected"
)

var transitions = map[Status]map[Status]bool{
	StatusPending:   {StatusReviewing: true},
	StatusReviewing: {StatusReleased: true, StatusRejected: true},
	StatusRejected:  {StatusPending: true},
	StatusReleased:  {},
}

type Review struct {
	ID           string    `json:"id"`
	QuarantineID string    `json:"quarantine_id"`
	DatasetID    string    `json:"dataset_id"`
	Status       Status    `json:"status"`
	Reason       string    `json:"reason,omitempty"`
	Attempts     int       `json:"attempts"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r Review) CanTransition(next Status) bool {
	allowed, ok := transitions[r.Status]
	return ok && allowed[next]
}

func (r Review) Terminal() bool {
	return r.Status == StatusReleased || r.Status == StatusRejected
}
