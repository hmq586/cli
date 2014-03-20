package net

import (
	"cf/configuration"
	"cf/errors"
	"encoding/json"
)

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

var uaaErrorHandler = func(statusCode int, body []byte) error {
	response := uaaErrorResponse{}
	json.Unmarshal(body, &response)

	if response.Code == "invalid_token" {
		return errors.NewInvalidTokenError(response.Description)
	} else {
		return errors.NewHttpError(statusCode, response.Code, response.Description)
	}
}

func NewUAAGateway(config configuration.Reader) Gateway {
	return newGateway(uaaErrorHandler, config)
}
