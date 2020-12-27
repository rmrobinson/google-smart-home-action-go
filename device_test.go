package action

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceJSONSerializeDeserialize(t *testing.T) {
	origDevice := NewSimpleAVReceiver("test-id", []DeviceInput{
		{
			Key: "input-1",
			Names: []DeviceInputName{
				{
					LanguageCode: "en-US",
					Synonyms: []string{
						"First Input",
						"DVD Player",
					},
				},
			},
		},
		{
			Key: "input-2",
			Names: []DeviceInputName{
				{
					LanguageCode: "en-US",
					Synonyms: []string{
						"Second Input",
						"Computer",
					},
				},
			},
		},
	}, 100, true, false)
	origDevice.OtherDeviceIDs = append(origDevice.OtherDeviceIDs, OtherDeviceID{
		AgentID:  "agent-id-test",
		DeviceID: "test-id-other",
	})

	serializedBytes, serializeErr := json.Marshal(origDevice)
	assert.Nil(t, serializeErr)

	convDevice := Device{}
	deserializeErr := json.Unmarshal(serializedBytes, &convDevice)
	assert.Nil(t, deserializeErr)

	reserializedBytes, reserializedErr := json.Marshal(convDevice)
	assert.Nil(t, reserializedErr)
	assert.Equal(t, serializedBytes, reserializedBytes)
}
