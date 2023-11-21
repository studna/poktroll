package polyzap_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzap"
	"github.com/pokt-network/poktroll/testutil/testpolylog"
)

var (
	expectedTime                   = time.Now()
	expectedTimestampEventContains = fmt.Sprintf(`"ts":%d.`, expectedTime.Unix())
	expectedTimeEventContains      = fmt.Sprintf(`"Time":%d.`, expectedTime.Unix())
	expectedDuration               = time.Millisecond + (250 * time.Nanosecond) // 1000250
	expectedDurationEventContains  = fmt.Sprintf(`"Dur":%f`, expectedDuration.Seconds())
)

func TestZapPolyLogger_AllLevels_AllEventMethods(t *testing.T) {
	tests := []testpolylog.EventMethodsTest{
		{
			Msg:                    "Msg",
			ExpectedOutputContains: "Msg",
		},
		{
			MsgFmt:                 "%s",
			MsgFmtArgs:             []any{"Msgf"},
			ExpectedOutputContains: "Msgf",
		},
		{
			Key:                    "Str",
			Value:                  "str_value",
			ExpectedOutputContains: `"Str":"str_value"`,
		},
		{
			Key:                    "Bool",
			Value:                  true,
			ExpectedOutputContains: `"Bool":true`,
		},
		{
			Key:                    "Int",
			Value:                  int(42),
			ExpectedOutputContains: `"Int":42`,
		},
		{
			Key:                    "Int8",
			Value:                  int8(42),
			ExpectedOutputContains: `"Int8":42`,
		},
		{
			Key:                    "Int16",
			Value:                  int16(42),
			ExpectedOutputContains: `"Int16":42`,
		},
		{
			Key:                    "Int32",
			Value:                  int32(42),
			ExpectedOutputContains: `"Int32":42`,
		},
		{
			Key:                    "Int64",
			Value:                  int64(42),
			ExpectedOutputContains: `"Int64":42`,
		},
		{
			Key:                    "Uint",
			Value:                  uint(42),
			ExpectedOutputContains: `"Uint":42`,
		},
		{
			Key:                    "Uint8",
			Value:                  uint8(42),
			ExpectedOutputContains: `"Uint8":42`,
		},
		{
			Key:                    "Uint16",
			Value:                  uint16(42),
			ExpectedOutputContains: `"Uint16":42`,
		},
		{
			Key:                    "Uint32",
			Value:                  uint32(42),
			ExpectedOutputContains: `"Uint32":42`,
		},
		{
			Key:                    "Uint64",
			Value:                  uint64(42),
			ExpectedOutputContains: `"Uint64":42`,
		},
		{
			Key:                    "Float32",
			Value:                  float32(420.69),
			ExpectedOutputContains: `"Float32":420.69`,
		},
		{
			Key:                    "Float64",
			Value:                  float64(420.69),
			ExpectedOutputContains: `"Float64":420.69`,
		},
		{
			EventMethodName:        "Err",
			Value:                  fmt.Errorf("%d", 42),
			ExpectedOutputContains: `"error":"42"`,
		},
		{
			EventMethodName:        "Timestamp",
			ExpectedOutputContains: expectedTimestampEventContains,
		},
		{
			Key:                    "Time",
			Value:                  expectedTime,
			ExpectedOutputContains: expectedTimeEventContains,
		},
		{
			Key:                    "Dur",
			Value:                  expectedDuration,
			ExpectedOutputContains: expectedDurationEventContains,
		},
		{
			EventMethodName: "Fields",
			Value: map[string]any{
				"key1": "value1",
				"key2": 42,
			},
			// TODO_IMPROVE: assert on all key/value pairs. Zap doesn't seem to
			// provide any guarantee around the oder of the fields. This requires
			// changing the test and helper structure to support this.
			ExpectedOutputContains: `"key2":42`,
		},
		{
			EventMethodName: "Fields",
			Value:           []any{"key1", "value1", "key2", 42},
			// TODO_IMPROVE: assert on all key/value pairs. Zap doesn't seem to
			// provide any guarantee around the oder of the fields. This requires
			// changing the test and helper structure to support this.
			ExpectedOutputContains: `"key2":42`,
		},
	}

	levels := []polyzap.Level{
		polyzap.DebugLevel,
		polyzap.InfoLevel,
		polyzap.WarnLevel,
		polyzap.ErrorLevel,
	}

	// TODO_IN_THIS_COMMIT: comment...
	for _, level := range levels {
		testpolylog.RunEventMethodTests(
			t,
			level,
			tests,
			newTestLogger,
			newTestEventWithLevel,
			"*polyzap.zapEvent",
			getExpectedLevelOutputContains,
		)
	}
}

// TODO_TEST: that exactly all expected levels log at each level.

// TODO_TEST: #Enabled() and #Discard()

// TODO_TEST/TODO_COMMUNITY: implement polyzap.Logger#With() such.
func TestZerologLogger_With(t *testing.T) {
	t.SkipNow()

	logger, logOutput := newTestLogger(t, polyzap.DebugLevel)

	logger.Debug().Msg("before")
	require.Contains(t, logOutput.String(), "before")

	logger = logger.With("key", "value")

	logger.Debug().Msg("after")
	require.Contains(t, logOutput.String(), "after")
	require.Contains(t, logOutput.String(), `"key":"value"`)
}

// TODO_TEST/TODO_COMMUNITY: test-drive (TDD) out `polyzap.Logger#WithContext()`.
func TestZerologLogger_WithContext(t *testing.T) {
	t.SkipNow()
}

func TestZerologLogger_WithLevel(t *testing.T) {
	logger, logOutput := newTestLogger(t, polyzap.DebugLevel)
	logger.WithLevel(polyzap.DebugLevel).Msg("WithLevel()")

	require.Contains(t, logOutput.String(), "WithLevel()")
}

func TestZerologLogger_Write(t *testing.T) {
	testOutput := "Write()"
	logger, logOutput := newTestLogger(t, polyzap.DebugLevel)

	n, err := logger.Write([]byte(testOutput))
	require.NoError(t, err)
	require.Lenf(t, testOutput, n, "expected %d bytes to be written", len(testOutput))

	require.Contains(t, logOutput.String(), testOutput)
}

func newTestLogger(t *testing.T, level polylog.Level) (polylog.Logger, *bytes.Buffer) {
	t.Helper()

	// Redirect standard log output to logOutput buffer.
	logOutput := new(bytes.Buffer)
	opts := []polylog.LoggerOption{
		polyzap.WithOutput(logOutput),
		polyzap.WithLevel(polyzap.Level(level.Int())),
	}

	logger := polyzap.NewLogger(opts...)

	return logger, logOutput
}

func newTestEventWithLevel(
	t *testing.T,
	logger polylog.Logger,
	level polylog.Level,
) polylog.Event {
	t.Helper()

	switch level.String() {
	case zap.DebugLevel.String():
		return logger.Debug()
	case zap.InfoLevel.String():
		return logger.Info()
	case zap.WarnLevel.String():
		return logger.Warn()
	case zap.ErrorLevel.String():
		return logger.Error()
	default:
		panic(fmt.Errorf("level not yet supported: %s", level.String()))
	}
}

func getExpectedLevelOutputContains(level polylog.Level) string {
	return fmt.Sprintf(`"level":%q`, level.String())
}
