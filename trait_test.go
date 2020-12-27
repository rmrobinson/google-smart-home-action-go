package action

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceStateJSONSerializeDeserialize(t *testing.T) {
	origState := NewDeviceState(true)
	origState.RecordBrightness(50)
	origState.RecordColorHSV(100, 100, 100)
	origState.RecordOnOff(true)
	origState.Status = "ONLINE"

	serializedBytes, serializeErr := json.Marshal(origState)
	assert.Nil(t, serializeErr)

	convState := DeviceState{}
	deserializeErr := json.Unmarshal(serializedBytes, &convState)
	assert.Nil(t, deserializeErr)

	reserializedBytes, reserializedErr := json.Marshal(convState)
	assert.Nil(t, reserializedErr)
	assert.Equal(t, serializedBytes, reserializedBytes)
}
