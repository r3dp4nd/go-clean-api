package audit

import (
	"context"
	"time"
)

type Event struct {
	ID            string
	Type          string
	AggregateType string
	AggregateID   string
	Payload       map[string]any
	CreatedAt     time.Time
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type NoopRecorder struct{}

var _ Recorder = (*NoopRecorder)(nil)

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (r *NoopRecorder) Record(ctx context.Context, event Event) error {
	return nil
}
