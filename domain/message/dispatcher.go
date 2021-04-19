package message

type DispatcherEvent struct {
	Type    string                 `json:"event_type"`
	Payload DispatcherEventPayload `json:"payload"`
}

type DispatcherEventPayload struct {
	ID            string            `json:"id"`
	Account       string            `json:"account"`
	Recipient     string            `json:"recipient"`
	CorrelationID string            `json:"correlation_id"`
	Service       string            `json:"service"`
	URL           string            `json:"url"`
	Labels        map[string]string `json:"labels"`
	Status        string            `json:"status"`
}
