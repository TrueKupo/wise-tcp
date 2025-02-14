package log

import (
	"bytes"
	std "log"
	"testing"
)

func setupTestLogger() (*defaultLogger, *bytes.Buffer) {
	var output bytes.Buffer
	std.SetOutput(&output)
	std.SetFlags(0)

	logger := Default().(*defaultLogger)

	return logger, &output
}

func assertLogOutput(t *testing.T, got, want string) {
	if got != want {
		t.Errorf("unexpected log output:\n got: %q\nwant: %q", got, want)
	}
}

func TestDefaultLogger_LogLevels(t *testing.T) {
	logger, output := setupTestLogger()

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		expected string
	}{
		{"Info", logger.Info, "info message", "[INFO] info message\n"},
		{"Debug", logger.Debug, "debug message", "[DEBUG] debug message\n"},
		{"Warn", logger.Warn, "warn message", "[WARN] warn message\n"},
		{"Error", logger.Error, "error message", "[ERROR] error message\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.message)
			got := output.String()
			assertLogOutput(t, got, tt.expected)
			output.Reset()
		})
	}
}

func TestDefaultLogger_Fatal(t *testing.T) {
	logger, output := setupTestLogger()

	var exitCalled bool
	mockExitHandler := func(code int) {
		exitCalled = true
	}

	logger.exitHandler = mockExitHandler

	logger.Fatal("fatal message")

	got := output.String()
	want := "[FATAL] fatal message\n"
	assertLogOutput(t, got, want)

	if !exitCalled {
		t.Errorf("expected exit handler to be called")
	}
}

func TestDefaultLogger_ConcurrentLogging(t *testing.T) {
	logger, output := setupTestLogger()

	const numRoutines = 10
	const numMessages = 100
	done := make(chan struct{}, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			for j := 0; j < numMessages; j++ {
				logger.Info("goroutine %d: message %d", id, j)
			}
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < numRoutines; i++ {
		<-done
	}

	logLines := bytes.Split(output.Bytes(), []byte("\n"))
	expectedLines := numRoutines * numMessages

	if len(logLines)-1 != expectedLines {
		t.Errorf("expected %d log lines, got %d", expectedLines, len(logLines)-1)
	}
}

func TestDefaultLogger_EmptyMessage(t *testing.T) {
	logger, output := setupTestLogger()

	logger.Info("")
	assertLogOutput(t, output.String(), "[INFO] \n")
}

func TestDefaultLogger_FormatMismatch(t *testing.T) {
	logger, output := setupTestLogger()

	logger.Debug("debug with args: %d", 42)
	assertLogOutput(t, output.String(), "[DEBUG] debug with args: 42\n")
	output.Reset()

	logger.Warn("warn with args: %d")
	assertLogOutput(t, output.String(), "[WARN] warn with args: %!d(MISSING)\n")
}
