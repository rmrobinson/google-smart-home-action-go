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
		pSyncResp, err := s.provider.Sync(r.Context(), userID)
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
		syncResp.Payload.Devices = pSyncResp.Devices

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
		pQueryReq := &QueryRequest{
			AgentID: userID,
		}
		for _, device := range fulfillmentReq.Inputs[0].Query.Devices {
			pQueryReq.Devices = append(pQueryReq.Devices, DeviceArg{
				ID:         device.ID,
				CustomData: device.CustomData,
			})
		}

		pQueryResp, err := s.provider.Query(r.Context(), pQueryReq)
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
		queryResp.Payload.Devices = map[string]DeviceState{}
		for deviceID, state := range pQueryResp.States {
			state.Status = "SUCCESS"
			queryResp.Payload.Devices[deviceID] = state
		}

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
		pExecuteReq := &ExecuteRequest{
			AgentID: userID,
		}
		for _, command := range fulfillmentReq.Inputs[0].Execute.Commands {
			devices := []DeviceArg{}
			for _, device := range command.Devices {
				devices = append(devices, DeviceArg{
					ID:         device.ID,
					CustomData: device.CustomData,
				})
			}
			pExecuteReq.Commands = append(pExecuteReq.Commands, CommandArg{
				TargetDevices: devices,
				Commands:      command.Execution,
			})
		}

		pExecuteResp, err := s.provider.Execute(r.Context(), pExecuteReq)
		if err != nil {
			s.logger.Info("execute error",
				zap.Error(err),
			)

			// TODO: clean this up possibly using better error handling.
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Fail to execute"))
			return
		}

		executeResp := &executeResponse{
			RequestID: fulfillmentReq.RequestID,
		}

		if len(pExecuteResp.UpdatedDevices) > 0 {
			commandSuccessResp := executeRespPayload{
				Status: "SUCCESS",
				States: pExecuteResp.UpdatedState.State,
			}
			commandSuccessResp.States["online"] = true
			for _, id := range pExecuteResp.UpdatedDevices {
				commandSuccessResp.IDs = append(commandSuccessResp.IDs, id)
			}

			executeResp.Payload.Commands = append(executeResp.Payload.Commands, commandSuccessResp)
		}

		if len(pExecuteResp.OfflineDevices) > 0 {
			commandOfflineResp := executeRespPayload{
				Status: "OFFLINE",
			}
			for _, id := range pExecuteResp.OfflineDevices {
				commandOfflineResp.IDs = append(commandOfflineResp.IDs, id)
			}

			executeResp.Payload.Commands = append(executeResp.Payload.Commands, commandOfflineResp)
		}

		for errCode, details := range pExecuteResp.FailedDevices {
			commandFailResp := executeRespPayload{
				Status:    "ERROR",
				ErrorCode: errCode,
			}
			for _, id := range details.Devices {
				commandFailResp.IDs = append(commandFailResp.IDs, id)
			}

			executeResp.Payload.Commands = append(executeResp.Payload.Commands, commandFailResp)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(executeResp)
		if err != nil {
			s.logger.Info("error serializing after writing ok",
				zap.Error(err),
			)
		}
		return
	case "action.devices.DISCONNECT":
		s.provider.Disconnect(r.Context(), userID)

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

type deviceHandle struct {
	ID         string                 `json:"id"`
	CustomData map[string]interface{} `json:"customData"`
}

type queryPayload struct {
	Devices []deviceHandle `json:"devices"`
}
type executePayload struct {
	Commands []struct {
		Devices   []deviceHandle `json:"devices"`
		Execution []Command      `json:"execution"`
	} `json:"commands"`
}
type executeRespPayload struct {
	IDs       []string               `json:"ids,omitempty"`
	Status    string                 `json:"status,omitempty"`
	ErrorCode string                 `json:"errorCode,omitempty"`
	States    map[string]interface{} `json:"states,omitempty"`
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
type executeResponse struct {
	RequestID string `json:"requestId,omitempty"`
	Payload   struct {
		Commands []executeRespPayload `json:"commands"`
	} `json:"payload"`
}
