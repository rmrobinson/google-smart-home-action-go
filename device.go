package action

import (
	"encoding/json"
	"sort"
)

// DeviceName contains different ways of identifying the device
type DeviceName struct {
	// DefaultNames (not user settable)
	DefaultNames []string
	// Name supplied by the user for display purposes
	Name string
	// Nicknames given to this, should a user have multiple ways to refer to the device
	Nicknames []string
}

// DeviceInfo contains different properties of the device
type DeviceInfo struct {
	// Manufacturer of the device
	Manufacturer string
	// Model of the device
	Model string
	// HwVersion of the device
	HwVersion string
	// SwVersion of the device
	SwVersion string
}

// OtherDeviceID contains alternative ways to identify this device.
type OtherDeviceID struct {
	AgentID  string
	DeviceID string
}

// Device represents a single provider-supplied device profile.
type Device struct {
	// ID of the device
	ID string

	// Type of the device.
	// See https://developers.google.com/assistant/smarthome/guides is a list of possible types
	Type string

	// Traits of the device.
	// See https://developers.google.com/assistant/smarthome/traits for a list of possible traits
	// The set of assigned traits will dictate which actions can be performed on the device
	Traits map[string]bool

	// Name of the device.
	Name DeviceName

	// WillReportState using the ReportState API (should be true)
	WillReportState bool

	// RoomHint guides Google as to which room this device is in
	RoomHint string

	// Attributes linked to the defined traits
	Attributes map[string]interface{}

	// DeviceInfo that is physically defined
	DeviceInfo DeviceInfo

	// OtherDeviceIDs allows for this to be logically linked to other devices
	OtherDeviceIDs []OtherDeviceID

	// CustomData specified which will be included unmodified in subsequent requests.
	CustomData map[string]interface{}
}

// NewDevice creates a new device ready for setting things in.
func NewDevice(id string, typ string) *Device {
	return &Device{
		ID:         id,
		Type:       typ,
		Traits:     map[string]bool{},
		Attributes: map[string]interface{}{},
		CustomData: map[string]interface{}{},
	}
}

// DeviceInputName represents the human-readable name shown for an input
type DeviceInputName struct {
	LanguageCode string   `json:"lang"`
	Synonyms     []string `json:"name_synonym"`
}

// DeviceInput represents a single input of a device
type DeviceInput struct {
	Key   string            `json:"key"`
	Names []DeviceInputName `json:"names"`
}

// NewSimpleAVReceiver creates a new device with the attributes for a simple AV receiver setup.
func NewSimpleAVReceiver(id string, inputs []DeviceInput, maxLevel int, canMute bool, onlyCommand bool) *Device {
	d := NewDevice(id, "action.devices.types.AUDIO_VIDEO_RECEIVER")
	d.AddOnOffTrait(false, false)
	d.AddInputSelectorTrait(inputs, canMute)
	d.AddVolumeTrait(maxLevel, canMute, onlyCommand)
	return d
}

// NewLight creates a new device with the attributes for an on-off light.
// This can be customized with any of the light-related traits (Color, Brightness).
func NewLight(id string) *Device {
	d := NewDevice(id, "action.devices.types.LIGHT")
	d.AddOnOffTrait(false, false)
	return d
}

// NewOutlet creates a new device with the attributes for an on-off outlet.
func NewOutlet(id string) *Device {
	d := NewDevice(id, "action.devices.types.OUTLET")
	d.AddOnOffTrait(false, false)
	return d
}

// NewSwitch creates a new device with the attributes for an on-off switch.
// This can be customized with the Brightness trait.
func NewSwitch(id string) *Device {
	d := NewDevice(id, "action.devices.types.SWITCH")
	d.AddOnOffTrait(false, false)
	return d
}

// AddBrightnessTrait indicates this device is capable of having its brightness controlled.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only switch).
// See https://developers.google.com/assistant/smarthome/traits/brightness
func (d *Device) AddBrightnessTrait(onlyCommand bool) *Device {
	d.Traits["action.devices.traits.Brightness"] = true
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

// AddColourTrait indicates this device is capable of having its colour display controlled using the specified color model.
// It is mutually exclusive to support RGB or HSV.
// It is possible to support either one of RGB and HSV alongside color temperature. See AddColorTemperatureSetting
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourTrait(model string, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.ColorSetting"] = true
	if onlyCommand {
		d.Attributes["commandOnlyColorSetting"] = true
	}
	d.Attributes["colorModel"] = model

	return d
}

// AddColourTemperatureTrait indicates this device is capable of having its colour display controlled using the colour temperature model.
// This can be set alongside AddColorSetting to indicate both color temperature and another algorithm are supported.
// If the device does not support querying, set onlyCommand to true (i.e. a write-only lightbulb).
// See https://developers.google.com/assistant/smarthome/traits/colorsetting
func (d *Device) AddColourTemperatureTrait(minTempK int, maxTempK int, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.ColorSetting"] = true

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

// AddInputSelectorTrait indicates this device is capable of having its input selected.
// See https://developers.google.com/assistant/smarthome/traits/inputselector
func (d *Device) AddInputSelectorTrait(availableInputs []DeviceInput, ordered bool) *Device {
	d.Traits["action.devices.traits.InputSelector"] = true
	d.Attributes["availableInputs"] = availableInputs
	d.Attributes["orderedInputs"] = ordered

	return d
}

// AddOnOffTrait indicates this device is capable of having its state toggled on or off.
// If the device can be commanded but not queried, set onlyCommand to true (i.e. a write-only switch).
// If the devie cannot be commanded but only queried, set onlyQuery to true (i.e. a sensor).
// See https://developers.google.com/assistant/smarthome/traits/onoff
func (d *Device) AddOnOffTrait(onlyCommand, onlyQuery bool) *Device {
	d.Traits["action.devices.traits.OnOff"] = true
	if onlyCommand {
		d.Attributes["commandOnlyOnOff"] = true
	}
	if onlyQuery {
		d.Attributes["queryOnlyOnOff"] = true
	}

	return d
}

// AddVolumeTrait indicates this device is capable of having its volume controlled
// See https://developers.google.com/assistant/smarthome/traits/volume
func (d *Device) AddVolumeTrait(maxLevel int, canMute bool, onlyCommand bool) *Device {
	d.Traits["action.devices.traits.Volume"] = true
	if onlyCommand {
		d.Attributes["commandOnlyVolume"] = true
	}
	d.Attributes["volumeMaxLevel"] = maxLevel
	d.Attributes["volumeCanMuteAndUnmute"] = canMute

	return d
}

// MarshalJSON is a custom JSON serializer for our Device
func (d Device) MarshalJSON() ([]byte, error) {
	dr := deviceRaw{}

	dr.ID = d.ID
	dr.Type = d.Type
	for trait := range d.Traits {
		dr.Traits = append(dr.Traits, trait)
	}
	sort.Strings(dr.Traits)
	dr.Name.DefaultNames = d.Name.DefaultNames
	dr.Name.Name = d.Name.Name
	dr.Name.Nicknames = d.Name.Nicknames
	dr.WillReportState = d.WillReportState
	dr.RoomHint = d.RoomHint
	dr.Attributes = d.Attributes
	dr.DeviceInfo.Manufacturer = d.DeviceInfo.Manufacturer
	dr.DeviceInfo.Model = d.DeviceInfo.Model
	dr.DeviceInfo.HwVersion = d.DeviceInfo.HwVersion
	dr.DeviceInfo.SwVersion = d.DeviceInfo.SwVersion
	for _, otherDeviceID := range d.OtherDeviceIDs {
		dr.OtherDeviceIDs = append(dr.OtherDeviceIDs, otherDeviceIDraw{
			AgentID:  otherDeviceID.AgentID,
			DeviceID: otherDeviceID.DeviceID,
		})
	}
	dr.CustomData = d.CustomData

	return json.Marshal(dr)
}

// UnmarshalJSON is a custom JSON deserializer for our Device
func (d *Device) UnmarshalJSON(data []byte) error {
	dr := deviceRaw{}
	if err := json.Unmarshal(data, &dr); err != nil {
		return err
	}

	if d.Traits == nil {
		d.Traits = map[string]bool{}
	}

	d.ID = dr.ID
	d.Type = dr.Type
	for _, trait := range dr.Traits {
		d.Traits[trait] = true
	}
	d.Name.DefaultNames = dr.Name.DefaultNames
	d.Name.Name = dr.Name.Name
	d.Name.Nicknames = dr.Name.Nicknames
	d.WillReportState = dr.WillReportState
	d.RoomHint = dr.RoomHint
	d.Attributes = dr.Attributes
	d.DeviceInfo.Manufacturer = dr.DeviceInfo.Manufacturer
	d.DeviceInfo.Model = dr.DeviceInfo.Model
	d.DeviceInfo.HwVersion = dr.DeviceInfo.HwVersion
	d.DeviceInfo.SwVersion = dr.DeviceInfo.SwVersion
	for _, otherDeviceID := range dr.OtherDeviceIDs {
		d.OtherDeviceIDs = append(d.OtherDeviceIDs, OtherDeviceID{
			AgentID:  otherDeviceID.AgentID,
			DeviceID: otherDeviceID.DeviceID,
		})
	}

	d.CustomData = dr.CustomData
	if d.CustomData == nil {
		d.CustomData = map[string]interface{}{}
	}

	return nil
}

type otherDeviceIDraw struct {
	AgentID  string `json:"agentId,omitempty"`
	DeviceID string `json:"deviceId,omitempty"`
}

type deviceRaw struct {
	ID     string   `json:"id,omitempty"`
	Type   string   `json:"type,omitempty"`
	Traits []string `json:"traits,omitempty"`

	Name struct {
		DefaultNames []string `json:"defaultNames,omitempty"`
		Name         string   `json:"name,omitempty"`
		Nicknames    []string `json:"nicknames,omitempty"`
	} `json:"name,omitempty"`

	WillReportState bool                   `json:"willReportState"`
	RoomHint        string                 `json:"roomHint,omitempty"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`

	DeviceInfo struct {
		Manufacturer string `json:"manufacturer,omitempty"`
		Model        string `json:"model,omitempty"`
		HwVersion    string `json:"hwVersion,omitempty"`
		SwVersion    string `json:"swVersion,omitempty"`
	} `json:"deviceInfo,omitempty"`

	OtherDeviceIDs []otherDeviceIDraw     `json:"otherDeviceIds,omitempty"`
	CustomData     map[string]interface{} `json:"customData,omitempty"`
}
