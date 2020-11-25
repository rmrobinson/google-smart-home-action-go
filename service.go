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

// QueryRequest includes what is being asked for by the Google Smart Home Action when querying.
// The customData is a JSON object originally returned during the Sync operation.
type QueryRequest struct {
	Devices []DeviceArg
}

// ExecuteRequest includes what is being asked for by the Google Smart Home Action when making a change.
// The customData is a JSON object originally returned during the Sync operation.
type ExecuteRequest struct {
	Commands []struct {
		TargetDevices []DeviceArg
		Params        Command
	} `json:"commands"`
}

// AccessTokenValidator allows for the auth token supplied by Google to be validated.
type AccessTokenValidator interface {
	// Validate performs the actual token validation. Returning an error will force validation to fail.
	// The user ID that corresponds to the token should be returned on success.
	Validate(context.Context, string) (string, error)
}

// Provider exposes methods that can be invoked by the Google Smart Home Action intents
type Provider interface {
	Sync(context.Context) ([]*Device, error)
	Disconnect(context.Context) error
	Query(context.Context, *QueryRequest) error
	Execute(context.Context, *ExecuteRequest) error
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
