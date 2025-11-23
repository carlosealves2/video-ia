package servicediscovery

import "errors"

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrConnectionFailed = errors.New("connection to service discovery failed")
	ErrTimeout         = errors.New("request timeout")
)
