package action

import (
	"encoding/json"
	"fmt"
)

// Command defines which command, and what details, are being specified.
// Only one of the contained fields will be set at any point in time.
type Command struct {
	Name      string
	Challenge map[string]interface{}

	Generic *CommandGeneric

	BrightnessAbsolute            *CommandBrightnessAbsolute
	BrightnessRelative            *CommandBrightnessRelative
	ColorAbsolute                 *CommandColorAbsolute
	OnOff                         *CommandOnOff
	Mute                          *CommandMute
	SetVolume                     *CommandSetVolume
	AdjustVolume                  *CommandSetVolumeRelative
	SetInput                      *CommandSetInput
	NextInput                     *CommandNextInput
	PreviousInput                 *CommandPreviousInput
	SetFanSpeed                   *CommandSetFanSpeed
	ThermostatTemperatureSetpoint *CommandThermostatTemperatureSetpoint
	OpenClose                     *CommandOpenClose
	OpenCloseRelative             *CommandOpenCloseRelative
}

// MarshalJSON is a custom JSON serializer for our Command
func (c Command) MarshalJSON() ([]byte, error) {
	var details interface{}

	switch c.Name {
	case "action.devices.commands.BrightnessAbsolute":
		details = c.BrightnessAbsolute
	case "action.devices.commands.BrightnessRelative":
		details = c.BrightnessRelative
	case "action.devices.commands.ColorAbsolute":
		details = c.ColorAbsolute
	case "action.devices.commands.OnOff":
		details = c.OnOff
	case "action.devices.commands.mute":
		details = c.Mute
	case "action.devices.commands.setVolume":
		details = c.SetVolume
	case "action.devices.commands.volumeRelative":
		details = c.AdjustVolume
	case "action.devices.commands.SetInput":
		details = c.SetInput
	case "action.devices.commands.SetFanSpeed":
		details = c.SetFanSpeed
	case "action.devices.commands.ThermostatTemperatureSetpoint":
		details = c.ThermostatTemperatureSetpoint
	case "action.devices.commands.NextInput":
		details = c.NextInput
	case "action.devices.commands.PreviousInput":
		details = c.PreviousInput
	case "action.devices.commands.OpenClose":
		details = c.OpenClose
	case "action.devices.commands.OpenCloseRelative":
		details = c.OpenCloseRelative
	default:
		return json.Marshal(c.Generic)
	}

	var tmp struct {
		Command   string                 `json:"command"`
		Params    interface{}            `json:"params"`
		Challenge map[string]interface{} `json:"challenge,omitempty"`
	}
	tmp.Command = c.Name
	tmp.Params = details
	if c.Challenge != nil {
		tmp.Challenge = c.Challenge
	}
	return json.Marshal(tmp)
}

// UnmarshalJSON is a custom JSON deserializer for our Command
func (c *Command) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Command   string                 `json:"command"`
		Params    json.RawMessage        `json:"params"`
		Challenge map[string]interface{} `json:"challenge,omitempty"`
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	c.Name = tmp.Command
	if tmp.Challenge != nil {
		c.Challenge = tmp.Challenge
	}

	var details interface{}
	switch tmp.Command {
	case "action.devices.commands.BrightnessAbsolute":
		c.BrightnessAbsolute = &CommandBrightnessAbsolute{}
		details = c.BrightnessAbsolute
	case "action.devices.commands.BrightnessRelative":
		c.BrightnessRelative = &CommandBrightnessRelative{}
		details = c.BrightnessRelative
	case "action.devices.commands.ColorAbsolute":
		c.ColorAbsolute = &CommandColorAbsolute{}
		details = c.ColorAbsolute
	case "action.devices.commands.OnOff":
		c.OnOff = &CommandOnOff{}
		details = c.OnOff
	case "action.devices.commands.mute":
		c.Mute = &CommandMute{}
		details = c.Mute
	case "action.devices.commands.setVolume":
		c.SetVolume = &CommandSetVolume{}
		details = c.SetVolume
	case "action.devices.commands.volumeRelative":
		c.AdjustVolume = &CommandSetVolumeRelative{}
		details = c.AdjustVolume
	case "action.devices.commands.SetInput":
		c.SetInput = &CommandSetInput{}
		details = c.SetInput
	case "action.devices.commands.SetFanSpeed":
		c.SetFanSpeed = &CommandSetFanSpeed{}
		details = c.SetFanSpeed
	case "action.devices.commands.ThermostatTemperatureSetpoint":
		c.ThermostatTemperatureSetpoint = &CommandThermostatTemperatureSetpoint{}
		details = c.ThermostatTemperatureSetpoint
	case "action.devices.commands.NextInput":
		c.NextInput = &CommandNextInput{}
		details = c.NextInput
	case "action.devices.commands.PreviousInput":
		c.PreviousInput = &CommandPreviousInput{}
		details = c.PreviousInput
	case "action.devices.commands.OpenClose":
		c.OpenClose = &CommandOpenClose{}
		details = c.OpenClose
	case "action.devices.commands.OpenCloseRelative":
		c.OpenCloseRelative = &CommandOpenCloseRelative{}
		details = c.OpenCloseRelative
	default:
		c.Generic = &CommandGeneric{}
		err := json.Unmarshal(data, c.Generic)
		if err != nil {
			return err
		}
		return nil
	}
	paramsRawJson := tmp.Params
	if len(paramsRawJson) == 0 {
		paramsRawJson = []byte("null")
	}

	err = json.Unmarshal(paramsRawJson, details)
	if err != nil {
		return fmt.Errorf("error unmarshaling command JSON 'params' value, %s, into details: %w", string(paramsRawJson), err)
	}

	return nil
}

// CommandGeneric contains a command definition which hasn't been parsed into a specific command structure.
// This is intended to support newly defined commands which callers of this SDK may handle but this does not yet support.
type CommandGeneric struct {
	Command   string                 `json:"command"`
	Params    map[string]interface{} `json:"params"`
	Challenge map[string]string      `json:"challenge,omitempty"`
}

// CommandBrightnessAbsolute requests to set the brightness to an absolute value
// See https://developers.google.com/assistant/smarthome/traits/brightness
type CommandBrightnessAbsolute struct {
	Brightness int `json:"brightness"`
}

// CommandBrightnessRelative requests to set the brightness to a relative level
// Only one of the two fields will be set.
// See https://developers.google.com/assistant/smarthome/traits/brightness
type CommandBrightnessRelative struct {
	RelativePercent int `json:"brightnessRelativePercent"`
	RelativeWeight  int `json:"brightnessRelativeWeight"`
}

// CommandColorAbsolute requests to set the colour of a light to a particular value.
// Only one of temperature, RGB and HSV will be set.
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
type CommandColorAbsolute struct {
	Color struct {
		Name        string `json:"name"`
		Temperature int    `json:"temperature"`
		RGB         int    `json:"spectrumRGB"`
		HSV         struct {
			Hue        float64 `json:"hue"`
			Saturation float64 `json:"saturation"`
			Value      float64 `json:"value"`
		} `json:"spectrumHSV"`
	} `json:"color"`
}

// CommandOnOff requests to turn the entity on or off.
// See https://developers.google.com/assistant/smarthome/traits/onoff
type CommandOnOff struct {
	On bool `json:"on"`
}

// CommandMute requests the device be muted.
// See https://developers.google.com/assistant/smarthome/traits/volume
type CommandMute struct {
	Mute bool `json:"mute"`
}

// CommandSetFanSpeed requests the device fan speed be set to the specified
// value.
//
// See https://developers.google.com/assistant/smarthome/traits/fanspeed
type CommandSetFanSpeed struct {
	// The requested speed setting of the fan, which corresponds to one of the
	// settings specified during the sync response.
	FanSpeed *string `json:"fanSpeed,omitempty"`

	// The requested speed setting percentage (0-100).
	FanSpeedPercent *float32 `json:"fanSpeedPercent,omitempty"`
}

// CommandThermostatTemperatureSetpoint requests the target temperature be set
// to a particular temperature.
//
// [ThermostatTemperatureSetpoint documentation by Google]: https://developers.home.google.com/cloud-to-cloud/traits/temperaturesetting#action.devices.commands.thermostattemperaturesetpoint
type CommandThermostatTemperatureSetpoint struct {
	// Target temperature setpoint. Supports up to one decimal place.
	ThermostatTemperatureSetpointCelcius float32 `json:"thermostatTemperatureSetpoint"`
}

// CommandSetVolume requests the device volume be set to the specified value.
// See https://developers.google.com/assistant/smarthome/traits/volume
type CommandSetVolume struct {
	Level int `json:"volumeLevel"`
}

// CommandSetVolumeRelative requests the device volume be increased or decreased.
// See https://developers.google.com/assistant/smarthome/traits/volume
type CommandSetVolumeRelative struct {
	Amount int `json:"relativeSteps"`
}

// CommandSetInput requests the device input be changed.
// See https://developers.google.com/assistant/smarthome/traits/inputselector
type CommandSetInput struct {
	NewInput string `json:"newInput"`
}

// CommandNextInput requests the device input be changed to the next logical one.
// See https://developers.google.com/assistant/smarthome/traits/inputselector
type CommandNextInput struct {
}

// CommandPreviousInput requests the device input be changed to the previous logical one.
// See https://developers.google.com/assistant/smarthome/traits/inputselector
type CommandPreviousInput struct {
}

// CommandOpenClose requests to open the device to the specified value.
// See https://developers.google.com/assistant/smarthome/traits/openclose
type CommandOpenClose struct {
	OpenPercent int     `json:"openPercent"`
	Direction   *string `json:"openDirection,omitempty"`
}

// CommandOpenCloseRelative requests to adjust the open-close state of the device relative to the current state.
// This command is only available if commandOnlyOpenClose is set to false.
// See https://developers.google.com/assistant/smarthome/traits/openclose
type CommandOpenCloseRelative struct {
	OpenRelativePercent int     `json:"openRelativePercent"`
	Direction           *string `json:"openDirection,omitempty"`
}
