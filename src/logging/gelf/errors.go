package gelf

import "errors"

var (
	// ErrChunkTooSmall triggered when chunk size to small.
	ErrChunkTooSmall = errors.New("chunk size too small")

	// ErrChunkTooLarge triggered when chunk size too large.
	ErrChunkTooLarge = errors.New("chunk size too large")

	// ErrUnknownCompressionType triggered when passed invalid compression type.
	ErrUnknownCompressionType = errors.New("unknown compression type")

	// chunkedMagicBytes chunked message magic bytes.
	chunkedMagicBytes = []byte{0x1e, 0x0f}
)
