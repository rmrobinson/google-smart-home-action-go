package action

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandJSONSerializeDeserialize(t *testing.T) {
	origCommand := Command{
		Name: "action.devices.commands.OnOff",
		OnOff: &CommandOnOff{
			On: true,
		},
	}

	serializedBytes, serializeErr := json.Marshal(origCommand)
	assert.Nil(t, serializeErr)

	convCommand := Command{}
	deserializeErr := json.Unmarshal(serializedBytes, &convCommand)
	assert.Nil(t, deserializeErr)
	assert.NotNil(t, convCommand.OnOff)
	assert.True(t, convCommand.OnOff.On)
	assert.Equal(t, "action.devices.commands.OnOff", convCommand.Name)

	reserializedBytes, reserializedErr := json.Marshal(convCommand)
	assert.Nil(t, reserializedErr)
	assert.Equal(t, serializedBytes, reserializedBytes)
}
