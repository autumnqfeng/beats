package add_so_metadata

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
)

const (
	processorName                          = "add_so_metadata"
	keyKubernetesPodName                   = "kubernetes.pod.name"
	keyKubernetesNamespace                 = "kubernetes.namespace"
	keyKubernetesContainerName             = "kubernetes.container.name"
	keyKubernetesAnnotationsControllerKind = "kubernetes.annotations.qihoo.cloud/controller-kind"
	keyKubernetesLabelAppName              = "kubernetes.labels.app"
)

func init() {
	processors.RegisterPlugin(processorName, newQihooMetadataProcessor)
}

type addQihooMetadata struct {
	log *logp.Logger
}

func newQihooMetadataProcessor(cfg *common.Config) (processors.Processor, error) {
	return &addQihooMetadata{logp.NewLogger(processorName)}, nil
}

func (d *addQihooMetadata) Run(event *beat.Event) (*beat.Event, error) {
	defaultControllerKind := "deployment"
	defaultTopicInfix := "docker"
	defaultDeploymentName := "none"
	defaultAppName := "none"

	kubernetesAnnotationsControllerKind, err := event.Fields.GetValue(keyKubernetesAnnotationsControllerKind)
	if err != nil {
		d.log.Debugf("Error while get %s fields. %s ,err %v", keyKubernetesAnnotationsControllerKind, event.Fields.String(), err)
	} else {
		defaultControllerKind = strings.ToLower(kubernetesAnnotationsControllerKind.(string))
	}

	if defaultControllerKind != "deployment" {
		defaultTopicInfix = defaultControllerKind
	}
	event.Fields["controller_kind"] = defaultControllerKind

	kubernetesLabelAppName, err := event.Fields.GetValue(keyKubernetesLabelAppName)
	if err != nil {
		d.log.Debugf("Error while get %s fields. %s ,err %v", keyKubernetesLabelAppName, event.Fields.String(), err)
	} else {
		defaultAppName = kubernetesLabelAppName.(string)
		if defaultControllerKind == "deployment" {
			defaultDeploymentName = defaultAppName
		}
	}

	// 为了兼容之前的deployment 和topic字段
	// 非 deployment 或者无法取到app label的deployment为none
	// topic 如果类型为deployment 则为 k8s_docker_{{deploymentName}}, 其他类型 则为 k8s_{{controllerKind}}_{{appName}}
	event.Fields["deployment"] = defaultDeploymentName
	event.Fields["topic"] = fmt.Sprintf("k8s_%s_%s", defaultTopicInfix, defaultAppName)
	event.Fields["app"] = defaultAppName

	kubernetesPodName, err := event.Fields.GetValue(keyKubernetesPodName)
	if err != nil {
		d.log.Debugf("Error while get %s fields. %s", keyKubernetesPodName, event.Fields.String())
	} else {
		event.Fields["pod"] = kubernetesPodName
	}

	kubernetesNamespace, err := event.Fields.GetValue(keyKubernetesNamespace)
	if err != nil {
		d.log.Debugf("Error while get %s fields. %s ", keyKubernetesNamespace, event.Fields.String())
	} else {
		event.Fields["namespace"] = kubernetesNamespace
	}

	kubernetesContainerName, err := event.Fields.GetValue(keyKubernetesContainerName)
	if err != nil {
		d.log.Debugf("Error while get %s fields. %s", keyKubernetesContainerName, event.Fields.String())
	} else {
		event.Fields["container_name"] = kubernetesContainerName
	}

	if value, err := event.Fields.GetValue("log.file.path"); err == nil {
		event.Fields["source"] = value
	}

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

	if topic, err := event.Fields.GetValue("fields.log_topic"); err == nil {
		event.Fields["topic"] = topic
	}

	if appFields, err := event.Fields.GetValue("fields.app_field"); err == nil {
		fields := strings.Split(appFields.(string), ",")
		for _, val := range fields {
			values := strings.Split(val, "=")
			if values[0] == "IDC" {
				event.Fields["cluster"] = strings.ToLower(values[1])
			}
			//if values[0] == "AppName" {
			//	event.Fields["deployment"] = values[1]
			//}
			//if values[0] == "PodName" {
			//	event.Fields["pod"] = values[1]
			//}
		}
	}

	return event, nil
}

func (d *addQihooMetadata) String() string {
	return processorName
}
