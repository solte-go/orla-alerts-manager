package gelf

import (
	"compress/gzip"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net"
	"time"
)

const (
	// MinChunkSize minimal chunk size in bytes.
	MinChunkSize = 512

	// MaxChunkSize maximal chunk size in bytes.
	MaxChunkSize = 8192

	// MaxChunkCount maximal chunk per message count.
	MaxChunkCount = 128

	// DefaultChunkSize is default WAN chunk size.
	DefaultChunkSize = 1420

	// CompressionNone don't use compression.
	CompressionNone = 0

	// CompressionGzip use gzip compression.
	CompressionGzip = 1

	// CompressionZlib use zlib compression.
	CompressionZlib = 2
)

// Option interface.
type Option interface {
	apply(conf *optionConfiguration) error
}

// optionConfiguration.
type optionConfiguration struct {
	addr             string
	host             string
	port             int
	protocol         string
	version          string
	enabler          zap.AtomicLevel
	encoder          zapcore.EncoderConfig
	chunkSize        int
	writeSyncers     []zapcore.WriteSyncer
	compressionType  int
	compressionLevel int
}

// implement zapcore.Core.
type wrappedCore struct {
	core zapcore.Core
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(conf *optionConfiguration) error

// NewCore zap core constructor.
func NewCore(options ...Option) (_ zapcore.Core, err error) {
	var conf = optionConfiguration{
		addr:     "127.0.0.1",
		host:     "localhost",
		port:     80,
		protocol: "tcp",
		encoder: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			NameKey:        "_logger",
			LevelKey:       "level",
			CallerKey:      "_caller",
			MessageKey:     "short_message",
			StacktraceKey:  "full_message",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeName:     zapcore.FullNameEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeLevel:    levelEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
		version:          "1.1",
		enabler:          zap.NewAtomicLevel(),
		chunkSize:        DefaultChunkSize,
		writeSyncers:     make([]zapcore.WriteSyncer, 0, 8),
		compressionType:  CompressionGzip,
		compressionLevel: gzip.BestCompression,
	}

	for _, option := range options {
		if err = option.apply(&conf); err != nil {
			return nil, err
		}
	}

	var w = &writer{
		chunkSize:        conf.chunkSize,
		chunkDataSize:    conf.chunkSize - 12, // chunk size - chunk header size
		compressionType:  conf.compressionType,
		compressionLevel: conf.compressionLevel,
		protocol:         conf.protocol,
	}

	dialer := net.Dialer{
		Timeout:       0,
		Deadline:      time.Time{},
		FallbackDelay: 0,
		KeepAlive:     0,
		Resolver:      nil,
		Control:       nil,
	}

	//TODO FIX before use
	//fmt.Printf("%s  %s:%d\n", conf.protocol, conf.addr, conf.port)
	if w.conn, _ = dialer.Dial(conf.protocol, fmt.Sprintf("%s:%d", conf.addr, conf.port)); err != nil {
		//return nil, err
	}

	var core = zapcore.NewCore(
		zapcore.NewJSONEncoder(conf.encoder),
		zapcore.AddSync(w),
		conf.enabler,
	)

	return &wrappedCore{
		core: core.With([]zapcore.Field{
			zap.String("host", conf.host),
			zap.String("version", conf.version),
		}),
	}, nil
}

// apply implements Option.
func (f optionFunc) apply(conf *optionConfiguration) error {
	return f(conf)
}

// Addr set GELF address.
func Addr(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.addr = value
		return nil
	})
}

// Host set GELF host.
func Host(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.host = value
		return nil
	})
}

// Protocol set GELF host.
func Protocol(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.protocol = value
		return nil
	})
}

// Port set GELF host.
func Port(value int) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.port = value
		return nil
	})
}

// Version set GELF version.
func Version(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.version = value
		return nil
	})
}

// MessageKey set zapcore.EncoderConfig MessageKey property.
func MessageKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.MessageKey = escapeKey(value)
		return nil
	})
}

// LevelKey set zapcore.EncoderConfig LevelKey property.
func LevelKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.LevelKey = escapeKey(value)
		return nil
	})
}

// TimeKey set zapcore.EncoderConfig TimeKey property.
func TimeKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.TimeKey = escapeKey(value)
		return nil
	})
}

// NameKey set zapcore.EncoderConfig NameKey property.
func NameKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.NameKey = escapeKey(value)
		return nil
	})
}

// CallerKey set zapcore.EncoderConfig CallerKey property.
func CallerKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.CallerKey = escapeKey(value)
		return nil
	})
}

// FunctionKey set zapcore.EncoderConfig FunctionKey property.
func FunctionKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.FunctionKey = escapeKey(value)
		return nil
	})
}

// StacktraceKey set zapcore.EncoderConfig StacktraceKey property.
func StacktraceKey(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.StacktraceKey = escapeKey(value)
		return nil
	})
}

// SkipLineEnding set zapcore.EncoderConfig SkipLineEnding property.
func SkipLineEnding(value bool) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.SkipLineEnding = value
		return nil
	})
}

// LineEnding set zapcore.EncoderConfig LineEnding property.
func LineEnding(value string) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.LineEnding = value
		return nil
	})
}

// EncodeDuration set zapcore.EncoderConfig EncodeDuration property.
func EncodeDuration(value zapcore.DurationEncoder) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.EncodeDuration = value
		return nil
	})
}

// EncodeCaller set zapcore.EncoderConfig EncodeCaller property.
func EncodeCaller(value zapcore.CallerEncoder) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.EncodeCaller = value
		return nil
	})
}

// EncodeName set zapcore.EncoderConfig EncodeName property.
func EncodeName(value zapcore.NameEncoder) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.EncodeName = value
		return nil
	})
}

// NewReflectedEncoder set zapcore.EncoderConfig NewReflectedEncoder property.
func NewReflectedEncoder(value func(io.Writer) zapcore.ReflectedEncoder) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.encoder.NewReflectedEncoder = value
		return nil
	})
}

// Level set logging level.
func Level(value zapcore.Level) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.enabler.SetLevel(value)
		return nil
	})
}

// LevelString set logging level.
func LevelString(value string) Option {
	return optionFunc(func(conf *optionConfiguration) (err error) {
		err = conf.enabler.UnmarshalText([]byte(value))
		return err
	})
}

// CompressionType set GELF compression type.
func CompressionType(value int) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		switch value {
		case CompressionNone, CompressionGzip, CompressionZlib:
		default:
			return ErrUnknownCompressionType
		}

		conf.compressionType = value

		return nil
	})
}

// CompressionLevel set GELF compression level.
func CompressionLevel(value int) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		conf.compressionLevel = value
		return nil
	})
}

// levelEncoder maps the zap log levels to the gelf levels.
func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendInt(7)
	case zapcore.InfoLevel:
		enc.AppendInt(6)
	case zapcore.WarnLevel:
		enc.AppendInt(4)
	case zapcore.ErrorLevel:
		enc.AppendInt(3)
	case zapcore.DPanicLevel:
		enc.AppendInt(0)
	case zapcore.PanicLevel:
		enc.AppendInt(0)
	case zapcore.FatalLevel:
		enc.AppendInt(0)
	}
}

// Enabled implementation of zapcore.Core.
func (w *wrappedCore) Enabled(l zapcore.Level) bool {
	return w.core.Enabled(l)
}

// With implementation of zapcore.Core.
func (w *wrappedCore) With(fields []zapcore.Field) zapcore.Core {
	return &wrappedCore{core: w.core.With(w.escape(fields))}
}

// Write implementation of zapcore.Core.
func (w *wrappedCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	return w.core.Write(e, w.escape(fields))
}

// Sync implementation of zapcore.Core.
func (w *wrappedCore) Sync() error {
	return w.core.Sync()
}

// Check implementation of zapcore.Core.
func (w *wrappedCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if w.Enabled(e.Level) {
		return ce.AddCore(e, w)
	}
	return ce
}

// escape prefixed additional gelf fields.
func (w *wrappedCore) escape(fields []zapcore.Field) []zapcore.Field {
	if len(fields) == 0 {
		return fields
	}

	var escaped = make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		field.Key = escapeKey(field.Key)
		escaped = append(escaped, field)
	}

	return escaped
}

// escapeKey append prefix to additional field keys.
func escapeKey(value string) string {
	switch value {
	case "id":
		return "__id"
	case "version", "host", "short_message", "full_message", "timestamp", "level":
		return value
	}

	if len(value) == 0 {
		return "_"
	}

	if value[0] == '_' {
		return value
	}

	return "_" + value
}

// ChunkSize set GELF chunk size.
func ChunkSize(value int) Option {
	return optionFunc(func(conf *optionConfiguration) error {
		if value < MinChunkSize {
			return ErrChunkTooSmall
		}

		if value > MaxChunkSize {
			return ErrChunkTooLarge
		}

		conf.chunkSize = value

		return nil
	})
}
