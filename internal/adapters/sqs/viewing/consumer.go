package viewing

import "time"

// AppointmentDetailType represents the detail-type values published by agent-service.
type AppointmentDetailType string

const (
	AppointmentCreated   AppointmentDetailType = "AppointmentCreated"
	AppointmentCancelled AppointmentDetailType = "AppointmentCancelled"
)

// AppointmentEvent is the full event envelope published by agent-service to EventBridge
// and delivered to this service via SQS.
type AppointmentEvent struct {
	DetailType AppointmentDetailType `json:"detailType"`
	Metadata   AppointmentMetadata   `json:"metadata"`
	Payload    Appointment           `json:"payload"`
}

type AppointmentMetadata struct {
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	SchemaVersion string    `json:"schemaVersion"`
	CorrelationID string    `json:"correlationId"`
}

type Appointment struct {
	AgentEmail string              `json:"agentEmail"`
	Location   string              `json:"location"`
	Start      time.Time           `json:"start"`
	End        time.Time           `json:"end"`
	Attendee   AppointmentAttendee `json:"attendee"`
}

type AppointmentAttendee struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
