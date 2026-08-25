package job

import "context"

type Queue interface {
	Enqueue(context.Context, string) error
	Dequeue(context.Context) (string, error)
}
type Lease interface {
	Acquire(context.Context, string, string) error
	Renew(context.Context, string, string) error
	Release(context.Context, string, string) error
}
