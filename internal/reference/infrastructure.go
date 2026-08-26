package reference

type PostgresRepository struct {
	DSN                string
	MaximumConnections int
}

func NewPostgresRepository(dsn string, maximum int) *PostgresRepository {
	return &PostgresRepository{DSN: dsn, MaximumConnections: maximum}
}
func (r *PostgresRepository) DriverName() string { return "postgres" }
