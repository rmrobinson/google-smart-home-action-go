package action

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testAuthenticator struct {
	validToken string
	userID     string
}

func (ta *testAuthenticator) Validate(_ context.Context, token string) (string, error) {
	if token == ta.validToken {
		return ta.userID, nil
	}
	return "", nil
}

type testProvider struct {
	syncResp []*Device
	syncErr  error

	queryReq  *QueryRequest
	queryResp map[string]DeviceState
	queryErr  error

	executeReq              *ExecuteRequest
	executeRespDeviceState  DeviceState
	executeRespUpdated      []string
	executeRespOffline      []string
	executeRespFailed       []string
	executeRespFailedReason string
	executeErr              error
}

func (tp *testProvider) Sync(context.Context, string) (*SyncResponse, error) {
	return &SyncResponse{
		Devices: tp.syncResp,
	}, tp.syncErr
}

func (tp *testProvider) Disconnect(context.Context, string) error {
	return nil
}

func (tp *testProvider) Query(_ context.Context, req *QueryRequest) (*QueryResponse, error) {
	tp.queryReq = req
	return &QueryResponse{
		States: tp.queryResp,
	}, tp.queryErr
}

func (tp *testProvider) Execute(_ context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	tp.executeReq = req
	return &ExecuteResponse{
		UpdatedState:   tp.executeRespDeviceState,
		UpdatedDevices: tp.executeRespUpdated,
		OfflineDevices: tp.executeRespOffline,
		FailedDevices: map[string]struct {
			Devices []string
		}{
			tp.executeRespFailedReason: struct {
				Devices []string
			}{
				Devices: tp.executeRespFailed,
			},
		},
	}, tp.executeErr
}

func TestGoogleFulfillmentHandlerSync(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{
		validToken: "asdf",
		userID:     "1836.15267389",
	}
	provider := &testProvider{}

	d1 := NewOutlet("123")
	d1.Name = DeviceName{
		DefaultNames: []string{
			"My Outlet 1234",
		},
		Name: "Night light",
		Nicknames: []string{
			"wall plug",
		},
	}
	d1.WillReportState = false
	d1.RoomHint = "kitchen"
	d1.DeviceInfo = DeviceInfo{
		Manufacturer: "lights-out-inc",
		Model:        "hs1234",
		HwVersion:    "3.2",
		SwVersion:    "11.4",
	}
	d1.OtherDeviceIDs = []OtherDeviceID{
		{
			DeviceID: "local-device-id",
		},
	}
	d1.CustomData = map[string]interface{}{
		"fooValue": 74,
		"barValue": true,
		"bazValue": "foo",
	}

	d2 := NewLight("456")
	d2.Name = DeviceName{
		DefaultNames: []string{
			"lights out inc. bulb A19 color hyperglow",
		},
		Name: "lamp1",
		Nicknames: []string{
			"reading lamp",
		},
	}
	d2.WillReportState = false
	d2.RoomHint = "office"
	d2.DeviceInfo = DeviceInfo{
		Manufacturer: "lights out inc.",
		Model:        "hg11",
		HwVersion:    "1.2",
		SwVersion:    "5.4",
	}
	d2.CustomData = map[string]interface{}{
		"fooValue": 12,
		"barValue": false,
		"bazValue": "bar",
	}
	d2.AddBrightnessTrait(false).AddColourTrait(RGB, false).AddColourTemperatureTrait(2000, 9000, false)

	provider.syncResp = []*Device{d1, d2}

	svc := NewService(logger, authenticator, provider, nil)

	req, err := http.NewRequest(http.MethodPost, GoogleFulfillmentPath, bytes.NewBuffer([]byte(`{
		"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
		"inputs": [
		  {
			"intent": "action.devices.SYNC"
		  }
		]
	}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "bearer asdf")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.GoogleFulfillmentHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"requestId":"ff36a3cc-ec34-11e6-b1a0-64510650abcf","payload":{"agentUserId":"1836.15267389","devices":[{"id":"123","type":"action.devices.types.OUTLET","traits":["action.devices.traits.OnOff"],"name":{"defaultNames":["My Outlet 1234"],"name":"Night light","nicknames":["wall plug"]},"willReportState":false,"roomHint":"kitchen","deviceInfo":{"manufacturer":"lights-out-inc","model":"hs1234","hwVersion":"3.2","swVersion":"11.4"},"otherDeviceIds":[{"deviceId":"local-device-id"}],"customData":{"barValue":true,"bazValue":"foo","fooValue":74}},{"id":"456","type":"action.devices.types.LIGHT","traits":["action.devices.traits.Brightness","action.devices.traits.ColorSetting","action.devices.traits.OnOff"],"name":{"defaultNames":["lights out inc. bulb A19 color hyperglow"],"name":"lamp1","nicknames":["reading lamp"]},"willReportState":false,"roomHint":"office","attributes":{"colorModel":"rgb","colorTemperatureRange":{"temperatureMaxK":9000,"temperatureMinK":2000},"commandOnlyColorSetting":false},"deviceInfo":{"manufacturer":"lights out inc.","model":"hg11","hwVersion":"1.2","swVersion":"5.4"},"customData":{"barValue":false,"bazValue":"bar","fooValue":12}}]}}
`, rr.Body.String())
}

func TestGoogleFulfillmentHandlerQuery(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{
		validToken: "asdf",
		userID:     "1836.15267389",
	}
	provider := &testProvider{}

	d1 := NewOutlet("123")
	d1.CustomData = map[string]interface{}{
		"fooValue": 74,
		"barValue": true,
		"bazValue": "foo",
	}

	d2 := NewLight("456")
	d2.CustomData = map[string]interface{}{
		"fooValue": 12,
		"barValue": false,
		"bazValue": "bar",
	}
	d2.AddBrightnessTrait(false).AddColourTrait(RGB, false).AddColourTemperatureTrait(2000, 9000, false)

	d1State := NewDeviceState(true)
	d1State.RecordOnOff(true)

	d2State := NewDeviceState(true)
	d2State.RecordOnOff(true).RecordBrightness(80).RecordColorRGB(31655)

	provider.queryResp = map[string]DeviceState{
		d1.ID: d1State,
		d2.ID: d2State,
	}

	svc := NewService(logger, authenticator, provider, nil)

	req, err := http.NewRequest(http.MethodPost, GoogleFulfillmentPath, bytes.NewBuffer([]byte(`{
		"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
		"inputs": [
		  {
			"intent": "action.devices.QUERY",
			"payload": {
			  "devices": [
				{
				  "id": "123",
				  "customData": {
					"fooValue": 74,
					"barValue": true,
					"bazValue": "foo"
				  }
				},
				{
				  "id": "456",
				  "customData": {
					"fooValue": 12,
					"barValue": false,
					"bazValue": "bar"
				  }
				}
			  ]
			}
		  }
		]
	  }`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "bearer asdf")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.GoogleFulfillmentHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"requestId":"ff36a3cc-ec34-11e6-b1a0-64510650abcf","payload":{"devices":{"123":{"on":true,"online":true,"status":"SUCCESS"},"456":{"brightness":80,"color":{"spectrumRgb":31655},"on":true,"online":true,"status":"SUCCESS"}}}}
`, rr.Body.String())
}

func TestGoogleFulfillmentHandlerExecute(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{
		validToken: "asdf",
		userID:     "1836.15267389",
	}
	provider := &testProvider{}

	d1 := NewOutlet("123")
	d1.CustomData = map[string]interface{}{
		"fooValue": 74,
		"barValue": true,
		"bazValue": "sheepdip",
	}

	d2 := NewLight("456")
	d2.CustomData = map[string]interface{}{
		"fooValue": 36,
		"barValue": false,
		"bazValue": "moarsheep",
	}
	d2.AddBrightnessTrait(false).AddColourTrait(RGB, false).AddColourTemperatureTrait(2000, 9000, false)

	provider.executeRespDeviceState = NewDeviceState(true)
	provider.executeRespDeviceState.RecordOnOff(true)
	provider.executeRespUpdated = []string{"123"}
	provider.executeRespFailed = []string{"456"}
	provider.executeRespFailedReason = "deviceTurnedOff"

	svc := NewService(logger, authenticator, provider, nil)

	req, err := http.NewRequest(http.MethodPost, GoogleFulfillmentPath, bytes.NewBuffer([]byte(`{
		"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
		"inputs": [
		  {
			"intent": "action.devices.EXECUTE",
			"payload": {
			  "commands": [
				{
				  "devices": [
					{
					  "id": "123",
					  "customData": {
						"fooValue": 74,
						"barValue": true,
						"bazValue": "sheepdip"
					  }
					},
					{
					  "id": "456",
					  "customData": {
						"fooValue": 36,
						"barValue": false,
						"bazValue": "moarsheep"
					  }
					}
				  ],
				  "execution": [
					{
					  "command": "action.devices.commands.OnOff",
					  "params": {
						"on": true
					  }
					}
				  ]
				}
			  ]
			}
		  }
		]
	  }`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "bearer asdf")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.GoogleFulfillmentHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"requestId":"ff36a3cc-ec34-11e6-b1a0-64510650abcf","payload":{"commands":[{"ids":["123"],"status":"SUCCESS","states":{"on":true,"online":true}},{"ids":["456"],"status":"ERROR","errorCode":"deviceTurnedOff"}]}}
`, rr.Body.String())
}

func TestGoogleFulfillmentHandlerDisconnect(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{
		validToken: "asdf",
		userID:     "1836.15267389",
	}
	provider := &testProvider{}

	svc := NewService(logger, authenticator, provider, nil)

	req, err := http.NewRequest(http.MethodPost, GoogleFulfillmentPath, bytes.NewBuffer([]byte(`{
		"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
		"inputs": [
		  {
			"intent": "action.devices.DISCONNECT"
		  }
		]
	  }`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "bearer asdf")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(svc.GoogleFulfillmentHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{}`, rr.Body.String())
}

var badInputTests = []struct {
	name        string
	contentType string
	authHeader  string
	input       string

	expectedStatusCode int
}{
	{
		"non-json isn't supported",
		"text/plain",
		"",
		"",
		http.StatusUnsupportedMediaType,
	},
	{
		"authorization header required",
		"application/json",
		"",
		"",
		http.StatusUnauthorized,
	},
	{
		"bearer token required",
		"application/json",
		"Basic creds",
		"",
		http.StatusUnauthorized,
	},
	{
		"valid token required",
		"application/json",
		"Bearer tokenBad",
		"",
		http.StatusUnauthorized,
	},
	{
		"valid json required",
		"application/json",
		"Bearer tokenOK",
		"{{{}",
		http.StatusBadRequest,
	},
	{
		"only one input supported",
		"application/json",
		"Bearer tokenOK",
		`{
			"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
			"inputs": [
			  {
				"intent": "action.devices.SYNC"
			  },
			  {
				"intent": "action.devices.SYNC"
			  }
			]
		  }`,
		http.StatusBadRequest,
	},
	{
		"unknown intent specified",
		"application/json",
		"Bearer tokenOK",
		`{
			"requestId": "ff36a3cc-ec34-11e6-b1a0-64510650abcf",
			"inputs": [
			  {
				"intent": "action.devices.GOOGLE"
			  }
			]
		  }`,
		http.StatusBadRequest,
	},
}

func TestGoogleFulfillmentHandlerBadInput(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{
		validToken: "tokenOK",
		userID:     "userOK",
	}
	provider := &testProvider{}

	svc := NewService(logger, authenticator, provider, nil)

	for _, tt := range badInputTests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, GoogleFulfillmentPath, bytes.NewBuffer([]byte(tt.input)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("content-type", tt.contentType)
			req.Header.Set("authorization", tt.authHeader)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(svc.GoogleFulfillmentHandler)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
		})
	}
}

func formatJSON(s string) string {
	var out interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		panic(fmt.Errorf("%q is not valid JSON: %w", s, err))
	}
	formatted, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(formatted)
}
