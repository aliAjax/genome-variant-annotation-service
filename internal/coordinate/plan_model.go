package coordinate

type PlanStatus string

const (
	PlanQueued    PlanStatus = "queued"
	PlanRunning   PlanStatus = "running"
	PlanPaused    PlanStatus = "paused"
	PlanCompleted PlanStatus = "completed"
	PlanFailed    PlanStatus = "failed"
)

type Plan struct {
	ID          string
	Status      PlanStatus
	TotalChunks int
	NextChunk   int
	LastError   string
}

func (p Plan) CanTransition(next PlanStatus) bool {
	transitions := map[PlanStatus]map[PlanStatus]bool{
		PlanQueued:    {PlanRunning: true},
		PlanRunning:   {PlanPaused: true, PlanCompleted: true, PlanFailed: true},
		PlanPaused:    {PlanRunning: true, PlanCompleted: true},
		PlanCompleted: {PlanRunning: true},
		PlanFailed:    {PlanCompleted: true},
	}
	return transitions[p.Status][next]
}
