package main

import (
	"context"

	action "github.com/rmrobinson/google-smart-home-action-go"
	"go.uber.org/zap"
)

type lightbulb struct {
	id         string
	name       string
	isOn       bool
	brightness int

	color struct {
		hue        float64
		saturation float64
		value      float64
	}
}

func (l *lightbulb) GetState() action.DeviceState {
	return action.NewDeviceState(true).RecordOnOff(l.isOn).RecordBrightness(l.brightness).RecordColorHSV(l.color.hue, l.color.saturation, l.color.value)
}

type receiver struct {
	id        string
	name      string
	isOn      bool
	volume    int
	muted     bool
	currInput string
}

func (r *receiver) GetState() action.DeviceState {
	return action.NewDeviceState(true).RecordOnOff(r.isOn).RecordInput(r.currInput).RecordVolume(r.volume, r.muted)
}

type echoService struct {
	logger  *zap.Logger
	service *action.Service
	agentID string

	lights   map[string]lightbulb
	receiver receiver
}

func (es *echoService) Sync(context.Context, string) (*action.SyncResponse, error) {
	es.logger.Debug("sync")

	resp := &action.SyncResponse{}
	for _, lb := range es.lights {
		ad := action.NewLight(lb.id)
		ad.Name = action.DeviceName{
			DefaultNames: []string{
				"Test lamp",
			},
			Name: lb.name,
		}
		ad.WillReportState = false
		ad.RoomHint = "test room"
		ad.DeviceInfo = action.DeviceInfo{
			Manufacturer: "faltung systems",
			Model:        "tl001",
			HwVersion:    "0.2",
			SwVersion:    "0.3",
		}
		ad.AddOnOffTrait(false, false).AddBrightnessTrait(false).AddColourTrait(action.HSV, false)

		resp.Devices = append(resp.Devices, ad)
	}

	inputs := []action.DeviceInput{
		{
			Key: "input_1",
			Names: []action.DeviceInputName{
				{
					Synonyms: []string{
						"Input 1",
						"Google Chromecast Audio",
					},
					LanguageCode: "en",
				},
			},
		},
		{
			Key: "input_2",
			Names: []action.DeviceInputName{
				{
					Synonyms: []string{
						"Input 2",
						"Raspberry Pi",
					},
					LanguageCode: "en",
				},
			},
		},
	}
	ar := action.NewSimpleAVReceiver(es.receiver.id, inputs, 100, true, false)
	ar.Name = action.DeviceName{
		DefaultNames: []string{
			"Test receiver",
		},
		Name: es.receiver.name,
	}
	ar.WillReportState = true
	ar.RoomHint = "test room"
	ar.DeviceInfo = action.DeviceInfo{
		Manufacturer: "faltung systems",
		Model:        "tavr001",
		HwVersion:    "0.2",
		SwVersion:    "0.3",
	}

	resp.Devices = append(resp.Devices, ar)

	return resp, nil
}
func (es *echoService) Disconnect(context.Context, string) error {
	es.logger.Debug("disconnect")
	return nil
}
func (es *echoService) Query(_ context.Context, req *action.QueryRequest) (*action.QueryResponse, error) {
	es.logger.Debug("query")

	resp := &action.QueryResponse{
		States: map[string]action.DeviceState{},
	}

	for _, deviceArg := range req.Devices {
		if light, found := es.lights[deviceArg.ID]; found {
			resp.States[deviceArg.ID] = light.GetState()
		} else if es.receiver.id == deviceArg.ID {
			resp.States[deviceArg.ID] = es.receiver.GetState()
		}
	}

	return resp, nil
}
func (es *echoService) Execute(_ context.Context, req *action.ExecuteRequest) (*action.ExecuteResponse, error) {
	es.logger.Debug("execute")

	resp := &action.ExecuteResponse{
		UpdatedState: action.NewDeviceState(true),
	}

	for _, commandArg := range req.Commands {
		for _, command := range commandArg.Commands {
			es.logger.Debug("received command",
				zap.String("command", command.Name),
			)

			for _, deviceArg := range commandArg.TargetDevices {
				if es.receiver.id == deviceArg.ID {
					if command.OnOff != nil {
						es.receiver.isOn = command.OnOff.On
						resp.UpdatedState.RecordOnOff(es.receiver.isOn)
					} else if command.SetVolume != nil {
						es.receiver.volume = command.SetVolume.Level
						resp.UpdatedState.RecordVolume(es.receiver.volume, false)
					} else if command.AdjustVolume != nil {
						es.receiver.volume += command.AdjustVolume.Amount
						resp.UpdatedState.RecordVolume(es.receiver.volume, false)
					} else if command.SetInput != nil {
						es.receiver.currInput = command.SetInput.NewInput
						resp.UpdatedState.RecordInput(es.receiver.currInput)
					} else {
						es.logger.Info("unsupported command",
							zap.String("command", command.Name),
						)
						continue
					}

					resp.UpdatedDevices = append(resp.UpdatedDevices, deviceArg.ID)
					continue
				} else if device, found := es.lights[deviceArg.ID]; found {
					if command.OnOff != nil {
						device.isOn = command.OnOff.On
						resp.UpdatedState.RecordOnOff(device.isOn)
						es.lights[deviceArg.ID] = device
					} else if command.BrightnessAbsolute != nil {
						device.brightness = command.BrightnessAbsolute.Brightness
						resp.UpdatedState.RecordBrightness(device.brightness)
						es.lights[deviceArg.ID] = device
					} else if command.BrightnessRelative != nil {
						device.brightness += command.BrightnessRelative.RelativeWeight
						resp.UpdatedState.RecordBrightness(device.brightness)
						es.lights[deviceArg.ID] = device
					} else if command.ColorAbsolute != nil {
						device.color.hue = command.ColorAbsolute.Color.HSV.Hue
						device.color.saturation = command.ColorAbsolute.Color.HSV.Saturation
						device.color.value = command.ColorAbsolute.Color.HSV.Value
						resp.UpdatedState.RecordColorHSV(device.color.hue, device.color.saturation, device.color.value)
						es.lights[deviceArg.ID] = device
					} else {
						es.logger.Info("unsupported command",
							zap.String("command", command.Name))
						continue
					}

					resp.UpdatedDevices = append(resp.UpdatedDevices, deviceArg.ID)
					continue
				}

				es.logger.Info("device not found",
					zap.String("device_id", deviceArg.ID),
					zap.String("command", command.Name),
				)
			}
		}
	}

	return resp, nil
}

func (es *echoService) toggleLight1() {
	l := es.lights["123"]
	l.isOn = !l.isOn
	es.lights["123"] = l
	err := es.service.ReportState(context.Background(), es.agentID, map[string]action.DeviceState{
		"123": l.GetState(),
	})
	if err != nil {
		es.logger.Error("unable to report state",
			zap.Error(err),
		)
	}
}
