package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
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
