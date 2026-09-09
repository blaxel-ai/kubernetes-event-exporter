package sinks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/blaxel-ai/kubernetes-event-exporter/pkg/kube"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRawJSONTemplatePreservesInvolvedObjectLabels(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Name = "job-render-summary-ws123-exec-abc-123-7f9c4"
	ev.InvolvedObject.Namespace = "test-my-workspace"
	ev.InvolvedObject.Labels = map[string]string{
		"workspace":    "my-workspace",
		"job":          "render-summary",
		"execution-id": "exec-abc-123",
		"escaped":      `quote " slash \\ newline \n`,
	}

	layout := map[string]interface{}{
		"data": map[interface{}]interface{}{
			"involvedObject": map[interface{}]interface{}{
				"kind":   "{{ .InvolvedObject.Kind }}",
				"name":   "{{ .InvolvedObject.Name }}",
				"labels": "{{ rawJson .InvolvedObject.Labels }}",
			},
		},
	}

	res, err := convertLayoutTemplate(layout, ev)
	require.NoError(t, err)

	data := res["data"].(map[string]interface{})
	involvedObject := data["involvedObject"].(map[string]interface{})
	labels, ok := involvedObject["labels"].(map[string]interface{})
	require.True(t, ok, "labels must remain a JSON object, not a JSON-encoded string")
	require.Equal(t, "my-workspace", labels["workspace"])
	require.Equal(t, "render-summary", labels["job"])
	require.Equal(t, "exec-abc-123", labels["execution-id"])
	require.Equal(t, ev.InvolvedObject.Labels["escaped"], labels["escaped"])

	payload, err := json.Marshal(res)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"data": {
			"involvedObject": {
				"kind": "Pod",
				"name": "job-render-summary-ws123-exec-abc-123-7f9c4",
				"labels": {
					"workspace": "my-workspace",
					"job": "render-summary",
					"execution-id": "exec-abc-123",
					"escaped": "quote \" slash \\\\ newline \\n"
				}
			}
		}
	}`, string(payload))
}

func TestRawJSONTemplateAllowsSurroundingWhitespace(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{"execution-id": "exec-with-hyphens"}

	res, err := convertLayoutTemplate(map[string]interface{}{
		"labels": "  {{ rawJson .InvolvedObject.Labels }}\n",
	}, ev)
	require.NoError(t, err)

	labels, ok := res["labels"].(map[string]interface{})
	require.True(t, ok, "standalone rawJson with surrounding whitespace must remain a JSON object")
	require.Equal(t, "exec-with-hyphens", labels["execution-id"])
}

func TestMixedRawJSONTemplateRemainsString(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.InvolvedObject.Labels = map[string]string{"execution-id": "exec-with-hyphens"}

	res, err := convertLayoutTemplate(map[string]interface{}{
		"message": "labels={{ rawJson .InvolvedObject.Labels }}",
	}, ev)
	require.NoError(t, err)
	require.IsType(t, "", res["message"])
	require.Contains(t, res["message"], rawJSONTemplatePrefix)
}

func TestRawJSONTemplatePreservesNilInvolvedObjectLabels(t *testing.T) {
	ev := &kube.EnhancedEvent{}

	res, err := convertLayoutTemplate(map[string]interface{}{
		"involvedObject": map[interface{}]interface{}{
			"labels": "{{ rawJson .InvolvedObject.Labels }}",
		},
	}, ev)
	require.NoError(t, err)

	involvedObject := res["involvedObject"].(map[string]interface{})
	require.Contains(t, involvedObject, "labels")
	require.Nil(t, involvedObject["labels"])

	payload, err := json.Marshal(res)
	require.NoError(t, err)
	require.JSONEq(t, `{"involvedObject":{"labels":null}}`, string(payload))
}

func TestStaticRawJSONPrefixStringRemainsString(t *testing.T) {
	res, err := convertLayoutTemplate(map[string]interface{}{
		"message": rawJSONTemplatePrefix + `{"not":"a-template"}`,
	}, &kube.EnhancedEvent{})
	require.NoError(t, err)
	require.Equal(t, rawJSONTemplatePrefix+`{"not":"a-template"}`, res["message"])
}

func TestTemplateJSONLookingStringsRemainStrings(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Count = 12

	res, err := convertLayoutTemplate(map[string]interface{}{
		"count": "{{ .Count }}",
	}, ev)
	require.NoError(t, err)
	require.Equal(t, "12", res["count"])
}

func TestLayoutConvert(t *testing.T) {
	ev := &kube.EnhancedEvent{}
	ev.Namespace = "default"
	ev.Type = "Warning"
	ev.InvolvedObject.Kind = "Pod"
	ev.InvolvedObject.Name = "nginx-server-123abc-456def"
	ev.Message = "Successfully pulled image \"nginx:latest\""
	ev.FirstTimestamp = v1.Time{Time: time.Now()}

	// Because Go, when parsing yaml, its []interface, not []string
	var tagz interface{} = make([]interface{}, 2)
	tagz.([]interface{})[0] = "sre"
	tagz.([]interface{})[1] = "ops"

	layout := map[string]interface{}{
		"detail": map[interface{}]interface{}{
			"message":   "{{ .Message }}",
			"kind":      "{{ .InvolvedObject.Kind }}",
			"name":      "{{ .InvolvedObject.Name }}",
			"namespace": "{{ .Namespace }}",
			"type":      "{{ .Type }}",
			"tags":      tagz,
		},
		"eventType": "kube-event",
		"region":    "us-west-2",
		"createdAt": "{{ .GetTimestampMs }}", // TODO: Test Int casts
	}

	res, err := convertLayoutTemplate(layout, ev)
	require.NoError(t, err)
	require.Equal(t, res["eventType"], "kube-event")

	val, ok := res["detail"].(map[string]interface{})

	require.True(t, ok, "cannot cast to event")

	val2, ok2 := val["message"].(string)
	require.True(t, ok2, "cannot cast message to string")

	require.Equal(t, val2, ev.Message)
}
