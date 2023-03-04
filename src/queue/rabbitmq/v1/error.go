package v1

import "errors"

var (
	errFailedToStartPublisher = errors.New("failed to start publisher")
	ErrNoSuchConnection       = errors.New("can't find connection with provided name")
	ErrEmptyConnectionString  = errors.New("connections with the empty string is not allowed")
	//errFailedToStartConsumer  = errors.New("failed to start consumer")
)
