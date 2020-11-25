package action

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// GoogleFulfillmentHandler must be registered on an HTTPS endpoint at the path specified by GoogleFulfillmentPath
// This HTTPS endpoint needs to be registered on the Smart Home Actions fulfillment path.
// See https://developers.google.com/assistant/smarthome/concepts/fulfillment-authentication or https://developers.google.com/assistant/smarthome/develop/process-intents for details.
func (s *Service) GoogleFulfillmentHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check if we have a valid request.
	contentType := r.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Request not JSON"))
		return
	}

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access Token Required"))
		return
	}

	authTokenParts := strings.Split(authHeader, " ")
	if len(authTokenParts) != 2 || strings.ToLower(authTokenParts[0]) != "bearer" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access Token Must Be Bearer"))
		return
	}

	userID, err := s.atValidator.Validate(r.Context(), authTokenParts[1])
	if err != nil {
		s.logger.Info("error validating token",
			zap.String("token", authTokenParts[1]),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access Token Invalid"))
		return

	}

	if len(userID) < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access Token Invalid"))
		return
	}

	// We have a valid request. Let's deserialize then do something with it.

	fulfillmentReq := &fulfillmentRequest{}
	err = json.NewDecoder(r.Body).Decode(fulfillmentReq)
	if err != nil {
		s.logger.Info("error deserializing body",
			zap.Error(err),
		)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("JSON Deserialization Failed"))
		return
	}

	if len(fulfillmentReq.Inputs) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported number of inputs"))
		return
	}

	// Actually do something and get the response
	s.logger.Debug("processing intent",
		zap.String("request_id", fulfillmentReq.RequestID),
		zap.String("intent", fulfillmentReq.Inputs[0].Intent),
	)

	switch fulfillmentReq.Inputs[0].Intent {
	case "action.devices.SYNC":
		devices, err := s.provider.Sync(r.Context())
		if err != nil {
			s.logger.Info("sync error",
				zap.Error(err),
			)

			// TODO: clean this up possibly using better error handling.
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Fail to sync"))
			return
		}

		syncResp := &syncResponse{
			RequestID: fulfillmentReq.RequestID,
		}
		syncResp.Payload.UserID = userID
		syncResp.Payload.Devices = devices

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(syncResp)
		if err != nil {
			s.logger.Info("error serializing after writing ok",
				zap.Error(err),
			)
		}
		return
	case "action.devices.QUERY":
		queryReq := &QueryRequest{}
		for _, device := range fulfillmentReq.Inputs[0].Query.Devices {
			queryReq.Devices = append(queryReq.Devices, DeviceArg{
				ID:         device.ID,
				CustomData: device.CustomData,
			})
		}

		deviceState, err := s.provider.Query(r.Context(), queryReq)
		if err != nil {
			s.logger.Info("query error",
				zap.Error(err),
			)

			// TODO: clean this up possibly using better error handling.
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Fail to query"))
			return
		}

		queryResp := &queryResponse{
			RequestID: fulfillmentReq.RequestID,
		}
		queryResp.Payload.Devices = deviceState

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(queryResp)
		if err != nil {
			s.logger.Info("error serializing after writing ok",
				zap.Error(err),
			)
		}
		return
	case "action.devices.EXECUTE":
		s.provider.Sync(r.Context())

		w.Write([]byte("{}"))
		return
	case "action.devices.DISCONNECT":
		s.provider.Disconnect(r.Context())

		w.Write([]byte("{}"))
		return
	}

	s.logger.Info("unsupported intent name specified",
		zap.String("request_id", fulfillmentReq.RequestID),
		zap.String("intent", fulfillmentReq.Inputs[0].Intent),
	)

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Unsupported intent name specified"))
}

// fulfillmentRequest matches the request format documented at https://developers.google.com/assistant/smarthome/develop/process-intents
// It appears to be generated from a protobuf file but I was unable to locate the proper one.
type fulfillmentRequest struct {
	RequestID string             `json:"requestId"`
	Inputs    []fulfillmentInput `json:"inputs"`
}

// fulfillmentInput matches the intent format documented at https://developers.google.com/assistant/smarthome/reference/intent/sync (first of the 4 intents)
type fulfillmentInput struct {
	Intent string

	// based on the supplied intent one of the 2 below fields may be set
	Query   *queryPayload
	Execute *executePayload
}

func (i *fulfillmentInput) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Intent  string          `json:"intent"`
		Payload json.RawMessage `json:"payload"`
	}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	i.Intent = tmp.Intent
	switch tmp.Intent {
	case "action.devices.QUERY":
		payload := &queryPayload{}
		err = json.Unmarshal(tmp.Payload, payload)
		if err != nil {
			return err
		}
		i.Query = payload
	case "action.devices.EXECUTE":
		payload := &executePayload{}
		err = json.Unmarshal(tmp.Payload, payload)
		if err != nil {
			return err
		}
		i.Execute = payload

	}

	return nil
}

type queryPayload struct {
	Devices []struct {
		ID         string                 `json:"id"`
		CustomData map[string]interface{} `json:"customData"`
	} `json:"devices"`
}
type executePayload struct {
	Commands []struct {
		Devices []struct {
			ID         string                 `json:"id"`
			CustomData map[string]interface{} `json:"customData"`
		} `json:"devices"`
		Execution []struct {
			Command string          `json:"command"`
			Params  json.RawMessage `json:"params"`
		} `json:"execution"`
	} `json:"commands"`
}

type syncResponse struct {
	RequestID string `json:"requestId,omitempty"`
	Payload   struct {
		UserID    string    `json:"agentUserId,omitempty"`
		ErrorCode string    `json:"errorCode,omitempty"`
		DebugInfo string    `json:"debugString,omitempty"`
		Devices   []*Device `json:"devices,omitempty"`
	} `json:"payload"`
}

type queryResponse struct {
	RequestID string `json:"requestId,omitempty"`
	Payload   struct {
		Devices map[string]DeviceState `json:"devices"`
	} `json:"payload"`
}

// MarshalJSON is a custom JSON serializer for our Device
func (d Device) MarshalJSON() ([]byte, error) {
	dr := deviceRaw{}

	dr.ID = d.ID
	dr.Type = d.Type
	for trait := range d.Traits {
		dr.Traits = append(dr.Traits, trait)
	}
	dr.Name.DefaultNames = d.Name.DefaultNames
	dr.Name.Name = d.Name.Name
	dr.Name.Nicknames = d.Name.Nicknames
	dr.WillReportState = d.WillReportState
	dr.RoomHint = d.RoomHint
	dr.Attributes = d.attributes
	dr.DeviceInfo.Manufacturer = d.DeviceInfo.Manufacturer
	dr.DeviceInfo.Model = d.DeviceInfo.Model
	dr.DeviceInfo.HwVersion = d.DeviceInfo.HwVersion
	dr.DeviceInfo.SwVersion = d.DeviceInfo.SwVersion
	for _, otherDeviceID := range d.OtherDeviceIDs {
		dr.OtherDeviceIDs = append(dr.OtherDeviceIDs, otherDeviceIDraw{
			AgentID:  otherDeviceID.AgentID,
			DeviceID: otherDeviceID.DeviceID,
		})
	}
	dr.CustomData = d.CustomData

	return json.Marshal(dr)
}

type otherDeviceIDraw struct {
	AgentID  string `json:"agentId,omitempty"`
	DeviceID string `json:"deviceId,omitempty"`
}

// Device represents a single provider-supplied device profile.
type deviceRaw struct {
	// ID of the device
	ID string `json:"id,omitempty"`

	// Type of the device.
	// See https://developers.google.com/assistant/smarthome/guides is a list of possible types
	Type string `json:"type,omitempty"`

	// Traits of the device.
	// See https://developers.google.com/assistant/smarthome/traits for a list of possible traits
	// The set of assigned traits will dictate which actions can be performed on the device
	Traits []string `json:"traits,omitempty"`

	// Name of the device.
	Name struct {
		// DefaultNames (not user settable)
		DefaultNames []string `json:"defaultNames,omitempty"`
		// Name supplied by the user for display purposes
		Name string `json:"name,omitempty"`
		// Nicknames given to this, should a user have multiple ways to refer to the device
		Nicknames []string `json:"nicknames,omitempty"`
	} `json:"name,omitempty"`

	// WillReportState using the ReportState API (should be true)
	WillReportState bool `json:"willReportState"`

	// RoomHint guides Google as to which room this device is in
	RoomHint string `json:"roomHint,omitempty"`

	// Attributes linked to the defined traits
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// DeviceInfo that is physically defined
	DeviceInfo struct {
		Manufacturer string `json:"manufacturer,omitempty"`
		Model        string `json:"model,omitempty"`
		HwVersion    string `json:"hwVersion,omitempty"`
		SwVersion    string `json:"swVersion,omitempty"`
	} `json:"deviceInfo,omitempty"`

	OtherDeviceIDs []otherDeviceIDraw `json:"otherDeviceIds,omitempty"`

	// CustomData specified which will be included unmodified in subsequent requests.
	CustomData map[string]interface{} `json:"customData,omitempty"`
}
