package action

// DeviceState contains the state of a device.
type DeviceState struct {
	Online bool
	Status string

	state map[string]interface{}
}

// RecordBrightness adds the current brightness to the device.
// Should only be applied to devices with the Brightness trait
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (ds *DeviceState) RecordBrightness(brightness int) *DeviceState {
	ds.state["brightness"] = brightness
	return ds
}

// RecordColorTemperature adds the current color temperature (in Kelvin) to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds *DeviceState) RecordColorTemperature(temperatureK int) *DeviceState {
	ds.state["color"] = map[string]interface{}{
		"temperatureK": temperatureK,
	}
	return ds
}

// RecordColorRGB adds the current color in RGB to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds *DeviceState) RecordColorRGB(spectrumRgb int) *DeviceState {
	ds.state["color"] = map[string]interface{}{
		"spectrumRgb": spectrumRgb,
	}
	return ds
}

// RecordColorHSV adds the current color in HSV to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds *DeviceState) RecordColorHSV(hue float64, saturation float64, value float64) *DeviceState {
	ds.state["color"] = map[string]interface{}{
		"spectrumHsv": map[string]interface{}{
			"hue":        hue,
			"saturation": saturation,
			"value":      value,
		},
	}
	return ds
}

// RecordInput adds the current input active to the device.
// Should only be applied to devices with the InputSelector trait
// See https://developers.google.com/assistant/smarthome/traits/inputselector
func (ds *DeviceState) RecordInput(input string) *DeviceState {
	ds.state["input"] = input
	return ds
}

// RecordOnOff adds the current on/off state to the device.
// Should only be applied to devices with the OnOff trait
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (ds *DeviceState) RecordOnOff(on int) *DeviceState {
	ds.state["on"] = on
	return ds
}

// AddBrightnessTrait indicates this device is capable of having its brightness controlled.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only switch).
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (d *Device) AddBrightnessTrait(onlyCommand bool) *Device {
	d.Traits["action.devices.traits.Brightness"] = true
	if onlyCommand {
		d.attributes["commandOnlyBrightness"] = true
	}

	return d
}

// ColorModel defines which model of the color wheel the device supports.
const (
	RGB = "rgb"
	HSV = "hsv"
)

// AddColourTrait indicates this device is capable of having its colour display controlled using the specified color model.
// It is mutually exclusive to support RGB or HSV.
// It is possible to support either one of RGB and HSV alongside color temperature. See AddColorTemperatureSetting
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourTrait(model string, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.ColorSetting"] = true
	if onlyCommand {
		d.attributes["commandOnlyColorSetting"] = true
	}
	d.attributes["colorModel"] = model

	return d
}

// AddColourTemperatureTrait indicates this device is capable of having its colour display controlled using the colour temperature model.
// This can be set alongside AddColorSetting to indicate both color temperature and another algorithm are supported.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourTemperatureTrait(minTempK int, maxTempK int, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.ColorSetting"] = true

	if onlyCommand {
		d.attributes["commandOnlyColorSetting"] = true
	} else {
		d.attributes["commandOnlyColorSetting"] = false
	}
	d.attributes["colorTemperatureRange"] = map[string]int{
		"temperatureMinK": minTempK,
		"temperatureMaxK": maxTempK,
	}

	return d
}

// AddInputSelectorTrait indicates this device is capable of having its input selected.
// See https://developers.google.com/assistant/smarthome/traits/inputselector
func (d *Device) AddInputSelectorTrait(availableInputs []DeviceInput, ordered bool) *Device {
	d.Traits["action.devices.traits.InputSelector"] = true
	d.attributes["availableInputs"] = availableInputs
	d.attributes["orderedInputs"] = ordered

	return d
}

// AddOnOffTrait indicates this device is capable of having its state toggled on or off.
// If the device can be commanded but not queried, set onlyCommand to true (i.e. a write-only switch).
// If the devie cannot be commanded but only queried, set onlyQuery to true (i.e. a sensor).
// See https://developers.google.com/assistant/smarthome/traits/onoff
func (d *Device) AddOnOffTrait(onlyCommand, onlyQuery bool) *Device {
	d.Traits["action.devices.traits.OnOff"] = true
	if onlyCommand {
		d.attributes["commandOnlyOnOff"] = true
	}
	if onlyQuery {
		d.attributes["queryOnlyOnOff"] = true
	}

	return d
}

// AddVolumeTrait indicates this device is capable of having its volume controlled
// See https://developers.google.com/assistant/smarthome/traits/volume
func (d *Device) AddVolumeTrait(maxLevel int, canMute bool, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.Volume"] = true
	if onlyCommand {
		d.attributes["commandOnlyVolume"] = true
	}
	d.attributes["volumeMaxLevel"] = maxLevel
	d.attributes["volumeCanMuteAndUnmute"] = canMute

	return d
}
