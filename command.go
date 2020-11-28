package action

// Command defines which command, and what details, are being specified.
// Only one of the contained fields will be set at any point in time.
type Command struct {
	Name    string
	Generic *CommandGeneric

	BrightnessAbsolute *CommandBrightnessAbsolute
	BrightnessRelative *CommandBrightnessRelative
	ColorAbsolute      *CommandColorAbsolute
	OnOff              *CommandOnOff
	Mute               *CommandMute
	SetVolume          *CommandSetVolume
	AdjustVolume       *CommandSetVolumeRelative
	SetInput           *CommandSetInput
	NextInput          *CommandNextInput
	PreviousInput      *CommandPreviousInput
}

// CommandGeneric contains a command definition which hasn't been parsed into a specific command structure.
// This is intended to support newly defined commands which callers of this SDK may handle but this does not yet support.
type CommandGeneric struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
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
	Name        string `json:"name"`
	Temperature int    `json:"temperature"`
	RGB         int    `json:"spectrumRGB"`
	HSV         struct {
		Hue        float64 `json:"hue"`
		Saturation float64 `json:"saturation"`
		Value      float64 `json:"value"`
	} `json:"spectrumHSV"`
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
