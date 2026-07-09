package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/r3dp4nd/go-clean-api/internal/audit"
)

const actorHeader = "X-Actor"

func auditContextFromRequest(r *http.Request) context.Context {
	actor := strings.TrimSpace(r.Header.Get(actorHeader))
	if actor == "" {
		actor = audit.DefaultActor
	}

	return audit.WithMetadata(r.Context(), audit.Metadata{
		RequestID: getRequestID(r.Context()),
		Actor:     actor,
	})
}
