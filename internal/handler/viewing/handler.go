package viewing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	sqsviewing "github.com/troysnowden/viewing-service/internal/adapters/sqs/viewing"
	appviewing "github.com/troysnowden/viewing-service/internal/application/viewing"
	domain "github.com/troysnowden/viewing-service/internal/domain/viewing"
)

type scheduler interface {
	Handle(ctx context.Context, cmd appviewing.Command) error
}

type Handler struct {
	logger    *slog.Logger
	scheduler scheduler
}

func New(logger *slog.Logger, scheduler scheduler) *Handler {
	return &Handler{logger: logger, scheduler: scheduler}
}

func (h *Handler) Handle(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	var resp events.SQSEventResponse

	for _, record := range event.Records {
		if err := h.processRecord(ctx, record); err != nil {
			h.logger.ErrorContext(ctx, "failed to process record",
				"messageId", record.MessageId,
				"error", err,
			)
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
		}
	}

	return resp, nil
}

func (h *Handler) processRecord(ctx context.Context, record events.SQSMessage) error {
	var appt sqsviewing.AppointmentEvent
	if err := json.Unmarshal([]byte(record.Body), &appt); err != nil {
		return fmt.Errorf("unmarshal appointment event: %w", err)
	}

	detailType, err := mapDetailType(appt.DetailType)
	if err != nil {
		return err
	}

	cmd := appviewing.Command{
		CorrelationID: appt.Metadata.CorrelationID,
		SchemaVersion: appt.Metadata.SchemaVersion,
		DetailType:    detailType,
		AgentEmail:    appt.Payload.AgentEmail,
		Address:       appt.Payload.Location,
		Start:         appt.Payload.Start,
		End:           appt.Payload.End,
		Attendee: domain.Attendee{
			Name:  appt.Payload.Attendee.Name,
			Email: appt.Payload.Attendee.Email,
			Phone: appt.Payload.Attendee.Phone,
		},
	}

	return h.scheduler.Handle(ctx, cmd)
}

func mapDetailType(t sqsviewing.AppointmentDetailType) (domain.DetailType, error) {
	switch t {
	case sqsviewing.AppointmentCreated:
		return domain.DetailTypeScheduled, nil
	case sqsviewing.AppointmentCancelled:
		return domain.DetailTypeCancelled, nil
	default:
		return "", fmt.Errorf("unknown appointment detail type: %q", t)
	}
}
