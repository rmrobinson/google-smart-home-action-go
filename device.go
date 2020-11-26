package action

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
	attributes map[string]interface{}

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
		attributes: map[string]interface{}{},
		CustomData: map[string]interface{}{},
	}
}

// DeviceInput represents a single input of a device
type DeviceInput struct {
	Key   string `json:"key"`
	Names []struct {
		LanguageCode string   `json:"lang"`
		Synonyms     []string `json:"name_synonym"`
	} `json:"names"`
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
