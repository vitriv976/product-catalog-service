// Package repo implements repository patterns
package repo

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/m_outbox"
)

// OutboxRepo implements the outbox repository
type OutboxRepo struct {
	model *m_outbox.Model
}

// NewOutboxRepo creates a new outbox repository
func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{
		model: m_outbox.NewModel(),
	}
}

// InsertMut returns a mutation to insert an outbox event
func (r *OutboxRepo) InsertMut(event *contracts.OutboxEvent) *spanner.Mutation {
	data := &m_outbox.Data{
		EventID:     event.EventID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload:     event.Payload,
		Status:      event.Status,
		CreatedAt:   event.CreatedAt.(time.Time),
	}

	return r.model.InsertMut(data)
}

// EnrichDomainEvent converts a domain event to an outbox event with metadata
func EnrichDomainEvent(event domain.DomainEvent, aggregateID string) *contracts.OutboxEvent {
	payloadBytes, _ := json.Marshal(event)

	return &contracts.OutboxEvent{
		EventID:     uuid.New().String(),
		EventType:   event.EventType(),
		AggregateID: aggregateID,
		Payload:     string(payloadBytes),
		Status:      m_outbox.StatusPending,
		CreatedAt:   time.Now(),
	}
}
