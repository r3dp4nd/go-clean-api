package audit

import (
	"context"
	"time"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type Event struct {
	ID            string
	Type          string
	AggregateType string
	AggregateID   string
	Payload       map[string]any
	CreatedAt     time.Time
}

type ListEventsInput struct {
	Page     int
	PageSize int
}

type ListEventsResult struct {
	Items      []Event
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type Reader interface {
	ListByAggregate(
		ctx context.Context,
		aggregateType string,
		aggregateID string,
		input ListEventsInput,
	) (ListEventsResult, error)
}

type NoopRecorder struct{}

var _ Recorder = (*NoopRecorder)(nil)

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (r *NoopRecorder) Record(ctx context.Context, event Event) error {
	return nil
}

func NormalizeListEventsInput(input ListEventsInput) ListEventsInput {
	if input.Page <= 0 {
		input.Page = DefaultPage
	}

	if input.PageSize <= 0 {
		input.PageSize = DefaultPageSize
	}

	if input.PageSize > MaxPageSize {
		input.PageSize = MaxPageSize
	}

	return input
}

func CalculateTotalPages(total int, pageSize int) int {
	if total == 0 {
		return 0
	}

	return (total + pageSize - 1) / pageSize
}
