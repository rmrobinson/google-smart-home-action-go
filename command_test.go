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

func TestCommandColorAbsoluteDeserializer(t *testing.T) {
	msg := `{
		"command": "action.devices.commands.ColorAbsolute",
		"params": {
		  "color": {
			"name": "magenta",
			"spectrumHSV": {
			  "hue": 300,
			  "saturation": 1,
			  "value": 1
			}
		  }
		}
	  }`

	cmd := Command{}
	err := json.Unmarshal([]byte(msg), &cmd)
	assert.Nil(t, err)
	assert.NotNil(t, cmd.ColorAbsolute)
	assert.Equal(t, 300.0, cmd.ColorAbsolute.Color.HSV.Hue)
	assert.Equal(t, 1.0, cmd.ColorAbsolute.Color.HSV.Saturation)
	assert.Equal(t, 1.0, cmd.ColorAbsolute.Color.HSV.Value)
}
