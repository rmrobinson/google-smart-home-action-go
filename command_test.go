package action

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestCommandUnmarshalJSON(t *testing.T) {
	for _, example := range []struct {
		name    string
		input   string
		want    *Command
		wantErr bool
	}{
		{
			name: "thermostat command - typical",
			input: `{
				"command":"action.devices.commands.ThermostatTemperatureSetpoint",
				"params":{
				   "thermostatTemperatureSetpoint": 42.42
				}
			 }`,
			want: &Command{
				Name: "action.devices.commands.ThermostatTemperatureSetpoint",
				Generic: &CommandGeneric{
					Command: "action.devices.commands.ThermostatTemperatureSetpoint",
					Params: map[string]interface{}{
						"thermostatTemperatureSetpoint": 42.42,
					},
				},
			},
		},
		{
			name: "thermostat command - empty params object",
			input: `{
				"command":"action.devices.commands.ThermostatTemperatureSetpoint",
				"params":{}
			 }`,
			want: &Command{
				Name: "action.devices.commands.ThermostatTemperatureSetpoint",
				Generic: &CommandGeneric{
					Command: "action.devices.commands.ThermostatTemperatureSetpoint",
					Params:  map[string]interface{}{},
				},
			},
		},
		{
			name:  "thermostat command - missing params object",
			input: `{"command":"action.devices.commands.ThermostatTemperatureSetpoint"}`,
			want: &Command{
				Name: "action.devices.commands.ThermostatTemperatureSetpoint",
				Generic: &CommandGeneric{
					Command: "action.devices.commands.ThermostatTemperatureSetpoint",
					Params:  nil,
				},
			},
		},
	} {
		t.Run(example.name, func(t *testing.T) {
			got := &Command{}
			err := json.Unmarshal([]byte(example.input), got)
			if gotErr := err != nil; gotErr != example.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, example.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(example.want, got); diff != "" {
				t.Errorf("unexpected diff in parsed command (-want, +got):\n  %s", diff)
			}

			roundtripResult := &Command{}
			if err := roundtripJSON(example.want, roundtripResult); err != nil {
				t.Fatalf("error encoding and decoding JSON: %v", err)
			}
			if diff := cmp.Diff(example.want, roundtripResult); diff != "" {
				t.Errorf("unexpected diff in roundtrip result (-original, +roundtrip result):\n  %s", diff)
			}
		})
	}
}

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

func roundtripJSON(in, out interface{}) error {
	data, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("error marshaling: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("error unmarshaling: %w", err)
	}
	return nil
}
