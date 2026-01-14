package logger

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestCallerTagging_DebugfIncludesFunction(t *testing.T) {
	var buf bytes.Buffer
	// Replace the Debug logger to capture output
	Debug = log.New(&buf, "", 0)
	enabledLevels[DebugLevel] = true
	prevInclude := includeCallerTag
	includeCallerTag = true
	defer func() { includeCallerTag = prevInclude }()

	Debugf("hello")

	out := buf.String()
	if !strings.Contains(out, "TestCallerTagging_DebugfIncludesFunction") {
		t.Fatalf("expected caller name in output, got: %q", out)
	}
}

func TestCallerTagging_DefaultOff(t *testing.T) {
	var buf bytes.Buffer
	Info = log.New(&buf, "", 0)
	enabledLevels[InfoLevel] = true

	prevInclude := includeCallerTag
	includeCallerTag = false
	defer func() { includeCallerTag = prevInclude }()

	Infof("hello")

	out := buf.String()
	if strings.Contains(out, "TestCallerTagging_DefaultOff") {
		t.Fatalf("expected caller info to be omitted, got: %q", out)
	}
}

func TestStructuredLogging_InfoKV(t *testing.T) {
	var buf bytes.Buffer
	Info = log.New(&buf, "", 0)
	enabledLevels[InfoLevel] = true

	InfoKV("test message", "key1", "value1", "key2", 42)

	out := buf.String()
	if !strings.Contains(out, "test message") {
		t.Fatalf("expected message in output, got: %q", out)
	}
	if !strings.Contains(out, "key1=value1") {
		t.Fatalf("expected key1=value1 in output, got: %q", out)
	}
	if !strings.Contains(out, "key2=42") {
		t.Fatalf("expected key2=42 in output, got: %q", out)
	}
}

func TestStructuredLogging_ErrorKV(t *testing.T) {
	var buf bytes.Buffer
	Error = log.New(&buf, "", 0)
	enabledLevels[ErrorLevel] = true

	ErrorKV("connection failed", "host", "localhost", "port", 5432)

	out := buf.String()
	if !strings.Contains(out, "connection failed") {
		t.Fatalf("expected message in output, got: %q", out)
	}
	if !strings.Contains(out, "host=localhost") {
		t.Fatalf("expected host=localhost in output, got: %q", out)
	}
	if !strings.Contains(out, "port=5432") {
		t.Fatalf("expected port=5432 in output, got: %q", out)
	}
}

func TestLevelFiltering_DisableDebug(t *testing.T) {
	var buf bytes.Buffer
	Debug = log.New(&buf, "", 0)
	Info = log.New(&buf, "", 0)

	// Disable DEBUG level
	enabledLevels = map[Level]bool{
		DebugLevel: false,
		InfoLevel:  true,
		WarnLevel:  true,
		ErrorLevel: true,
	}

	Debugf("should not appear")
	Infof("should appear")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Fatalf("debug message should be filtered out, got: %q", out)
	}
	if !strings.Contains(out, "should appear") {
		t.Fatalf("info message should appear, got: %q", out)
	}
}

func TestLevelFiltering_OnlyErrors(t *testing.T) {
	var buf bytes.Buffer
	Debug = log.New(&buf, "", 0)
	Info = log.New(&buf, "", 0)
	Warning = log.New(&buf, "", 0)
	Error = log.New(&buf, "", 0)

	// Only ERROR level enabled
	enabledLevels = map[Level]bool{
		DebugLevel: false,
		InfoLevel:  false,
		WarnLevel:  false,
		ErrorLevel: true,
	}

	Debugf("debug msg")
	Infof("info msg")
	Warnf("warn msg")
	Errorf("error msg")

	out := buf.String()
	if strings.Contains(out, "debug msg") || strings.Contains(out, "info msg") || strings.Contains(out, "warn msg") {
		t.Fatalf("only error should appear, got: %q", out)
	}
	if !strings.Contains(out, "error msg") {
		t.Fatalf("error message should appear, got: %q", out)
	}
}

func TestParseLevels_EmptyString(t *testing.T) {
	levels := parseLevels("")
	if !levels[DebugLevel] || !levels[InfoLevel] || !levels[WarnLevel] || !levels[ErrorLevel] ||
		!levels[NoticeLevel] || !levels[CritLevel] || !levels[AlertLevel] || !levels[EmergLevel] ||
		!levels[FatalLevel] {
		t.Fatalf("empty string should enable all levels, got: %+v", levels)
	}
}

func TestParseLevels_InfoAndError(t *testing.T) {
	levels := parseLevels("INFO,ERROR")
	if levels[DebugLevel] || !levels[InfoLevel] || levels[WarnLevel] || !levels[ErrorLevel] ||
		levels[NoticeLevel] || levels[CritLevel] || levels[AlertLevel] || levels[EmergLevel] || levels[FatalLevel] {
		t.Fatalf("expected only INFO and ERROR enabled, got: %+v", levels)
	}
}

func TestParseLevels_CaseInsensitive(t *testing.T) {
	levels := parseLevels("info,DeBuG,WARNING")
	if !levels[DebugLevel] || !levels[InfoLevel] || !levels[WarnLevel] || levels[ErrorLevel] ||
		levels[NoticeLevel] || levels[CritLevel] || levels[AlertLevel] || levels[EmergLevel] || levels[FatalLevel] {
		t.Fatalf("expected case-insensitive parsing, got: %+v", levels)
	}
}

func TestParseLevels_NewLevels(t *testing.T) {
	levels := parseLevels("notice,crit,alert,emerg,fatal")
	if !levels[NoticeLevel] || !levels[CritLevel] || !levels[AlertLevel] || !levels[EmergLevel] || !levels[FatalLevel] {
		t.Fatalf("expected NOTICE/CRIT/ALERT/EMERG/FATAL enabled, got: %+v", levels)
	}
	if levels[DebugLevel] || levels[InfoLevel] || levels[WarnLevel] || levels[ErrorLevel] {
		t.Fatalf("expected only NOTICE/CRIT/ALERT/EMERG/FATAL enabled, got: %+v", levels)
	}
}

func TestEnvironmentLevelFiltering(t *testing.T) {
	// Set environment variable
	os.Setenv("LOGGER_LEVELS", "ERROR")
	defer os.Unsetenv("LOGGER_LEVELS")

	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout, oldStderr := outStdout, outStderr
	defer func() { outStdout, outStderr = oldStdout, oldStderr }()
	outStdout = &stdoutBuf
	outStderr = &stderrBuf

	// Re-initialize logger to pick up env var
	Init(Config{})

	Debugf("debug msg")
	Infof("info msg")
	Errorf("error msg")

	out := stdoutBuf.String() + stderrBuf.String()
	// Should only contain error message
	if strings.Contains(out, "debug msg") || strings.Contains(out, "info msg") {
		t.Fatalf("only error level should be enabled via env var, got: %q", out)
	}
	if !strings.Contains(out, "error msg") {
		t.Fatalf("error message should appear, got: %q", out)
	}
}

func TestCallerInfo_IncludesLineNumber(t *testing.T) {
	var buf bytes.Buffer
	Info = log.New(&buf, "", 0)
	enabledLevels[InfoLevel] = true
	prevInclude := includeCallerTag
	includeCallerTag = true
	defer func() { includeCallerTag = prevInclude }()

	Infof("test message")

	out := buf.String()
	// Should contain line number (format: package.Function:line)
	if !strings.Contains(out, ":") {
		t.Fatalf("expected line number in caller info, got: %q", out)
	}
}
