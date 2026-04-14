package viewing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	domain "github.com/troysnowden/viewing-service/internal/domain/viewing"
)

type eventbridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Publisher struct {
	eventbridge eventbridgeClient
	busName     string
}

func New(eb *eventbridge.Client, busName string) *Publisher {
	return &Publisher{eventbridge: eb, busName: busName}
}

func (p *Publisher) Publish(ctx context.Context, event domain.Event) error {
	detail, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	out, err := p.eventbridge.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Detail:       aws.String(string(detail)),
				DetailType:   aws.String(string(event.DetailType)),
				EventBusName: aws.String(p.busName),
				Source:       aws.String(event.Metadata.Source),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("put events: %w", err)
	}
	if out.FailedEntryCount > 0 {
		return fmt.Errorf("eventbridge: event failed to publish")
	}

	return nil
}
