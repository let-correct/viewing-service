package viewing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/troysnowden/viewing-service/internal/domain/viewing"
)

type mockPublisher struct {
	published []domain.Event
	err       error
}

func (m *mockPublisher) Publish(_ context.Context, event domain.Event) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, event)
	return nil
}

func TestScheduleViewing(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 10, 30, 0, 0, time.UTC)
	attendee := domain.Attendee{Name: "John Smith", Email: "john@example.com", Phone: "0821234567"}

	baseCmd := Command{
		CorrelationID: "google-event-id",
		SchemaVersion: "1.0",
		DetailType:    domain.DetailTypeScheduled,
		AgentEmail:    "agent@example.com",
		Address:       "10 Main St, Dublin",
		Start:         start,
		End:           end,
		Attendee:      attendee,
	}

	tests := []struct {
		name       string
		cmd        Command
		publishErr error
		wantErr    bool
		wantType   domain.DetailType
	}{
		{
			name:     "ViewingScheduled event is published with correct fields",
			cmd:      baseCmd,
			wantType: domain.DetailTypeScheduled,
		},
		{
			name:     "ViewingCancelled event is published with correct detail type",
			cmd:      Command{CorrelationID: "google-event-id", SchemaVersion: "1.0", DetailType: domain.DetailTypeCancelled, AgentEmail: "agent@example.com", Address: "10 Main St, Dublin", Start: start, End: end, Attendee: attendee},
			wantType: domain.DetailTypeCancelled,
		},
		{
			name:       "publisher error is propagated",
			cmd:        baseCmd,
			publishErr: errors.New("eventbridge unavailable"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub := &mockPublisher{err: tt.publishErr}
			s := &ProcessAppointment{publisher: pub, now: func() time.Time { return fixedTime }}

			err := s.Handle(context.Background(), tt.cmd)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, pub.published, 1)

			got := pub.published[0]
			assert.Equal(t, tt.wantType, got.DetailType)
			assert.True(t, got.Metadata.Timestamp.Equal(fixedTime))
			assert.Equal(t, source, got.Metadata.Source)
			assert.Equal(t, schemaVersion, got.Metadata.SchemaVersion)
			assert.Equal(t, tt.cmd.CorrelationID, got.Metadata.CorrelationID)
			assert.Equal(t, tt.cmd.AgentEmail, got.Payload.AgentEmail)
			assert.Equal(t, tt.cmd.Address, got.Payload.Address)
			assert.True(t, got.Payload.Start.Equal(tt.cmd.Start))
			assert.True(t, got.Payload.End.Equal(tt.cmd.End))
			assert.Equal(t, tt.cmd.Attendee, got.Payload.Attendee)
		})
	}
}
