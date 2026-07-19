package worker

import "context"

type JobHandler interface {
	Handle(ctx context.Context, payload []byte) error
}

type Queue interface {
	Publish(ctx context.Context, stream string, payload []byte) error
	Consume(ctx context.Context, stream string, group string, handler JobHandler) error
}
