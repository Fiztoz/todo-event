package messaging

import "encoding/json"

const TaskSubject = "task.events"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
