package logger

type logMessage struct {
	Timestamp   string `json:"timestamp"`
	MessageType string `json:"type"`
	Producer    string `json:"producer"`
	Level       string `json:"level"`
	Message     string `json:"message"`
}
