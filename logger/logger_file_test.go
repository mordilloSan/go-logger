package logger

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func discardOutput() func() {
	oldStdout, oldStderr := outStdout, outStderr
	outStdout = io.Discard
	outStderr = io.Discard
	return func() {
		outStdout = oldStdout
		outStderr = oldStderr
	}
}

func TestFileLogging_ColorizedStripsAnsi(t *testing.T) {
	defer discardOutput()()
	// Create a temporary log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Initialize logger with file logging
	Init(Config{Levels: AllLevels(), Colorize: true, FilePath: logPath, IncludeLevelPrefix: true})
	defer Close()

	// Log some messages
	Infof("test info message")
	Warnln("test warning")
	ErrorKV("test error", "key", "value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Verify all messages are in the file
	if !strings.Contains(log, "test info message") {
		t.Errorf("log file should contain info message, got: %q", log)
	}
	if !strings.Contains(log, "test warning") {
		t.Errorf("log file should contain warning, got: %q", log)
	}
	if !strings.Contains(log, "test error") {
		t.Errorf("log file should contain error message, got: %q", log)
	}
	if !strings.Contains(log, "key=value") {
		t.Errorf("log file should contain structured fields, got: %q", log)
	}

	// Verify level prefixes are present
	if !strings.Contains(log, "[INFO]") {
		t.Errorf("log file should contain [INFO] prefix, got: %q", log)
	}
	if !strings.Contains(log, "[WARNING]") {
		t.Errorf("log file should contain [WARNING] prefix, got: %q", log)
	}
	if !strings.Contains(log, "[ERROR]") {
		t.Errorf("log file should contain [ERROR] prefix, got: %q", log)
	}

	// Verify no ANSI color codes in file
	if strings.Contains(log, "\033[") {
		t.Errorf("log file should not contain ANSI color codes, got: %q", log)
	}
}

func TestFileLogging_Plain(t *testing.T) {
	defer discardOutput()()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "prod.log")

	Init(Config{Levels: []Level{InfoLevel, ErrorLevel}, FilePath: logPath})
	defer Close()

	Infof("plain info")
	Errorf("plain error")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	if !strings.Contains(log, "plain info") {
		t.Errorf("log should contain plain info, got: %q", log)
	}
	if !strings.Contains(log, "plain error") {
		t.Errorf("log should contain plain error, got: %q", log)
	}
}

func TestFileLogging_Timestamps(t *testing.T) {
	defer discardOutput()()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "prod-ts.log")

	Init(Config{Levels: []Level{InfoLevel}, FilePath: logPath, IncludeLevelPrefix: true})
	defer Close()

	Infof("plain info")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected at least one log line in file")
	}

	first := lines[0]
	tsPattern := regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `)
	if !tsPattern.MatchString(first) {
		t.Fatalf("file logs should include date/time, got: %q", first)
	}
	if !strings.Contains(first, "[INFO]") {
		t.Fatalf("file logs should include level prefix, got: %q", first)
	}
}

func TestFileLogging_Append(t *testing.T) {
	defer discardOutput()()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append.log")

	// First initialization
	Init(Config{Levels: []Level{InfoLevel}, FilePath: logPath})
	Infof("first message")
	Close()

	// Second initialization (should append, not overwrite)
	Init(Config{Levels: []Level{InfoLevel}, FilePath: logPath})
	Infof("second message")
	Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Both messages should be present
	if !strings.Contains(log, "first message") {
		t.Errorf("log should contain first message, got: %q", log)
	}
	if !strings.Contains(log, "second message") {
		t.Errorf("log should contain second message, got: %q", log)
	}
}

func TestFileLogging_InvalidPath(t *testing.T) {
	defer discardOutput()()
	// Try to log to an invalid path (should not crash, just continue without file logging)
	invalidPath := "/nonexistent/directory/test.log"

	// This should not crash - it will print an error to stderr but continue
	Init(Config{Levels: []Level{InfoLevel}, FilePath: invalidPath})
	defer Close()

	// Logger should still work (just won't write to file)
	Infof("test message")

	// The logFile should be nil since the file couldn't be opened
	if logFile != nil {
		t.Errorf("logFile should be nil when path is invalid, got: %v", logFile)
	}
}

func TestFileLogging_Close(t *testing.T) {
	defer discardOutput()()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "close.log")

	Init(Config{Levels: []Level{InfoLevel}, FilePath: logPath})
	Infof("before close")

	// Close should succeed
	err := Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}

	// Second close should be safe (no-op)
	err = Close()
	if err != nil {
		t.Errorf("second Close() should not return error, got: %v", err)
	}

	// File should contain the message
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "before close") {
		t.Errorf("log should contain message, got: %q", string(content))
	}
}

func TestFileLogging_NoFile(t *testing.T) {
	defer discardOutput()()
	// Init without file (empty path) should work normally
	Init(Config{Levels: []Level{InfoLevel}})
	defer Close()

	// Should not crash
	Infof("test message")

	// Close should be safe even with no file
	err := Close()
	if err != nil {
		t.Errorf("Close() with no file should not error, got: %v", err)
	}
}

func TestFileLogging_AllLevels(t *testing.T) {
	defer discardOutput()()
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "levels.log")

	Init(Config{Levels: AllLevels(), FilePath: logPath, IncludeLevelPrefix: true})
	defer Close()

	// Test all logging methods
	Debugf("debug %s", "formatted")
	Debugln("debug plain")
	DebugKV("debug structured", "key", "value")

	Infof("info %s", "formatted")
	Infoln("info plain")
	InfoKV("info structured", "key", "value")

	Noticef("notice %s", "formatted")
	Noticeln("notice plain")
	NoticeKV("notice structured", "key", "value")

	Warnf("warn %s", "formatted")
	Warnln("warn plain")
	WarnKV("warn structured", "key", "value")

	Errorf("error %s", "formatted")
	Errorln("error plain")
	ErrorKV("error structured", "key", "value")

	Critf("crit %s", "formatted")
	Critln("crit plain")
	CritKV("crit structured", "key", "value")

	Alertf("alert %s", "formatted")
	Alertln("alert plain")
	AlertKV("alert structured", "key", "value")

	Emergf("emerg %s", "formatted")
	Emergln("emerg plain")
	EmergKV("emerg structured", "key", "value")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Verify all levels are logged
	expectedStrings := []string{
		"debug formatted", "debug plain", "debug structured",
		"info formatted", "info plain", "info structured",
		"notice formatted", "notice plain", "notice structured",
		"warn formatted", "warn plain", "warn structured",
		"error formatted", "error plain", "error structured",
		"crit formatted", "crit plain", "crit structured",
		"alert formatted", "alert plain", "alert structured",
		"emerg formatted", "emerg plain", "emerg structured",
		"[DEBUG]", "[INFO]", "[NOTICE]", "[WARNING]", "[ERROR]", "[CRIT]", "[ALERT]", "[EMERG]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(log, expected) {
			t.Errorf("log should contain %q, got: %q", expected, log)
		}
	}
}

func TestInit_DefaultConfig(t *testing.T) {
	defer discardOutput()()
	// Test that Init works with default config (no file logging)
	Init(Config{})
	defer Close()

	// Should not crash
	Infof("test message")

	// Close should be safe
	err := Close()
	if err != nil {
		t.Errorf("Close() should not error, got: %v", err)
	}
}
