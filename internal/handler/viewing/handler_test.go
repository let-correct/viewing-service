package viewing

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sqsviewing "github.com/troysnowden/viewing-service/internal/adapters/sqs/viewing"
	appviewing "github.com/troysnowden/viewing-service/internal/application/viewing"
	domain "github.com/troysnowden/viewing-service/internal/domain/viewing"
)

type mockScheduler struct {
	mock.Mock
}

func (m *mockScheduler) Handle(ctx context.Context, cmd appviewing.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func appointmentBody(detailType sqsviewing.AppointmentDetailType) string {
	start := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 10, 30, 0, 0, time.UTC)
	evt := sqsviewing.AppointmentEvent{
		DetailType: detailType,
		Metadata: sqsviewing.AppointmentMetadata{
			Timestamp:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Source:        "agent-service",
			SchemaVersion: "1.0",
			CorrelationID: "google-event-id",
		},
		Payload: sqsviewing.Appointment{
			AgentEmail: "agent@example.com",
			Location:   "10 Main St, Dublin",
			Start:      start,
			End:        end,
			Attendee:   sqsviewing.AppointmentAttendee{Name: "John Smith", Email: "john@example.com", Phone: "0821234567"},
		},
	}
	b, _ := json.Marshal(evt)
	return string(b)
}

func sqsRecord(messageID, body string) events.SQSMessage {
	return events.SQSMessage{MessageId: messageID, Body: body}
}

func TestHandle(t *testing.T) {
	start := time.Date(2024, 1, 20, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 20, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		records        []events.SQSMessage
		setupMock      func(*mockScheduler)
		wantFailureIDs []string
	}{
		{
			name:    "AppointmentCreated is mapped to ViewingScheduled and handled",
			records: []events.SQSMessage{sqsRecord("msg-1", appointmentBody(sqsviewing.AppointmentCreated))},
			setupMock: func(m *mockScheduler) {
				m.On("Handle", mock.Anything, appviewing.Command{
					CorrelationID: "google-event-id",
					SchemaVersion: "1.0",
					DetailType:    domain.DetailTypeScheduled,
					AgentEmail:    "agent@example.com",
					Address:       "10 Main St, Dublin",
					Start:         start,
					End:           end,
					Attendee:      domain.Attendee{Name: "John Smith", Email: "john@example.com", Phone: "0821234567"},
				}).Return(nil)
			},
		},
		{
			name:    "AppointmentCancelled is mapped to ViewingCancelled and handled",
			records: []events.SQSMessage{sqsRecord("msg-1", appointmentBody(sqsviewing.AppointmentCancelled))},
			setupMock: func(m *mockScheduler) {
				m.On("Handle", mock.Anything, mock.MatchedBy(func(cmd appviewing.Command) bool {
					return cmd.DetailType == domain.DetailTypeCancelled
				})).Return(nil)
			},
		},
		{
			name:           "invalid JSON body is reported as batch item failure",
			records:        []events.SQSMessage{sqsRecord("msg-bad", "not json")},
			setupMock:      func(m *mockScheduler) {},
			wantFailureIDs: []string{"msg-bad"},
		},
		{
			name:    "scheduler error is reported as batch item failure",
			records: []events.SQSMessage{sqsRecord("msg-1", appointmentBody(sqsviewing.AppointmentCreated))},
			setupMock: func(m *mockScheduler) {
				m.On("Handle", mock.Anything, mock.Anything).Return(errors.New("publish failed"))
			},
			wantFailureIDs: []string{"msg-1"},
		},
		{
			name: "multiple records: failed record does not block successful ones",
			records: []events.SQSMessage{
				sqsRecord("msg-ok", appointmentBody(sqsviewing.AppointmentCreated)),
				sqsRecord("msg-fail", appointmentBody(sqsviewing.AppointmentCreated)),
			},
			setupMock: func(m *mockScheduler) {
				m.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()
				m.On("Handle", mock.Anything, mock.Anything).Return(errors.New("publish failed")).Once()
			},
			wantFailureIDs: []string{"msg-fail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockScheduler{}
			tt.setupMock(m)

			h := New(newTestLogger(), m)
			resp, err := h.Handle(context.Background(), events.SQSEvent{Records: tt.records})

			require.NoError(t, err)

			failureIDs := make([]string, 0, len(resp.BatchItemFailures))
			for _, f := range resp.BatchItemFailures {
				failureIDs = append(failureIDs, f.ItemIdentifier)
			}
			if len(tt.wantFailureIDs) == 0 {
				assert.Empty(t, failureIDs)
			} else {
				assert.Equal(t, tt.wantFailureIDs, failureIDs)
			}
			m.AssertExpectations(t)
		})
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
