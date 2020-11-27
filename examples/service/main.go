package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	action "github.com/rmrobinson/google-smart-home-action-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

type auth0Authenticator struct {
	logger *zap.Logger

	domain string
	client *http.Client
	tokens map[string]string
}

func (a *auth0Authenticator) Validate(ctx context.Context, token string) (string, error) {
	a.logger.Debug("validate",
		zap.String("token", token),
	)

	if userid, found := a.tokens[token]; found {
		return userid, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/userinfo", a.domain), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	contentType := resp.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		return "", errors.New("content not json")
	}

	var respPayload struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}

	err = json.NewDecoder(resp.Body).Decode(&respPayload)
	if err != nil {
		return "", err
	}

	a.logger.Info("token validated",
		zap.String("userid", respPayload.Sub),
		zap.String("email", respPayload.Email),
	)
	a.tokens[token] = respPayload.Sub

	return respPayload.Sub, nil
}

type device struct {
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

type echoService struct {
	logger *zap.Logger

	devices map[string]device
}

func (es *echoService) Sync(context.Context) (*action.SyncResponse, error) {
	es.logger.Debug("sync")

	resp := &action.SyncResponse{}
	for _, d := range es.devices {
		ad := action.NewLight(d.id)
		ad.Name = action.DeviceName{
			DefaultNames: []string{
				"Test lamp",
			},
			Name: d.name,
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

	return resp, nil
}
func (es *echoService) Disconnect(context.Context) error {
	es.logger.Debug("disconnect")
	return nil
}
func (es *echoService) Query(_ context.Context, req *action.QueryRequest) (*action.QueryResponse, error) {
	es.logger.Debug("query")

	resp := &action.QueryResponse{
		States: map[string]action.DeviceState{},
	}

	for _, deviceArg := range req.Devices {
		if device, found := es.devices[deviceArg.ID]; found {
			resp.States[deviceArg.ID] = action.NewDeviceState(true, "SUCCESS").RecordOnOff(device.isOn).RecordBrightness(device.brightness).RecordColorHSV(device.color.hue, device.color.saturation, device.color.value)
		}
	}

	return resp, nil
}
func (es *echoService) Execute(_ context.Context, req *action.ExecuteRequest) (*action.ExecuteResponse, error) {
	es.logger.Debug("execute")

	resp := &action.ExecuteResponse{
		UpdatedState: action.NewDeviceState(true, "SUCCESS"),
	}

	for _, commandArg := range req.Commands {
		for _, command := range commandArg.Commands {
			es.logger.Debug("received command",
				zap.String("command", command.Name),
			)

			for _, deviceArg := range commandArg.TargetDevices {
				if device, found := es.devices[deviceArg.ID]; found {
					if command.OnOff != nil {
						device.isOn = command.OnOff.On
						resp.UpdatedState.RecordOnOff(device.isOn)
						es.devices[deviceArg.ID] = device
					} else if command.BrightnessAbsolute != nil {
						device.brightness = command.BrightnessAbsolute.Brightness
						resp.UpdatedState.RecordBrightness(device.brightness)
						es.devices[deviceArg.ID] = device
					} else if command.BrightnessRelative != nil {
						device.brightness += command.BrightnessRelative.RelativeWeight
						resp.UpdatedState.RecordBrightness(device.brightness)
						es.devices[deviceArg.ID] = device
					} else if command.ColorAbsolute != nil {
						device.color.hue = command.ColorAbsolute.HSV.Hue
						device.color.saturation = command.ColorAbsolute.HSV.Saturation
						device.color.value = command.ColorAbsolute.HSV.Value
						resp.UpdatedState.RecordColorHSV(device.color.hue, device.color.saturation, device.color.value)
						es.devices[deviceArg.ID] = device
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

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	auth := &auth0Authenticator{
		logger: logger,
		domain: "fal-qn9lmwum.us.auth0.com",
		client: &http.Client{},
		tokens: map[string]string{},
	}

	es := &echoService{
		logger: logger,
		devices: map[string]device{
			"123": {
				"123",
				"test device 1",
				false,
				40,
				struct {
					hue        float64
					saturation float64
					value      float64
				}{
					100,
					100,
					10,
				},
			},
			"456": {
				"456",
				"test device 2",
				false,
				40,
				struct {
					hue        float64
					saturation float64
					value      float64
				}{
					100,
					100,
					10,
				},
			},
		},
	}

	// Setup Google Assistant info
	svc := action.NewService(logger, auth, es)

	// Register callback from Google
	http.HandleFunc(action.GoogleFulfillmentPath, svc.GoogleFulfillmentHandler)

	// Setup LetsEncrypt
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("gsha-fulfillment.faltung.ca"), //Your domain here
		Cache:      autocert.DirCache("certs"),                            //Folder for storing certificates
	}

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))

	logger.Info("listening")

	log.Fatal(server.ListenAndServeTLS("", ""))
}
