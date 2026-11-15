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

// TestRecordInputUsesCurrentInputKey guards against a real bug: Google's
// InputSelector trait requires the reportable state field to be named
// "currentInput" (see RecordInput's doc comment), not "input" - the JSON
// round-trip test above wouldn't catch a wrong-but-consistent key name,
// since it only checks serialize/deserialize symmetry against itself.
func TestRecordInputUsesCurrentInputKey(t *testing.T) {
	ds := NewDeviceState(true)
	ds.RecordInput("hdmi1")

	assert.Equal(t, "hdmi1", ds.State["currentInput"])
	_, hasWrongKey := ds.State["input"]
	assert.False(t, hasWrongKey, `RecordInput must not set "input" - Google's API expects "currentInput"`)
}
