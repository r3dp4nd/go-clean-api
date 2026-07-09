package audit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu     sync.RWMutex
	nextID int
	events []Event
}

var _ Recorder = (*MemoryStore)(nil)
var _ Reader = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		nextID: 1,
		events: make([]Event, 0),
	}
}

func (s *MemoryStore) Record(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("%d", s.nextID)
		s.nextID++
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	if event.Payload == nil {
		event.Payload = map[string]any{}
	}

	s.events = append(s.events, event)

	return nil
}

func (s *MemoryStore) ListByAggregate(
	ctx context.Context,
	aggregateType string,
	aggregateID string,
	input ListEventsInput,
) (ListEventsResult, error) {
	if err := ctx.Err(); err != nil {
		return ListEventsResult{}, err
	}

	input = NormalizeListEventsInput(input)

	aggregateType = strings.TrimSpace(aggregateType)
	aggregateID = strings.TrimSpace(aggregateID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]Event, 0)

	for index := len(s.events) - 1; index >= 0; index-- {
		item := s.events[index]

		if item.AggregateType == aggregateType && item.AggregateID == aggregateID {
			filtered = append(filtered, item)
		}
	}

	total := len(filtered)
	totalPages := CalculateTotalPages(total, input.PageSize)

	offset := (input.Page - 1) * input.PageSize
	if offset >= total {
		return ListEventsResult{
			Items:      []Event{},
			Total:      total,
			Page:       input.Page,
			PageSize:   input.PageSize,
			TotalPages: totalPages,
		}, nil
	}

	end := offset + input.PageSize
	if end > total {
		end = total
	}

	return ListEventsResult{
		Items:      filtered[offset:end],
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
