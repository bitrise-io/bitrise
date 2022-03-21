package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gofrs/uuid"
)

type event struct {
	ID         string     `json:"id"`
	EventName  string     `json:"event_name"`
	Timestamp  int64      `json:"timestamp"`
	Properties Properties `json:"properties"`
}

func newEvent(name string, properties []Properties) event {
	return event{
		ID:         uuid.Must(uuid.NewV4()).String(),
		EventName:  name,
		Timestamp:  time.Now().UnixNano() / int64(time.Microsecond),
		Properties: merge(properties),
	}
}

func (e event) toJSON(writer io.Writer) {
	if err := json.NewEncoder(writer).Encode(e); err != nil {
		panic(fmt.Sprintf("Analytics event should be serializable to JSON: %s", err.Error()))
	}
}

func merge(properties []Properties) Properties {
	if len(properties) == 0 {
		return nil
	}
	m := map[string]interface{}{}
	for _, p := range properties {
		for key, value := range p {
			m[key] = value
		}
	}
	return m
}
