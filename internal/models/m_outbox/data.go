// Package m_outbox provides the outbox events table model
package m_outbox

import (
	"time"

	"cloud.google.com/go/spanner"
)

// Data represents an outbox event row in the database
type Data struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string // JSON string
	Status      string
	CreatedAt   time.Time
	ProcessedAt spanner.NullTime
}

// Model provides typed Spanner mutation operations
type Model struct{}

// NewModel creates a new outbox model
func NewModel() *Model {
	return &Model{}
}

// InsertMut creates an insert mutation
func (m *Model) InsertMut(data *Data) *spanner.Mutation {
	return spanner.Insert(TableName, []string{
		EventID,
		EventType,
		AggregateID,
		Payload,
		Status,
		CreatedAt,
		ProcessedAt,
	}, []interface{}{
		data.EventID,
		data.EventType,
		data.AggregateID,
		data.Payload, // JSON stored as string
		data.Status,
		data.CreatedAt,
		data.ProcessedAt,
	})
}

// UpdateMut creates an update mutation
func (m *Model) UpdateMut(eventID string, updates map[string]interface{}) *spanner.Mutation {
	if len(updates) == 0 {
		return nil
	}

	updates[EventID] = eventID

	columns := make([]string, 0, len(updates))
	values := make([]interface{}, 0, len(updates))

	columns = append(columns, EventID)
	values = append(values, eventID)

	for col, val := range updates {
		if col != EventID {
			columns = append(columns, col)
			values = append(values, val)
		}
	}

	return spanner.Update(TableName, columns, values)
}
