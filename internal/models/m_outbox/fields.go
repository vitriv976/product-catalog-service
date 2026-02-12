// Package m_outbox provides the outbox events table model
package m_outbox

// Table and field constants
const (
	TableName = "outbox_events"

	// Field names
	EventID     = "event_id"
	EventType   = "event_type"
	AggregateID = "aggregate_id"
	Payload     = "payload"
	Status      = "status"
	CreatedAt   = "created_at"
	ProcessedAt = "processed_at"
)

// Event status constants
const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusFailed    = "failed"
)
