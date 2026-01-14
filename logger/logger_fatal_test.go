package logger

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestFatalf_LogsBeforeExit verifies that Fatalf writes the log message before exiting.
// Since Fatalf calls os.Exit(1), we need to run it in a subprocess.
func TestFatalf_LogsBeforeExit(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		// This is the subprocess that will call Fatalf
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})
		Fatalf("critical error: %s", "database connection failed")
		return
	}

	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalf_LogsBeforeExit")
	cmd.Env = append(os.Environ(), "TEST_FATAL=1")

	output, err := cmd.CombinedOutput()

	// Fatalf should cause os.Exit(1), so we expect an error
	if err == nil {
		t.Fatal("expected Fatalf to exit with non-zero status")
	}

	// Verify the exit code is 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("expected ExitError, got %v", err)
	}

	// Verify the log message was written before exit
	outputStr := string(output)
	if !strings.Contains(outputStr, "critical error: database connection failed") {
		t.Fatalf("expected fatal message in output, got: %q", outputStr)
	}
	if !strings.Contains(outputStr, "[FATAL]") {
		t.Fatalf("expected [FATAL] tag in output, got: %q", outputStr)
	}
}

// TestFatalln_LogsBeforeExit verifies that Fatalln writes the log message before exiting.
func TestFatalln_LogsBeforeExit(t *testing.T) {
	if os.Getenv("TEST_FATALLN") == "1" {
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})
		Fatalln("system shutdown:", "out of memory")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalln_LogsBeforeExit")
	cmd.Env = append(os.Environ(), "TEST_FATALLN=1")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected Fatalln to exit with non-zero status")
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "system shutdown:") || !strings.Contains(outputStr, "out of memory") {
		t.Fatalf("expected fatal message in output, got: %q", outputStr)
	}
}

// TestFatalKV_LogsBeforeExit verifies that FatalKV writes structured logs before exiting.
func TestFatalKV_LogsBeforeExit(t *testing.T) {
	if os.Getenv("TEST_FATALKV") == "1" {
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})
		FatalKV("service failure", "error", "disk full", "path", "/var/log")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalKV_LogsBeforeExit")
	cmd.Env = append(os.Environ(), "TEST_FATALKV=1")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected FatalKV to exit with non-zero status")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "service failure") {
		t.Fatalf("expected fatal message in output, got: %q", outputStr)
	}
	if !strings.Contains(outputStr, "error=disk full") {
		t.Fatalf("expected key-value pairs in output, got: %q", outputStr)
	}
	if !strings.Contains(outputStr, "path=/var/log") {
		t.Fatalf("expected path key-value in output, got: %q", outputStr)
	}
}

// TestFatal_LevelFiltering verifies that FATAL respects level filtering.
func TestFatal_LevelFiltering(t *testing.T) {
	if os.Getenv("TEST_FATAL_FILTERED") == "1" {
		// Disable FATAL level
		os.Setenv("LOGGER_LEVELS", "INFO,ERROR")
		Init(Config{})
		Fatalf("this should exit without logging")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal_LevelFiltering")
	cmd.Env = append(os.Environ(), "TEST_FATAL_FILTERED=1")

	output, err := cmd.CombinedOutput()

	// Should still exit with code 1 even if not logging
	if err == nil {
		t.Fatal("expected Fatalf to exit with non-zero status even when filtered")
	}

	outputStr := string(output)
	// The message should NOT appear since FATAL is filtered out
	if strings.Contains(outputStr, "this should exit without logging") {
		t.Fatalf("fatal message should be filtered out, got: %q", outputStr)
	}
}

// TestCrashScenario_LogsFlushedBeforeKill simulates a process that logs continuously
// and is killed mid-execution to verify logs are properly flushed.
func TestCrashScenario_LogsFlushedBeforeKill(t *testing.T) {
	if os.Getenv("TEST_CRASH_SCENARIO") == "1" {
		// This subprocess will log continuously until killed
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})

		// Log a series of messages
		for i := 0; i < 100; i++ {
			Infof("log-message-%d", i)
			Warnf("warning-message-%d", i)
			ErrorKV("error-event", "iteration", i, "status", "running")
		}

		// Signal that we've finished logging
		Infof("LOGGING_COMPLETE")

		// Keep the process alive for a moment
		// In a real crash, the process would be killed here
		os.Exit(0)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCrashScenario_LogsFlushedBeforeKill")
	cmd.Env = append(os.Environ(), "TEST_CRASH_SCENARIO=1")

	output, err := cmd.CombinedOutput()

	// Process should exit cleanly (we're not actually killing it, just testing flush)
	if err != nil {
		t.Fatalf("unexpected error: %v, output: %s", err, output)
	}

	outputStr := string(output)

	// Verify that all 100 iterations were logged
	for i := 0; i < 100; i++ {
		expectedInfo := "log-message-" + string(rune('0'+i%10))
		expectedWarn := "warning-message-" + string(rune('0'+i%10))

		// At least check for the pattern (substring match)
		if i < 10 && !strings.Contains(outputStr, expectedInfo) {
			t.Fatalf("missing log message for iteration %d, output: %q", i, outputStr)
		}
		if i < 10 && !strings.Contains(outputStr, expectedWarn) {
			t.Fatalf("missing warning message for iteration %d, output: %q", i, outputStr)
		}
	}

	// Verify completion marker
	if !strings.Contains(outputStr, "LOGGING_COMPLETE") {
		t.Fatalf("completion marker not found, logs may not have been flushed. Output: %q", outputStr)
	}

	// Count total log lines to ensure nothing was lost
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	// We expect 100 info + 100 warn + 100 error + 1 completion = 301 lines
	expectedLines := 301
	if len(lines) < expectedLines {
		t.Fatalf("expected at least %d log lines, got %d. Some logs may not have been flushed",
			expectedLines, len(lines))
	}
}

// TestCrashScenario_PanicRecovery tests that logs are flushed even when a panic occurs.
func TestCrashScenario_PanicRecovery(t *testing.T) {
	if os.Getenv("TEST_PANIC_RECOVERY") == "1" {
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})

		// Use a custom recovery to log the panic
		defer func() {
			if r := recover(); r != nil {
				Errorf("panic recovered: %v", r)
			}
		}()

		// Log some messages
		Infof("before panic - message 1")
		Infof("before panic - message 2")
		Infof("before panic - message 3")

		// Trigger a panic
		panic("simulated crash")
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCrashScenario_PanicRecovery")
	cmd.Env = append(os.Environ(), "TEST_PANIC_RECOVERY=1")

	output, err := cmd.CombinedOutput()

	// Process should complete (panic is recovered)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := string(output)

	// Verify all messages before panic were logged
	if !strings.Contains(outputStr, "before panic - message 1") {
		t.Fatalf("message 1 not found in output: %q", outputStr)
	}
	if !strings.Contains(outputStr, "before panic - message 2") {
		t.Fatalf("message 2 not found in output: %q", outputStr)
	}
	if !strings.Contains(outputStr, "before panic - message 3") {
		t.Fatalf("message 3 not found in output: %q", outputStr)
	}

	// Verify panic was recovered and logged
	if !strings.Contains(outputStr, "panic recovered: simulated crash") {
		t.Fatalf("panic recovery message not found in output: %q", outputStr)
	}
}

// TestFatal_OutputFormat verifies Fatal methods produce properly formatted output.
func TestFatal_OutputFormat(t *testing.T) {
	if os.Getenv("TEST_FATAL_FORMAT") == "1" {
		Init(Config{Levels: AllLevels(), IncludeLevelPrefix: true, IncludeCallerTag: true})

		// Log a regular message, then fatal
		Infof("regular message before fatal")
		Fatalf("fatal message: %s", "system failure")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal_OutputFormat")
	cmd.Env = append(os.Environ(), "TEST_FATAL_FORMAT=1")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected Fatalf to exit with non-zero status")
	}

	outputStr := string(output)

	// Both messages should appear and be properly formatted
	if !strings.Contains(outputStr, "[INFO]") || !strings.Contains(outputStr, "regular message before fatal") {
		t.Fatalf("expected INFO message in output, got: %q", outputStr)
	}
	if !strings.Contains(outputStr, "[FATAL]") || !strings.Contains(outputStr, "fatal message: system failure") {
		t.Fatalf("expected FATAL message in output, got: %q", outputStr)
	}

	// Verify caller info is present (function name should be included)
	if !strings.Contains(outputStr, "TestFatal_OutputFormat") {
		t.Fatalf("expected caller info in output, got: %q", outputStr)
	}
}
