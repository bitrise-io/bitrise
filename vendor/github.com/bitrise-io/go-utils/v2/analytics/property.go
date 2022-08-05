package analytics

import (
	"github.com/mitchellh/mapstructure"
)

// Properties ...
type Properties map[string]interface{}

// Merge ...
func (p Properties) Merge(properties Properties) Properties {
	r := Properties{}
	for key, value := range p {
		r[key] = value
	}
	for key, value := range properties {
		r[key] = value
	}
	return r
}

// AppendIfNotEmpty ...
func (p Properties) AppendIfNotEmpty(key string, value string) {
	if value != "" {
		p[key] = value
	}
}

// NewProperty Converts struct into Properties map based on json tags
func NewProperty(item interface{}) Properties {
	result := Properties{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &result,
	})
	if err != nil {
		return Properties{}
	}
	if err := decoder.Decode(item); err != nil {
		return Properties{}
	}
	return result
}
