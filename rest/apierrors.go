package rest

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	// Code indicates semantics based on HTTP status codes
	Code         int   `json:"code"`
	WrappedError error `json:"error"`
}

func (pattern APIError) MarshalJSON() ([]byte, error) {
	intermediary := struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}{
		Code:  pattern.Code,
		Error: pattern.WrappedError.Error(),
	}
	bytes, err := json.Marshal(intermediary)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (pattern APIError) UnWrap() error {
	return pattern.WrappedError
}

func (pattern APIError) Error() string {
	return fmt.Sprintf("APIError: %s", pattern.WrappedError.Error())
}
