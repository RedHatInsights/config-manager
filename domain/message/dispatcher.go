package message

// DispatcherEvent represents a message read off the playbook-dispatcher.runs
// topic.
type DispatcherEvent struct {
	Type    string                 `json:"event_type"`
	Payload DispatcherEventPayload `json:"payload"`
}

// DispatcherEventPayload represents the payload field of the
// DispatcherEvent.
type DispatcherEventPayload struct {
	ID            string            `json:"id"`
	OrgID         string            `json:"org_id"`
	Recipient     string            `json:"recipient"`
	CorrelationID string            `json:"correlation_id"`
	Service       string            `json:"service"`
	URL           string            `json:"url"`
	Labels        map[string]string `json:"labels"`
	Status        string            `json:"status"`
}
