package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestStdoutStderrRouting(t *testing.T) {
	// Capture stdout/stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout, oldStderr := outStdout, outStderr
	defer func() { outStdout, outStderr = oldStdout, oldStderr }()
	outStdout = &stdoutBuf
	outStderr = &stderrBuf

	Init(Config{Levels: []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel}})

	Infof("hello")
	Debugf("dbg")
	Warnf("careful")
	Errorf("boom")

	if got := stdoutBuf.String(); !strings.Contains(got, "hello") || !strings.Contains(got, "dbg") {
		t.Fatalf("stdout missing expected logs, got: %q", got)
	}
	if got := stderrBuf.String(); !strings.Contains(got, "careful") || !strings.Contains(got, "boom") {
		t.Fatalf("stderr missing expected logs, got: %q", got)
	}
}

func TestPlainOutput_NoAnsi(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout, oldStderr := outStdout, outStderr
	defer func() { outStdout, outStderr = oldStdout, oldStderr }()
	outStdout = &stdoutBuf
	outStderr = &stderrBuf

	Init(Config{Levels: []Level{InfoLevel, ErrorLevel}})
	Infof("plain-info")
	Errorf("plain-error")

	if got := stdoutBuf.String(); !strings.Contains(got, "plain-info") {
		t.Fatalf("stdout missing expected logs, got: %q", got)
	}
	if got := stderrBuf.String(); !strings.Contains(got, "plain-error") {
		t.Fatalf("stderr missing expected logs, got: %q", got)
	}
	if strings.Contains(stdoutBuf.String(), "\033[") || strings.Contains(stderrBuf.String(), "\033[") {
		t.Fatalf("output should be plain (no ANSI codes), got stdout=%q stderr=%q", stdoutBuf.String(), stderrBuf.String())
	}
}

func TestIncludeLevelPrefix_DefaultOff(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	t.Setenv("JOURNAL_STREAM", "")

	Init(Config{Levels: []Level{InfoLevel}})
	Infof("no level prefix")

	line := strings.SplitN(buf.String(), "\n", 2)[0]
	if strings.HasPrefix(line, "[INFO]") {
		t.Fatalf("expected no level prefix by default, got: %q", line)
	}
	if !strings.Contains(line, "no level prefix") {
		t.Fatalf("expected message in output, got: %q", line)
	}
}

func TestColorizedOutput_UsesAnsi(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	Init(Config{Levels: []Level{InfoLevel}, Colorize: true, IncludeLevelPrefix: true})
	Infof("color-info")

	if got := buf.String(); !strings.Contains(got, "\033[") {
		t.Fatalf("expected ANSI color codes when Colorize is enabled, got: %q", got)
	}
}

func TestLevelFiltering_DisablesDebug(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()

	outStdout = &buf
	Init(Config{Levels: []Level{InfoLevel}})
	Debugf("debug-disabled")
	Infof("info-enabled")
	if got := buf.String(); strings.Contains(got, "debug-disabled") {
		t.Fatalf("debug should be disabled by config, got: %q", got)
	}
	if got := buf.String(); !strings.Contains(got, "info-enabled") {
		t.Fatalf("info should be enabled by config, got: %q", got)
	}
}

func TestStdout_NoTimestamps(t *testing.T) {
	var stdoutBuf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &stdoutBuf

	t.Setenv("JOURNAL_STREAM", "")

	Init(Config{Levels: []Level{InfoLevel}, IncludeLevelPrefix: true})
	Infoln("no timestamp expected")

	line := strings.SplitN(stdoutBuf.String(), "\n", 2)[0]
	if !strings.HasPrefix(line, "[INFO] ") {
		t.Fatalf("stdout should start with level prefix, got: %q", line)
	}
	if len(line) >= 5 && line[0] >= '0' && line[0] <= '9' && line[4] == '/' {
		t.Fatalf("stdout should omit date/time when not logging to file, got: %q", line)
	}
}

func TestSyslogPrefixWhenJournalStreamSet(t *testing.T) {
	var stdoutBuf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &stdoutBuf

	t.Setenv("JOURNAL_STREAM", "1:2")

	Init(Config{Levels: []Level{DebugLevel}, IncludeLevelPrefix: true})
	Debugf("dbg")

	line := strings.SplitN(stdoutBuf.String(), "\n", 2)[0]
	if !strings.HasPrefix(line, "<7>[DEBUG] ") {
		t.Fatalf("stdout should include syslog prefix when JOURNAL_STREAM is set, got: %q", line)
	}
}

func TestSyslogPrefixForLevels(t *testing.T) {
	cases := map[string]string{
		"DEBUG":   "<7>",
		"INFO":    "<6>",
		"NOTICE":  "<5>",
		"WARNING": "<4>",
		"ERROR":   "<3>",
		"CRIT":    "<2>",
		"ALERT":   "<1>",
		"EMERG":   "<0>",
		"FATAL":   "<2>",
	}

	for level, want := range cases {
		if got := syslogPrefixForLevel(level); got != want {
			t.Fatalf("syslogPrefixForLevel(%q) = %q, want %q", level, got, want)
		}
	}
}
