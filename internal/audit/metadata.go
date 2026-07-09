package audit

import (
	"context"
	"strings"
)

const DefaultActor = "anonymous"

type Metadata struct {
	RequestID string
	Actor     string
}

type metadataContextKey struct{}

func WithMetadata(ctx context.Context, metadata Metadata) context.Context {
	metadata.RequestID = strings.TrimSpace(metadata.RequestID)
	metadata.Actor = strings.TrimSpace(metadata.Actor)

	if metadata.Actor == "" {
		metadata.Actor = DefaultActor
	}

	return context.WithValue(ctx, metadataContextKey{}, metadata)
}

func MetadataFromContext(ctx context.Context) (Metadata, bool) {
	metadata, ok := ctx.Value(metadataContextKey{}).(Metadata)
	if !ok {
		return Metadata{}, false
	}

	metadata.RequestID = strings.TrimSpace(metadata.RequestID)
	metadata.Actor = strings.TrimSpace(metadata.Actor)

	if metadata.Actor == "" {
		metadata.Actor = DefaultActor
	}

	return metadata, true
}
