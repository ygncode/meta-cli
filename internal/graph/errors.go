package graph

import (
	"errors"
	"fmt"
)

type GraphError struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      int    `json:"code"`
	FBTraceID string `json:"fbtrace_id"`
}

func (e *GraphError) Error() string {
	return fmt.Sprintf("graph api error %d: %s (type=%s)", e.Code, e.Message, e.Type)
}

type APIError struct {
	StatusCode int
	Graph      *GraphError
}

func (e *APIError) Error() string {
	if e.Graph != nil {
		return e.Graph.Error()
	}
	return fmt.Sprintf("http error %d", e.StatusCode)
}

func (e *APIError) Unwrap() error {
	if e.Graph != nil {
		return e.Graph
	}
	return nil
}

func IsTokenExpired(err error) bool {
	var ge *GraphError
	return errors.As(err, &ge) && ge.Code == 190
}

func IsPermissionDenied(err error) bool {
	var ge *GraphError
	return errors.As(err, &ge) && (ge.Code == 200 || ge.Code == 10)
}
