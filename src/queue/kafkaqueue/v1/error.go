package kafkaQueueV1

import (
	"github.com/pkg/errors"
)

var (
	contextExceeded                 = errors.New("context exceeded")
	ErrNoConnectionWithProvidedName = errors.New("no connection with provided name")
	ErrEmptyConnectionName          = errors.New("empty connection name not supported")
)
