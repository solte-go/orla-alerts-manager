package gelf_test

import (
	"encoding/json"
	"io"
	"testing"

	"orla-alerts/solte.lab/src/logging/gelf"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAddr(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Addr("127.0.0.1"),
		gelf.Port(3000),
		gelf.Protocol("udp"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestHost(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Host("google.com"),
		gelf.Port(80),
		gelf.Protocol("udp"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestVersion(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Host("google.com"),
		gelf.Port(80),
		gelf.Protocol("udp"),
		gelf.Version("1.2"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestMessageKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.MessageKey("custom_message"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestLevelKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LevelKey("custom_level"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestTimeKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.TimeKey("custom_time"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestNameKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.NameKey("custom_name"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestCallerKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.CallerKey("custom_caller"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestFunctionKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.FunctionKey("custom_function"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestStacktraceKey(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.StacktraceKey("custom_stacktrace"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestSkipLineEnding(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.SkipLineEnding(true),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestLineEnding(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LineEnding("\r\n"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestEncodeDuration(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeDuration(zapcore.NanosDurationEncoder),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestEncodeCaller(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeCaller(zapcore.FullCallerEncoder),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestEncodeName(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.EncodeName(zapcore.FullNameEncoder),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestNewReflectedEncoder(t *testing.T) {
	var newEncoder = func(writer io.Writer) zapcore.ReflectedEncoder {
		return json.NewEncoder(writer)
	}
	var core, err = gelf.NewCore(
		gelf.NewReflectedEncoder(newEncoder),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestLevel(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.Level(zap.DebugLevel),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestLevelString(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.LevelString("debug"),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}

func TestChunkSize(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.ChunkSize(2000),
	)
	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}

	core, err = gelf.NewCore(
		gelf.ChunkSize(gelf.MaxChunkSize + 1),
	)
	assert.Equal(t, gelf.ErrChunkTooLarge, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")

	core, err = gelf.NewCore(
		gelf.ChunkSize(gelf.MinChunkSize - 1),
	)
	assert.Equal(t, gelf.ErrChunkTooSmall, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")
}

func TestCompressionType(t *testing.T) {
	var (
		err              error
		core             zapcore.Core
		compressionTypes = []int{
			gelf.CompressionNone,
			gelf.CompressionGzip,
			gelf.CompressionZlib,
		}
	)

	for _, compressionType := range compressionTypes {
		core, err = gelf.NewCore(
			gelf.CompressionType(compressionType),
		)
		if assert.NoError(t, err, "Unexpected error") {
			assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
		}
	}

	core, err = gelf.NewCore(
		gelf.CompressionType(13),
	)
	assert.Equal(t, gelf.ErrUnknownCompressionType, err, "Unexpected error")
	assert.Nil(t, core, "Expected nil")
}

func TestCompressionLevel(t *testing.T) {
	var core, err = gelf.NewCore(
		gelf.CompressionLevel(9),
	)

	if assert.NoError(t, err, "Unexpected error") {
		assert.Implements(t, (*zapcore.Core)(nil), core, "Expect zapcore.Core")
	}
}
