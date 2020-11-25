package action

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testAuthenticator struct{}

func (ta *testAuthenticator) Validate(context.Context, string) (string, error) {
	return "1836.15267389", nil
}

type testProvider struct {
	syncResp []*Device
	syncErr  error

	queryReq   *QueryRequest
	executeReq *ExecuteRequest
}

func (tp *testProvider) Sync(context.Context) ([]*Device, error) {
	return tp.syncResp, tp.syncErr
}

func (tp *testProvider) Disconnect(context.Context) error {
	return nil
}

func (tp *testProvider) Query(_ context.Context, req *QueryRequest) error {
	tp.queryReq = req
	return nil
}

func (tp *testProvider) Execute(_ context.Context, req *ExecuteRequest) error {
	tp.executeReq = req
	return nil
}

func TestGoogleFulfillmentHandlerSync(t *testing.T) {
	logger := zaptest.NewLogger(t)

	authenticator := &testAuthenticator{}
	provider := &testProvider{}

	d1 := NewDevice()
	d1.ID = "123"
	d1.Type = "action.devices.types.OUTLET"
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
	d1.AddOnOff(false, false)

	d2 := NewDevice()
	d2.ID = "456"
	d2.Type = "action.devices.types.LIGHT"
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
	d2.AddOnOff(false, false).AddBrightness(false).AddColourSetting(RGB, false).AddColourTemperatureSetting(2000, 9000, false)

	provider.syncResp = []*Device{d1, d2}

	svc := NewService(logger, authenticator, provider)

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
	assert.Equal(t, `{"requestId":"ff36a3cc-ec34-11e6-b1a0-64510650abcf","payload":{"agentUserId":"1836.15267389","devices":[{"id":"123","type":"action.devices.types.OUTLET","traits":["action.devices.traits.OnOff"],"name":{"defaultNames":["My Outlet 1234"],"name":"Night light","nicknames":["wall plug"]},"willReportState":false,"roomHint":"kitchen","deviceInfo":{"manufacturer":"lights-out-inc","model":"hs1234","hwVersion":"3.2","swVersion":"11.4"},"otherDeviceIds":[{"deviceId":"local-device-id"}],"customData":{"barValue":true,"bazValue":"foo","fooValue":74}},{"id":"456","type":"action.devices.types.LIGHT","traits":["action.devices.traits.OnOff","action.devices.traits.Brightness","action.devices.traits.ColorSetting"],"name":{"defaultNames":["lights out inc. bulb A19 color hyperglow"],"name":"lamp1","nicknames":["reading lamp"]},"willReportState":false,"roomHint":"office","attributes":{"colorModel":"rgb","colorTemperatureRange":{"temperatureMaxK":9000,"temperatureMinK":2000},"commandOnlyColorSetting":false},"deviceInfo":{"manufacturer":"lights out inc.","model":"hg11","hwVersion":"1.2","swVersion":"5.4"},"customData":{"barValue":false,"bazValue":"bar","fooValue":12}}]}}
`, rr.Body.String())
}
