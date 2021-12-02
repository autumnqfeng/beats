package add_stark_metadata

import (
	"encoding/json"
	"strings"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
)

const (
	processorName = "add_stark_metadata"
)

func init() {
	processors.RegisterPlugin(processorName, newStarkMetadataProcessor)
}

type addStarkMetadata struct {
	log *logp.Logger
}

func newStarkMetadataProcessor(cfg *common.Config) (processors.Processor, error) {
	return &addStarkMetadata{logp.NewLogger(processorName)}, nil
}

func (d *addStarkMetadata) Run(event *beat.Event) (*beat.Event, error) {

	var to interface{}
	if message, err := event.Fields.GetValue("message"); err == nil {
		if text, ok := message.(string); ok {
			dec := json.NewDecoder(strings.NewReader(text))
			dec.UseNumber()
			err := dec.Decode(&to)
			if err == nil {
				switch to.(type) {
				case map[string]interface{}:
					for key, value := range to.(map[string]interface{}) {
						event.Fields[key] = value
					}
				}
			}
		}
	}
	if appFields, err := event.Fields.GetValue("fields.app_field"); err == nil {
		fields := strings.Split(appFields.(string), ",")
		for _, val := range fields {
			values := strings.Split(val, "=")
			if values[0] == "IDC" {
				event.Fields["cluster"] = strings.ToLower(values[1])
			}
			if values[0] == "deployment" {
				event.Fields["deployment"] = values[1]
			}
			if values[0] == "pod" {
				event.Fields["pod"] = values[1]
			}
			if values[0] == "namespace" {
				event.Fields["namespace"] = values[1]
			}

			if values[0] == "container" {
				event.Fields["container_name"] = values[1]
			}
		}
	}
	return event, nil
}

func (d *addStarkMetadata) String() string {
	return processorName
}
