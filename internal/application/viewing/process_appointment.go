package viewing

import (
	"context"
	"fmt"
	"time"

	domain "github.com/troysnowden/viewing-service/internal/domain/viewing"
)

const (
	source        = "viewing-service"
	schemaVersion = "1.0"
)

type publisher interface {
	Publish(ctx context.Context, event domain.Event) error
}

type ProcessAppointment struct {
	publisher publisher
	now       func() time.Time
}

func NewProcessAppointment(publisher publisher) *ProcessAppointment {
	return &ProcessAppointment{publisher: publisher, now: time.Now}
}

type Command struct {
	CorrelationID string
	SchemaVersion string
	DetailType    domain.DetailType
	AgentEmail    string
	Address       string
	Start         time.Time
	End           time.Time
	Attendee      domain.Attendee
}

func (s *ProcessAppointment) Handle(ctx context.Context, cmd Command) error {
	event := domain.Event{
		DetailType: cmd.DetailType,
		Metadata: domain.EventMetadata{
			Timestamp:     s.now().UTC(),
			Source:        source,
			SchemaVersion: schemaVersion,
			CorrelationID: cmd.CorrelationID,
		},
		Payload: domain.Viewing{
			AgentEmail: cmd.AgentEmail,
			Address:    cmd.Address,
			Start:      cmd.Start,
			End:        cmd.End,
			Attendee:   cmd.Attendee,
		},
	}

	if err := s.publisher.Publish(ctx, event); err != nil {
		return fmt.Errorf("publish viewing event: %w", err)
	}

	return nil
}
