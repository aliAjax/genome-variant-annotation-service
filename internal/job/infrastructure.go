package job

type PostgresRepository struct{ DSN string }

func NewPostgresRepository(dsn string) *PostgresRepository { return &PostgresRepository{DSN: dsn} }

type QueueConfig struct {
	Capacity     int
	Workers      int
	LeaseSeconds int
}

func DefaultQueueConfig() QueueConfig {
	return QueueConfig{Capacity: 1024, Workers: 4, LeaseSeconds: 30}
}
