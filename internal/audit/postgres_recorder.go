package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRecorder struct {
	pool *pgxpool.Pool
}

var _ Recorder = (*PostgresRecorder)(nil)

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
