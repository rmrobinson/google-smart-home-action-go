package action

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/homegraph/v1"
)

const (
	// GoogleFulfillmentPath represents the HTTP path which the Google fulfillment hander will call
	GoogleFulfillmentPath = "/fulfillment"
)

var (
	// ErrSyncFailed is returned if the request to HomeGraph to force a SYNC operation failed.
	// The log will contain more information about what occurred.
	ErrSyncFailed = errors.New("sync failed")
	// ErrReportStateFailed is returned if the request to HomeGraph to update a device failed.
	// The log will contain more information about what occurred.
	ErrReportStateFailed = errors.New("report state failed")
)

// DeviceArg contains the common fields used when executing requests against a device.
// It has 2 fields of note:
// - the ID of the device
// - any custom data supplied for the device ID as part of the original response to SYNC
type DeviceArg struct {
	ID         string
	CustomData map[string]interface{}
}

// CommandArg contains the fields used to execute a change on a set of devices.
// Only one of the various pointers in Command should be set per command.
type CommandArg struct {
	TargetDevices []DeviceArg
	Commands      []Command
}

// SyncResponse contains the set of devices to supply to the Google Smart Home Action when setting up.
type SyncResponse struct {
	Devices []*Device
}

// QueryRequest includes what is being asked for by the Google Smart Home Action when querying.
type QueryRequest struct {
	Devices []DeviceArg
}

// QueryResponse includes what should be returned in response to the query to the Google Home Smart Action.
// The States map should have the same IDs supplied in the request.
type QueryResponse struct {
	States map[string]DeviceState
}

// ExecuteRequest includes what is being asked for by the Google Assistant when making a change.
// The customData is a JSON object originally returned during the Sync operation.
type ExecuteRequest struct {
	Commands []CommandArg
}

// ExecuteResponse includes the results of an Execute command to be sent back to the Google home graph after an execute.
// Between the UpdatedDevices and FailedDevices maps all device IDs in the Execute request should be accounted for.
type ExecuteResponse struct {
	UpdatedState   DeviceState
	UpdatedDevices []string
	OfflineDevices []string
	FailedDevices  map[string]struct {
		Devices []string
	}
}

// AccessTokenValidator allows for the auth token supplied by Google to be validated.
type AccessTokenValidator interface {
	// Validate performs the actual token validation. Returning an error will force validation to fail.
	// The user ID that corresponds to the token should be returned on success.
	Validate(context.Context, string) (string, error)
}

// Provider exposes methods that can be invoked by the Google Smart Home Action intents
type Provider interface {
	Sync(context.Context) (*SyncResponse, error)
	Disconnect(context.Context) error
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
	Execute(context.Context, *ExecuteRequest) (*ExecuteResponse, error)
}

// Service links together the updates coming from the different sources and ensures they are consistent.
// Updates may come in via:
// - the GoogleCallbackHandler, which is registered externally on an HTTPS endpoint which the Google Smart Home Actions framework will POST to
// - the API calls made against this service by the underlying device
type Service struct {
	logger *zap.Logger

	atValidator AccessTokenValidator

	provider Provider

	deviceService *homegraph.DevicesService
}

// NewService creates a new service to handle Google Action operations.
// It is required that an access token validator be specified to properly process requests.
// This access token validator should be pointed to the same data source as the OAuth2 server configured in the Google Smart Home Actions portal in the OAuth2 account linking section.
func NewService(logger *zap.Logger, atValidator AccessTokenValidator, provider Provider, hgService *homegraph.Service) *Service {
	if atValidator == nil {
		logger.Fatal("empty access token validator not allowed")
	}
	if provider == nil {
		logger.Fatal("empty provider not allowed")
	}

	return &Service{
		logger:        logger,
		atValidator:   atValidator,
		provider:      provider,
		deviceService: homegraph.NewDevicesService(hgService),
	}
}

// RequestSync is used to trigger a Google HomeGraph sync operation.
// This should be called whenever the list of devices, or their properties, change.
// This will request a sync occur synchronously, so make sure that the Sync method is not
// blocked on anything this method may be doing.
func (s *Service) RequestSync(ctx context.Context, agentUserID string) error {
	call := s.deviceService.RequestSync(&homegraph.RequestSyncDevicesRequest{
		AgentUserId: agentUserID,
	})
	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		s.logger.Info("error requesting sync",
			zap.String("agent_user_id", agentUserID),
			zap.Error(err),
		)
		return err
	}
	if resp.ServerResponse.HTTPStatusCode != http.StatusOK {
		s.logger.Info("failed request sync",
			zap.String("agent_user_id", agentUserID),
			zap.Int("status_code", resp.ServerResponse.HTTPStatusCode),
		)
		return ErrSyncFailed
	}
	return nil
}

// ReportState is used to report a state change which occurred on a device to the Google HomeGraph.
// This should be called whenever a local action triggers a change, as well as after receiving an Execute callback.
// The supplied state argument should have a complete definition of the device state (i.e. do not perform incremental updates).
// The deviceStates map is indexed by device ID.
// This library does not attempt to report on state changes automatically as it is possible that the action
// triggers a change on the device that is not reflected in the initial request. It is best if the underlying
// service ensures that the Google HomeGraph is kept in sync through an explicit state update after execution.
func (s *Service) ReportState(ctx context.Context, agentUserID string, deviceStates map[string]DeviceState) error {
	jsonState, err := json.Marshal(deviceStates)
	if err != nil {
		s.logger.Info("error serializing device states to json",
			zap.String("agent_user_id", agentUserID),
			zap.Error(err),
		)
	}

	call := s.deviceService.ReportStateAndNotification(&homegraph.ReportStateAndNotificationRequest{
		AgentUserId: agentUserID,
		RequestId:   uuid.New().String(),
		Payload: &homegraph.StateAndNotificationPayload{
			Devices: &homegraph.ReportStateAndNotificationDevice{
				States: jsonState,
			},
		},
	})
	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		s.logger.Info("error requesting sync",
			zap.String("agent_user_id", agentUserID),
			zap.Error(err),
		)
		return err
	}
	if resp.ServerResponse.HTTPStatusCode != http.StatusOK {
		s.logger.Info("failed report state",
			zap.String("agent_user_id", agentUserID),
			zap.Int("status_code", resp.ServerResponse.HTTPStatusCode),
		)
		return ErrSyncFailed
	}
	return nil
}
