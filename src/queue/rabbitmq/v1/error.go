package v1

import "errors"

var (
    errFailedToStartPublisher  = errors.New("failed to start publisher")
    errFailedToStartConsumer   = errors.New("failed to start consumer")
    ErrNoSuchConnection        = errors.New("can't find connection with provided name")
    ErrEmptyConnectionString   = errors.New("connections with the empty string is not allowed")
    ErrQueueIsNotConfigured    = errors.New("the delivery channel has not been configured")
    ErrQueueIsNotReady         = errors.New("the delivery channel is empty, check if the are messages in queue")
    ErrCantReconnectWithServer = errors.New("error when trying to reconnect to rabbitMQ")
)
