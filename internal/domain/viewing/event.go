package viewing

import "time"

type DetailType string

const (
	DetailTypeScheduled  DetailType = "ViewingScheduled"
	DetailTypeCancelled  DetailType = "ViewingCancelled"
)

type EventMetadata struct {
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	SchemaVersion string    `json:"schemaVersion"`
	CorrelationID string    `json:"correlationId"`
}

type Attendee struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Viewing struct {
	AgentEmail string    `json:"agentEmail"`
	Address    string    `json:"address"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Attendee   Attendee  `json:"attendee"`
}

type Event struct {
	DetailType DetailType    `json:"detailType"`
	Metadata   EventMetadata `json:"metadata"`
	Payload    Viewing       `json:"payload"`
}
