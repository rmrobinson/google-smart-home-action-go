package action

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/api/homegraph/v1"
	"google.golang.org/api/option"
)

// newTestService points a Service at ts instead of the real HomeGraph API, so ReportState/
// ReportStateAndNotify's HTTP call can be asserted on without any real credentials or network
// access - option.WithHTTPClient/WithEndpoint bypass NewService's normal ADC lookup entirely.
func newTestService(t *testing.T, ts *httptest.Server) *Service {
	hg, err := homegraph.NewService(context.Background(),
		option.WithHTTPClient(ts.Client()),
		option.WithEndpoint(ts.URL),
		option.WithoutAuthentication(),
	)
	require.NoError(t, err)

	return &Service{
		logger:        zaptest.NewLogger(t),
		deviceService: homegraph.NewDevicesService(hg),
	}
}

func TestReportState_SendsStatesOnly(t *testing.T) {
	var gotBody homegraph.ReportStateAndNotificationRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&homegraph.ReportStateAndNotificationResponse{RequestId: gotBody.RequestId})
	}))
	defer ts.Close()

	svc := newTestService(t, ts)
	err := svc.ReportState(context.Background(), "user1", map[string]DeviceState{
		"light1": NewDeviceState(true),
	})
	require.NoError(t, err)

	require.Equal(t, "user1", gotBody.AgentUserId)
	require.NotNil(t, gotBody.Payload)
	require.NotNil(t, gotBody.Payload.Devices)
	assert.Empty(t, gotBody.Payload.Devices.Notifications)

	var states map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(gotBody.Payload.Devices.States, &states))
	assert.Contains(t, states, "light1")
}

func TestReportStateAndNotify_SendsStatesAndNotifications(t *testing.T) {
	var gotBody homegraph.ReportStateAndNotificationRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&homegraph.ReportStateAndNotificationResponse{RequestId: gotBody.RequestId})
	}))
	defer ts.Close()

	svc := newTestService(t, ts)
	notifications := map[string]interface{}{
		"camera1": map[string]interface{}{
			"ObjectDetection": map[string]interface{}{
				"priority":           0,
				"detectionTimestamp": 1700000000000,
				"objects":            map[string]interface{}{"unfamiliar": 1},
			},
		},
	}
	err := svc.ReportStateAndNotify(context.Background(), "user1", nil, notifications, "evt1")
	require.NoError(t, err)

	assert.Equal(t, "evt1", gotBody.EventId)
	require.NotNil(t, gotBody.Payload.Devices)
	// deviceStates was nil - states must be omitted entirely, not sent as "null"/"{}", since
	// ObjectDetection has no steady state to report.
	assert.Empty(t, gotBody.Payload.Devices.States)

	var gotNotifications map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(gotBody.Payload.Devices.Notifications, &gotNotifications))
	objDetection := gotNotifications["camera1"]["ObjectDetection"].(map[string]interface{})
	assert.Equal(t, float64(0), objDetection["priority"])
	assert.Equal(t, float64(1700000000000), objDetection["detectionTimestamp"])
}

func TestReportStateAndNotify_PropagatesHTTPFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	svc := newTestService(t, ts)
	err := svc.ReportStateAndNotify(context.Background(), "user1", nil, map[string]interface{}{"d1": map[string]interface{}{}}, "")
	assert.Error(t, err)
}
