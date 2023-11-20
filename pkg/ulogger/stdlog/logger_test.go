package stdlog_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/poktroll/pkg/ulogger"
	"github.com/pokt-network/poktroll/pkg/ulogger/stdlog"
)

var expectedMsgs = []string{
	"Msg",
	"Msgf",
	"Str=str_value",
	"Bool=true",
	"Int=42",
	//"Int8=42",
	//"Int16=42",
	//"Int32=42",
	//"Int64=42",
	//"Uint=42",
	//"Uint8=42",
	//"Uint16=42",
	//"Uint32=42",
	//"Uint64=42",
	//"Float32=420.69",
	//"Float64=420.69",
	//"Err=42",
	//"Func=0x",
	//"Timestamp=1700509048049",
	//"Time=2006-01-02T15:04:05Z07:00",
	//"Dur=",
	//"Fields=map[key1:value1 key2:value2]",
}

// TODO_IN_THIS_COMMIT: comment...
type funcMethodSpy struct{ mock.Mock }

// TODO_IN_THIS_COMMIT: comment...
func (m *funcMethodSpy) Fn(event ulogger.Event) {
	m.Called(event)
}

func TestStdLogULogger(t *testing.T) {
	// Redirect standard log output to logOutput buffer.
	logOutput := new(bytes.Buffer)
	log.SetOutput(logOutput)

	defer func() {
		// Reset output to default after test>
		log.SetOutput(os.Stderr)
	}()

	// TODO_IN_THIS_COMMIT: configuration ... debug levelString for this test
	logger := stdlog.NewUniversalLogger()

	logger.Debug().Msg("Msg")
	logger.Debug().Msgf("%s", "Msgf")
	logger.Debug().Str("Str", "str_value").Send()
	logger.Debug().Bool("Bool", true).Send()
	logger.Debug().Int("Int", 42).Send()
	//logger.Debug().Int8("Int8", 42).Send()
	//logger.Debug().Int16("Int16", 42).Send()
	//logger.Debug().Int32("Int32", 42).Send()
	//logger.Debug().Int64("Int64", 42).Send()
	//logger.Debug().Uint("Uint", 42).Send()
	//logger.Debug().Uint8("Uint8", 42).Send()
	//logger.Debug().Uint16("Uint16", 42).Send()
	//logger.Debug().Uint32("Uint32", 42).Send()
	//logger.Debug().Uint64("Uint64", 42).Send()
	//logger.Debug().Float32("Float32", 420.69).Send()
	//logger.Debug().Float64("Float64", 420.69).Send()
	//logger.Debug().Err(fmt.Errorf("%d", 42)).Send()
	//logger.Debug().Timestamp().Send()
	//logger.Debug().Time().Send()
	//logger.Debug().Dur().Send()
	//logger.Debug().Fields(map[string]string{
	//	"key1": "value1",
	//	"key2": "value2",
	//}).Send()

	// TODO_IN_THIS_COMMIT: comment...
	funcSpy := funcMethodSpy{}
	//logger.Debug().Func(funcSpy.Fn).Send()

	// TODO:
	// .Enabled()
	// .Discard()

	// Assert that the log output contains the expected messages. Split the log
	// output into lines and iterate over them.
	lines := strings.Split(logOutput.String(), "\n")
	lines = lines[:len(lines)-1] // Remove last empty line.
	// Assert that the log output contains the expected number of lines.
	// Intentionally not using `require` to provide additional error context.
	assert.Lenf(
		t, lines,
		len(expectedMsgs),
		"log output should contain %d lines, got: %d",
		len(expectedMsgs), len(lines),
	)

	for lineIdx, line := range lines {
		// Skip empty lines.
		//if line == "" {
		//	continue
		//}

		if strings.Contains(line, "Func=0x") {
			// Assert that the Func field contains the expected value.
			// TODO_IMPROVE: add coverage of an event which is disabled,
			// asserting that `Fn` is not called.
			funcSpy.AssertCalled(t, "Fn")
			continue
		}

		// Assert that each line contains the expected prefix.
		require.Contains(t, line, `[DEBUG] `)

		expectedMsg := expectedMsgs[lineIdx]
		require.Contains(t, line, expectedMsg)
	}
}

func TestSanity(t *testing.T) {
	logOutput := new(bytes.Buffer)
	logger := zerolog.New(logOutput)
	logger.Debug().Msg("hello")
	t.Log("logOutput:", logOutput.String())
}
