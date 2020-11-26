package action

import (
	"context"

	"go.uber.org/zap"
)

const (
	// GoogleFulfillmentPath represents the HTTP path which the Google fulfillment hander will call
	GoogleFulfillmentPath = "/fulfillment"
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
}

// NewService creates a new service to handle Google Action operations.
// It is required that an access token validator be specified to properly process requests.
// This access token validator should be pointed to the same data source as the OAuth2 server configured in the Google Smart Home Actions portal in the OAuth2 account linking section.
func NewService(logger *zap.Logger, atValidator AccessTokenValidator, provider Provider) *Service {
	if atValidator == nil {
		logger.Fatal("empty access token validator not allowed")
	}
	if provider == nil {
		logger.Fatal("empty provider not allowed")
	}

	return &Service{
		logger:      logger,
		atValidator: atValidator,
		provider:    provider,
	}
}
