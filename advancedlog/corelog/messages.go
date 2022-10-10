package corelog

type messageType string

const (
	logMessageType messageType = "log"
)

type logMessage struct {
	Timestamp   string `json:"timestamp"`
	MessageType string `json:"type"`
	Producer    string `json:"producer"`
	ProducerID  string `json:"producer_id,omitempty"`
	Level       string `json:"level"`
	Message     string `json:"message"`
}
