package polystd_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polystd"
	"github.com/pokt-network/poktroll/testutil/testpolylog"
)

var (
	expectedTime                   = time.Now()
	expectedTimestampLayout        = "2006-01-02T15:04:05-07:00"
	expectedTimestampEventContains = fmt.Sprintf(`"time":"%s"`, expectedTime.Format(expectedTimestampLayout))
	expectedTimeEventContains      = fmt.Sprintf(`"Time":"%s"`, expectedTime.Format(expectedTimestampLayout))
	expectedDuration               = time.Millisecond + (250 * time.Nanosecond) // 1000250

	expectedDurationEventContains = fmt.Sprintf(`"Dur":%q`, expectedDuration.String()) // 1.00025ms
)

func TestStdLogger_AllLevels_AllEventMethods(t *testing.T) {
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
			// TODO_IMPROVE: assert on all key/value pairs. go doesn't guarantee
			// iteration oder of map key/value pairs. This requires promoting this
			// case to its own test or refactoring and/or restructuring test and
			// helper to support this.
			ExpectedOutputContains: `"key2":42`,
		},
		{
			EventMethodName: "Fields",
			Value:           []any{"key1", "value1", "key2", 42},
			// TODO_IMPROVE: assert on all key/value pairs. go doesn't guarantee
			// iteration oder of the slice (?). This requires promoting this
			// case to its own test or refactoring and/or restructuring test and
			// helper to support this.
			ExpectedOutputContains: `"key2":42`,
		},
	}

	levels := []polystd.Level{
		polystd.DebugLevel,
		polystd.InfoLevel,
		polystd.WarnLevel,
		polystd.ErrorLevel,
	}

	// TODO_IN_THIS_COMMIT: comment...
	for _, level := range levels {
		testpolylog.RunEventMethodTests(
			t,
			level,
			tests,
			newTestLogger,
			newTestEventWithLevel,
			"*polystd.stdLogEvent",
			getExpectedLevelOutputContains,
		)
	}
}

// TODO_TEST/TODO_COMMUNITY: assert that exactly all expected levels log at
// each level.

// TODO_TEST/TODO_COMMUNITY: assert that #Enabled() returns false after
// #Discard() has ben called but not before.

// TODO_TEST/TODO_COMMUNITY: implement polystd.Logger#With() such.
func TestZerologLogger_With(t *testing.T) {
	t.SkipNow()

	logger, logOutput := newTestLogger(t, polystd.DebugLevel)

	logger.Debug().Msg("before")
	require.Contains(t, logOutput.String(), "before")

	logger = logger.With("key", "value")

	logger.Debug().Msg("after")
	require.Contains(t, logOutput.String(), "after")
	require.Contains(t, logOutput.String(), `"key":"value"`)
}

// TODO_TEST/TODO_COMMUNITY: test-drive (TDD) out `polystd.Logger#WithContext()`.
func TestZerologLogger_WithContext(t *testing.T) {
	t.SkipNow()
}

func TestZerologLogger_WithLevel(t *testing.T) {
	logger, logOutput := newTestLogger(t, polystd.DebugLevel)
	logger.WithLevel(polystd.DebugLevel).Msg("WithLevel()")

	require.Contains(t, logOutput.String(), "WithLevel()")
}

func TestZerologLogger_Write(t *testing.T) {
	testOutput := "Write()"
	logger, logOutput := newTestLogger(t, polystd.DebugLevel)

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
		polystd.WithOutput(logOutput),
		// NB: typically consumers would use polystd.<some>Level directly instead
		// of casting like this.
		polystd.WithLevel(polystd.Level(level.Int())),
	}

	// TODO_IN_THIS_COMMIT: configuration ... debug level for this test
	logger := polystd.NewLogger(opts...)

	return logger, logOutput
}

// TODO_TEST: that exactly all expected levels log at each level.

// TODO_TEST: #Enabled() and #Discard()

func newTestEventWithLevel(
	t *testing.T,
	logger polylog.Logger,
	level polylog.Level,
) polylog.Event {
	t.Helper()

	// Match on level string to determine which method to call on the logger.
	switch level.String() {
	case polystd.DebugLevel.String():
		return logger.Debug()
	case polystd.InfoLevel.String():
		return logger.Info()
	case polystd.WarnLevel.String():
		return logger.Warn()
	case polystd.ErrorLevel.String():
		return logger.Error()
	default:
		panic(fmt.Errorf("level not yet supported: %s", level.String()))
	}
}

func getExpectedLevelOutputContains(level polylog.Level) string {
	return fmt.Sprintf(`[%s]`, strings.ToUpper(level.String()))
}
