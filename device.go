package action

// DeviceName contains different ways of identifying the device
type DeviceName struct {
	// DefaultNames (not user settable)
	DefaultNames []string `json:"defaultNames,omitempty"`
	// Name supplied by the user for display purposes
	Name string `json:"name,omitempty"`
	// Nicknames given to this, should a user have multiple ways to refer to the device
	Nicknames []string `json:"nicknames,omitempty"`
}

// DeviceInfo contains different properties of the device
type DeviceInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	HwVersion    string `json:"hwVersion,omitempty"`
	SwVersion    string `json:"swVersion,omitempty"`
}

// OtherDeviceID contains alternative ways to identify this device.
type OtherDeviceID struct {
	AgentID  string `json:"agentId,omitempty"`
	DeviceID string `json:"deviceId,omitempty"`
}

// Device represents a single provider-supplied device profile.
type Device struct {
	// ID of the device
	ID string `json:"id,omitempty"`

	// Type of the device.
	// See https://developers.google.com/assistant/smarthome/guides is a list of possible types
	Type string `json:"type,omitempty"`

	// Traits of the device.
	// See https://developers.google.com/assistant/smarthome/traits for a list of possible traits
	// The set of assigned traits will dictate which actions can be performed on the device
	Traits []string `json:"traits,omitempty"`

	// Name of the device.
	Name DeviceName `json:"name,omitempty"`

	// WillReportState using the ReportState API (should be true)
	WillReportState bool `json:"willReportState"`

	// RoomHint guides Google as to which room this device is in
	RoomHint string `json:"roomHint,omitempty"`

	// Attributes linked to the defined traits
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// DeviceInfo that is physically defined
	DeviceInfo DeviceInfo `json:"deviceInfo,omitempty"`

	OtherDeviceIDs []OtherDeviceID `json:"otherDeviceIds,omitempty"`

	// CustomData specified which will be included unmodified in subsequent requests.
	CustomData map[string]interface{} `json:"customData,omitempty"`
}

// NewDevice creates a new device ready for setting things in.
func NewDevice() *Device {
	return &Device{
		Attributes: map[string]interface{}{},
		CustomData: map[string]interface{}{},
	}
}

// AddBrightness indicates this device is capable of having its brightness controlled.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only switch).
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (d *Device) AddBrightness(onlyCommand bool) *Device {
	d.Traits = append(d.Traits, "action.devices.traits.Brightness")
	if onlyCommand {
		d.Attributes["commandOnlyBrightness"] = true
	}

	return d
}

// ColorModel defines which model of the color wheel the device supports.
const (
	RGB = "rgb"
	HSV = "hsv"
)

// AddColourSetting indicates this device is capable of having its colour display controlled using the specified color model.
// It is mutually exclusive to support RGB or HSV.
// It is possible to support either one of RGB and HSV alongside color temperature. See AddColorTemperatureSetting
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourSetting(model string, onlyCommand bool) *Device {
	hasTrait := false
	for _, t := range d.Traits {
		if t == "action.devices.traits.ColorSetting" {
			hasTrait = true
			break
		}
	}
	if !hasTrait {
		d.Traits = append(d.Traits, "action.devices.traits.ColorSetting")
	}
	if onlyCommand {
		d.Attributes["commandOnlyColorSetting"] = true
	}
	d.Attributes["colorModel"] = model

	return d
}

// AddColourTemperatureSetting indicates this device is capable of having its colour display controlled using the colour temperature model.
// This can be set alongside AddColorSetting to indicate both color temperature and another algorithm are supported.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourTemperatureSetting(minTempK int, maxTempK int, onlyCommand bool) *Device {
	hasTrait := false
	for _, t := range d.Traits {
		if t == "action.devices.traits.ColorSetting" {
			hasTrait = true
			break
		}
	}
	if !hasTrait {
		d.Traits = append(d.Traits, "action.devices.traits.ColorSetting")
	}
	if onlyCommand {
		d.Attributes["commandOnlyColorSetting"] = true
	} else {
		d.Attributes["commandOnlyColorSetting"] = false
	}
	d.Attributes["colorTemperatureRange"] = map[string]int{
		"temperatureMinK": minTempK,
		"temperatureMaxK": maxTempK,
	}

	return d
}

// AddOnOff indicates this device is capable of having its state toggled on or off.
// If the device can be commanded but not queried, set onlyCommand to true (i.e. a write-only switch).
// If the devie cannot be commanded but only queried, set onlyQuery to true (i.e. a sensor).
// See https://developers.google.com/assistant/smarthome/traits/onoff
func (d *Device) AddOnOff(onlyCommand, onlyQuery bool) *Device {
	d.Traits = append(d.Traits, "action.devices.traits.OnOff")
	if onlyCommand {
		d.Attributes["commandOnlyOnOff"] = true
	}
	if onlyQuery {
		d.Attributes["queryOnlyOnOff"] = true
	}

	return d
}
