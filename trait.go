package action

import "encoding/json"

// DeviceState contains the state of a device.
type DeviceState struct {
	Online bool
	Status string

	State map[string]interface{}
}

// NewDeviceState creates a new device state to be added to as defined by the relevant traits on a device.
func NewDeviceState(online bool) DeviceState {
	return DeviceState{
		Online: online,
		State:  map[string]interface{}{},
	}
}

// RecordBrightness adds the current brightness to the device.
// Should only be applied to devices with the Brightness trait
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (ds DeviceState) RecordBrightness(brightness int) DeviceState {
	ds.State["brightness"] = brightness
	return ds
}

// RecordColorTemperature adds the current color temperature (in Kelvin) to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds DeviceState) RecordColorTemperature(temperatureK int) DeviceState {
	ds.State["color"] = map[string]interface{}{
		"temperatureK": temperatureK,
	}
	return ds
}

// RecordColorRGB adds the current color in RGB to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds DeviceState) RecordColorRGB(spectrumRgb int) DeviceState {
	ds.State["color"] = map[string]interface{}{
		"spectrumRgb": spectrumRgb,
	}
	return ds
}

// RecordColorHSV adds the current color in HSV to the device.
// Should only be applied to devices with the ColorSetting trait
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (ds DeviceState) RecordColorHSV(hue float64, saturation float64, value float64) DeviceState {
	ds.State["color"] = map[string]interface{}{
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
func (ds DeviceState) RecordInput(input string) DeviceState {
	ds.State["input"] = input
	return ds
}

// RecordOnOff adds the current on/off state to the device.
// Should only be applied to devices with the OnOff trait
// See https://developers.google.com/assistant/smarthome/traits/onoff
func (ds DeviceState) RecordOnOff(on bool) DeviceState {
	ds.State["on"] = on
	return ds
}

// RecordVolume adds the current volume state to the device.
// Should only be applied to devices with the Volume trait
// See https://developers.google.com/assistant/smarthome/traits/volume
func (ds DeviceState) RecordVolume(volume int, isMuted bool) DeviceState {
	ds.State["currentVolume"] = volume
	ds.State["isMuted"] = isMuted
	return ds
}

// MarshalJSON is a custom JSON serializer for our DeviceState
func (ds DeviceState) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{}
	payload["online"] = ds.Online
	if len(ds.Status) > 0 {
		payload["status"] = ds.Status
	}

	for k, v := range ds.State {
		payload[k] = v
	}

	return json.Marshal(payload)
}

// UnmarshalJSON is a custom JSON deserializer for our DeviceState
func (ds *DeviceState) UnmarshalJSON(data []byte) error {
	payload := map[string]interface{}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if online, ok := payload["online"]; ok {
		ds.Online = online.(bool)
		delete(payload, "online")
	}
	if status, ok := payload["status"]; ok {
		ds.Status = status.(string)
		delete(payload, "status")
	}

	ds.State = payload

	return nil
}
