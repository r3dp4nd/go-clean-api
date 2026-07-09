package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRecorder struct {
	pool *pgxpool.Pool
}

var _ Recorder = (*PostgresRecorder)(nil)
var _ Reader = (*PostgresRecorder)(nil)

func NewPostgresRecorder(pool *pgxpool.Pool) *PostgresRecorder {
	return &PostgresRecorder{
		pool: pool,
	}
}

func (r *PostgresRecorder) Record(ctx context.Context, event Event) error {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	const query = `
		INSERT INTO audit_events (
			event_type,
			aggregate_type,
			aggregate_id,
			payload
		)
		VALUES ($1, $2, $3, $4::jsonb)
	`

	if _, err := r.pool.Exec(
		ctx,
		query,
		event.Type,
		event.AggregateType,
		event.AggregateID,
		string(payloadJSON),
	); err != nil {
		return fmt.Errorf("record audit event: %w", err)
	}

	return nil
}

func (r *PostgresRecorder) ListByAggregate(
	ctx context.Context,
	aggregateType string,
	aggregateID string,
	input ListEventsInput,
) (ListEventsResult, error) {
	input = NormalizeListEventsInput(input)

	aggregateType = strings.TrimSpace(aggregateType)
	aggregateID = strings.TrimSpace(aggregateID)

	total, err := r.countByAggregate(ctx, aggregateType, aggregateID)
	if err != nil {
		return ListEventsResult{}, err
	}

	items, err := r.listByAggregate(ctx, aggregateType, aggregateID, input)
	if err != nil {
		return ListEventsResult{}, err
	}

	return ListEventsResult{
		Items:      items,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: CalculateTotalPages(total, input.PageSize),
	}, nil
}

func (r *PostgresRecorder) countByAggregate(
	ctx context.Context,
	aggregateType string,
	aggregateID string,
) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM audit_events
		WHERE aggregate_type = $1
		  AND aggregate_id = $2
	`

	var total int

	if err := r.pool.QueryRow(ctx, query, aggregateType, aggregateID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count audit events: %w", err)
	}

	return total, nil
}

func (r *PostgresRecorder) listByAggregate(
	ctx context.Context,
	aggregateType string,
	aggregateID string,
	input ListEventsInput,
) ([]Event, error) {
	offset := (input.Page - 1) * input.PageSize

	const query = `
		SELECT
			id::text,
			event_type,
			aggregate_type,
			aggregate_id,
			payload,
			created_at
		FROM audit_events
		WHERE aggregate_type = $1
		  AND aggregate_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		aggregateType,
		aggregateID,
		input.PageSize,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	items := make([]Event, 0)

	for rows.Next() {
		var item Event
		var payloadBytes []byte

		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.AggregateType,
			&item.AggregateID,
			&payloadBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}

		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &item.Payload); err != nil {
				return nil, fmt.Errorf("unmarshal audit payload: %w", err)
			}
		}

		if item.Payload == nil {
			item.Payload = map[string]any{}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}

	return items, nil
}
