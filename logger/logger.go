package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Levels define log severity.
type Level int

const (
	// DebugLevel enables debug logging.
	DebugLevel Level = iota
	// InfoLevel enables informational logging.
	InfoLevel
	// WarnLevel enables warning logging.
	WarnLevel
	// ErrorLevel enables error logging.
	ErrorLevel
	// FatalLevel enables fatal logging (exits after logging).
	FatalLevel
	// NoticeLevel enables notice logging.
	NoticeLevel
	// CritLevel enables critical logging.
	CritLevel
	// AlertLevel enables alert logging.
	AlertLevel
	// EmergLevel enables emergency logging.
	EmergLevel
)

// Config defines options for Init, including level filtering and output formatting.
// If Levels is nil, Init uses LOGGER_LEVELS when set; otherwise all levels are enabled.
type Config struct {
	// Levels limits which log levels are enabled; nil falls back to LOGGER_LEVELS or all levels.
	// Default: nil (all levels enabled)
	Levels []Level
	// Colorize enables ANSI color output for console logs.
	// Default: false
	Colorize bool
	// FilePath writes logs to this file (created/appended); empty disables file logging.
	// Default: "" (file logging disabled)
	FilePath string
	// IncludeLevelPrefix adds the [LEVEL] tag in console and file output.
	// Default: false
	IncludeLevelPrefix bool
	// IncludeCallerTag adds the [package.Function:line] tag in log messages.
	// Default: false
	IncludeCallerTag bool
}

// AllLevels returns all supported levels.
func AllLevels() []Level {
	return []Level{
		DebugLevel,
		InfoLevel,
		NoticeLevel,
		WarnLevel,
		ErrorLevel,
		CritLevel,
		AlertLevel,
		EmergLevel,
		FatalLevel,
	}
}

// global state
var (
	// log.Logger instances for formatted output
	// Debug is the logger for debug-level messages.
	Debug = log.New(io.Discard, "", 0)
	// Info is the logger for info-level messages.
	Info = log.New(io.Discard, "", 0)
	// Notice is the logger for notice-level messages.
	Notice = log.New(io.Discard, "", 0)
	// Warning is the logger for warning-level messages.
	Warning = log.New(io.Discard, "", 0)
	// Error is the logger for error-level messages.
	Error = log.New(io.Discard, "", 0)
	// Crit is the logger for critical-level messages.
	Crit = log.New(io.Discard, "", 0)
	// Alert is the logger for alert-level messages.
	Alert = log.New(io.Discard, "", 0)
	// Emerg is the logger for emergency-level messages.
	Emerg = log.New(io.Discard, "", 0)
	// Fatal is the logger for fatal-level messages.
	Fatal = log.New(io.Discard, "", 0)

	// Mutex for thread-safe logging across concurrent goroutines
	logMutex sync.Mutex

	// enabled levels (for filtering)
	enabledLevels = allLevelsEnabled()

	// logFile holds the file handle for file logging (if enabled)
	logFile *os.File

	// includeCallerTag controls whether caller info is added to log messages.
	includeCallerTag = false
)

// Dependency injection points for testing outputs.
var (
	outStdout io.Writer = os.Stdout
	outStderr io.Writer = os.Stderr
)

// Init initializes the logger with configurable levels and optional color output.
// If Config.Levels is nil, LOGGER_LEVELS is used when set; otherwise all levels are enabled.
// Call Close() to properly close the log file when shutting down.
func Init(config Config) {
	enabledLevels = resolveLevels(config.Levels)
	showLevel := config.IncludeLevelPrefix
	includeCallerTag = config.IncludeCallerTag

	// Open log file if specified
	var fileWriter io.Writer
	if config.FilePath != "" {
		f, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(outStderr, "failed to open log file %s: %v\n", config.FilePath, err)
		} else {
			logFile = f
			fileWriter = f
		}
	}

	if config.Colorize {
		Debug = newColorLogger(outStdout, "DEBUG", showLevel, fileWriter)
		Info = newColorLogger(outStdout, "INFO", showLevel, fileWriter)
		Notice = newColorLogger(outStdout, "NOTICE", showLevel, fileWriter)
		Warning = newColorLogger(outStderr, "WARNING", showLevel, fileWriter)
		Error = newColorLogger(outStderr, "ERROR", showLevel, fileWriter)
		Crit = newColorLogger(outStderr, "CRIT", showLevel, fileWriter)
		Alert = newColorLogger(outStderr, "ALERT", showLevel, fileWriter)
		Emerg = newColorLogger(outStderr, "EMERG", showLevel, fileWriter)
		Fatal = newColorLogger(outStderr, "FATAL", showLevel, fileWriter)
		return
	}

	Debug = newPlainLogger(outStdout, "DEBUG", showLevel, fileWriter)
	Info = newPlainLogger(outStdout, "INFO", showLevel, fileWriter)
	Notice = newPlainLogger(outStdout, "NOTICE", showLevel, fileWriter)
	Warning = newPlainLogger(outStderr, "WARNING", showLevel, fileWriter)
	Error = newPlainLogger(outStderr, "ERROR", showLevel, fileWriter)
	Crit = newPlainLogger(outStderr, "CRIT", showLevel, fileWriter)
	Alert = newPlainLogger(outStderr, "ALERT", showLevel, fileWriter)
	Emerg = newPlainLogger(outStderr, "EMERG", showLevel, fileWriter)
	Fatal = newPlainLogger(outStderr, "FATAL", showLevel, fileWriter)
}

// InitWithFile initializes the logger with a file path override.
func InitWithFile(config Config, filePath string) {
	config.FilePath = filePath
	Init(config)
}

// Close closes the log file if it was opened.
// Call this function when your application shuts down to ensure logs are flushed.
func Close() error {
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}

func resolveLevels(levels []Level) map[Level]bool {
	if levels != nil {
		return levelsFromSlice(levels)
	}
	if env := os.Getenv("LOGGER_LEVELS"); env != "" {
		return parseLevels(env)
	}
	return allLevelsEnabled()
}

func levelsFromSlice(levels []Level) map[Level]bool {
	m := make(map[Level]bool, len(levels))
	for _, level := range levels {
		m[level] = true
	}
	return m
}

func allLevelsEnabled() map[Level]bool {
	return map[Level]bool{
		DebugLevel:  true,
		InfoLevel:   true,
		NoticeLevel: true,
		WarnLevel:   true,
		ErrorLevel:  true,
		CritLevel:   true,
		AlertLevel:  true,
		EmergLevel:  true,
		FatalLevel:  true,
	}
}

// parseLevels parses a comma-separated list of level names.
// Empty string enables all levels.
func parseLevels(s string) map[Level]bool {
	m := map[Level]bool{}
	s = strings.TrimSpace(s)
	if s == "" {
		m[DebugLevel] = true
		m[InfoLevel] = true
		m[NoticeLevel] = true
		m[WarnLevel] = true
		m[ErrorLevel] = true
		m[CritLevel] = true
		m[AlertLevel] = true
		m[EmergLevel] = true
		m[FatalLevel] = true
		return m
	}
	for _, p := range strings.Split(s, ",") {
		switch strings.ToUpper(strings.TrimSpace(p)) {
		case "DEBUG":
			m[DebugLevel] = true
		case "INFO":
			m[InfoLevel] = true
		case "NOTICE":
			m[NoticeLevel] = true
		case "WARNING":
			m[WarnLevel] = true
		case "ERROR":
			m[ErrorLevel] = true
		case "CRIT", "CRITICAL":
			m[CritLevel] = true
		case "ALERT":
			m[AlertLevel] = true
		case "EMERG", "EMERGENCY":
			m[EmergLevel] = true
		case "FATAL":
			m[FatalLevel] = true
		}
	}
	return m
}

// isLevelEnabled checks if a level is enabled for logging.
func isLevelEnabled(level Level) bool {
	return enabledLevels[level]
}

// newColorLogger returns a colored logger for the level.
// If fileWriter is provided, logs are written to both console and file.
func newColorLogger(out io.Writer, level string, showLevel bool, fileWriter io.Writer) *log.Logger {
	colors := map[string]string{
		"DEBUG":   "\033[36m",
		"INFO":    "\033[32m",
		"NOTICE":  "\033[34m",
		"WARNING": "\033[33m",
		"ERROR":   "\033[31m",
		"CRIT":    "\033[91m",
		"ALERT":   "\033[95m",
		"EMERG":   "\033[97m",
		"FATAL":   "\033[35m",
	}
	reset := "\033[0m"
	prefix := ""
	if showLevel {
		prefix = fmt.Sprintf("%s[%s]%s", colors[level], level, reset)
	}

	// Combine console and file output if file writer is provided
	if fileWriter != nil {
		// Write colored output to console, plain output to file
		return log.New(io.MultiWriter(out, &plainFileWriter{w: fileWriter, level: level}), prefixForLog(prefix), log.LstdFlags)
	}
	return log.New(out, prefixForLog(prefix), log.LstdFlags)
}

// newPlainLogger returns a non-colored logger for stdout/stderr output.
// If fileWriter is provided, logs are written to both console and file.
func newPlainLogger(out io.Writer, level string, showLevel bool, fileWriter io.Writer) *log.Logger {
	prefix := ""
	if showLevel {
		prefix = fmt.Sprintf("[%s]", level)
	}
	outWriter := out
	if shouldUseSyslogPrefix() {
		if syslogPrefix := syslogPrefixForLevel(level); syslogPrefix != "" {
			outWriter = &syslogPrefixWriter{w: out, prefix: syslogPrefix}
		}
	}
	if fileWriter != nil {
		return log.New(io.MultiWriter(outWriter, &timestampWriter{w: fileWriter}), prefixForLog(prefix), 0)
	}
	return log.New(outWriter, prefixForLog(prefix), 0)
}

func prefixForLog(prefix string) string {
	if prefix == "" {
		return ""
	}
	return prefix + " "
}

func shouldUseSyslogPrefix() bool {
	return os.Getenv("JOURNAL_STREAM") != ""
}

func syslogPrefixForLevel(level string) string {
	switch level {
	case "EMERG":
		return "<0>"
	case "ALERT":
		return "<1>"
	case "CRIT":
		return "<2>"
	case "DEBUG":
		return "<7>"
	case "INFO":
		return "<6>"
	case "NOTICE":
		return "<5>"
	case "WARNING":
		return "<4>"
	case "ERROR":
		return "<3>"
	case "FATAL":
		return "<2>"
	default:
		return ""
	}
}

// syslogPrefixWriter prepends the syslog priority prefix to each line.
type syslogPrefixWriter struct {
	w      io.Writer
	prefix string
}

func (s *syslogPrefixWriter) Write(data []byte) (int, error) {
	if s.prefix == "" {
		return s.w.Write(data)
	}
	if len(data) == 0 {
		return 0, nil
	}
	buf := make([]byte, 0, len(data)+len(s.prefix))
	buf = append(buf, s.prefix...)
	for i, b := range data {
		buf = append(buf, b)
		if b == '\n' && i != len(data)-1 {
			buf = append(buf, s.prefix...)
		}
	}
	if _, err := s.w.Write(buf); err != nil {
		return 0, err
	}
	return len(data), nil
}

// plainFileWriter wraps a file writer to strip ANSI color codes before writing.
type plainFileWriter struct {
	w     io.Writer
	level string
}

func (p *plainFileWriter) Write(data []byte) (int, error) {
	// Strip ANSI color codes (basic implementation)
	s := string(data)
	// Remove color codes like \033[36m and \033[0m
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	// The log.Logger already adds the level prefix, so we just need to strip colors
	// Don't add duplicate level prefix here
	return p.w.Write([]byte(result.String()))
}

// timestampWriter prepends a timestamp to each log line for file outputs.
// Used to keep timestamps in files while omitting them from stdout/stderr output.
type timestampWriter struct {
	w io.Writer
}

func (t *timestampWriter) Write(data []byte) (int, error) {
	ts := time.Now().Format("2006/01/02 15:04:05 ")
	buf := make([]byte, 0, len(ts)+len(data))
	buf = append(buf, ts...)
	buf = append(buf, data...)
	return t.w.Write(buf)
}

// getCallerInfo returns formatted caller information at the specified stack depth.
// Returns "package.Function" format for better log clarity.
func getCallerInfo(depth int) string {
	pc, _, line, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	full := fn.Name()
	// Strip package path, keep package.Function
	lastSlash := strings.LastIndex(full, "/")
	if lastSlash >= 0 && lastSlash+1 < len(full) {
		full = full[lastSlash+1:]
	}
	return fmt.Sprintf("%s:%d", full, line)
}

func formatWithCaller(depth int, msg string) string {
	if !includeCallerTag {
		return msg
	}
	caller := getCallerInfo(depth + 1)
	return fmt.Sprintf("[%s] %s", caller, msg)
}

// encodeFields formats key-value pairs as "key=value" strings.
func encodeFields(keyvals ...any) string {
	if len(keyvals) == 0 {
		return ""
	}
	parts := make([]string, 0, len(keyvals)/2)
	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, keyvals[i+1]))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// --- Formatted logging methods (fmt.Sprintf style) ---

// Debugf logs a debug message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Debugf(format string, v ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Debug.Println(msg)
}

// Infof logs an informational message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Infof(format string, v ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Info.Println(msg)
}

// Noticef logs a notice message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Noticef(format string, v ...any) {
	if !isLevelEnabled(NoticeLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Notice.Println(msg)
}

// Warnf logs a warning message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Warnf(format string, v ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Warning.Println(msg)
}

// Errorf logs an error message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Errorf(format string, v ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Error.Println(msg)
}

// Critf logs a critical message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Critf(format string, v ...any) {
	if !isLevelEnabled(CritLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Crit.Println(msg)
}

// Alertf logs an alert message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Alertf(format string, v ...any) {
	if !isLevelEnabled(AlertLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Alert.Println(msg)
}

// Emergf logs an emergency message formatted with fmt.Sprintf.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Emergf(format string, v ...any) {
	if !isLevelEnabled(EmergLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Emerg.Println(msg)
}

// Fatalf logs a fatal message formatted with fmt.Sprintf and then calls os.Exit(1).
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Fatalf(format string, v ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprintf(format, v...)
	msg = formatWithCaller(2, msg)
	Fatal.Println(msg)
	os.Exit(1)
}

// --- Plain logging methods (Println style) ---

// Debugln logs a debug message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Debugln(v ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Debug.Println(msg)
}

// Infoln logs an informational message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Infoln(v ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Info.Println(msg)
}

// Noticeln logs a notice message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Noticeln(v ...any) {
	if !isLevelEnabled(NoticeLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Notice.Println(msg)
}

// Warnln logs a warning message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Warnln(v ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Warning.Println(msg)
}

// Errorln logs an error message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Errorln(v ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Error.Println(msg)
}

// Critln logs a critical message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Critln(v ...any) {
	if !isLevelEnabled(CritLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Crit.Println(msg)
}

// Alertln logs an alert message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Alertln(v ...any) {
	if !isLevelEnabled(AlertLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Alert.Println(msg)
}

// Emergln logs an emergency message by joining arguments with fmt.Sprint.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Emergln(v ...any) {
	if !isLevelEnabled(EmergLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Emerg.Println(msg)
}

// Fatalln logs a fatal message by joining arguments with fmt.Sprint and then calls os.Exit(1).
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func Fatalln(v ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	msg := fmt.Sprint(v...)
	msg = formatWithCaller(2, msg)
	Fatal.Println(msg)
	os.Exit(1)
}

// --- Structured logging methods (key-value pairs) ---

// DebugKV logs a debug message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func DebugKV(msg string, keyvals ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Debug.Println(line)
}

// InfoKV logs an info message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func InfoKV(msg string, keyvals ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Info.Println(line)
}

// NoticeKV logs a notice message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func NoticeKV(msg string, keyvals ...any) {
	if !isLevelEnabled(NoticeLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Notice.Println(line)
}

// WarnKV logs a warning message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func WarnKV(msg string, keyvals ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Warning.Println(line)
}

// ErrorKV logs an error message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func ErrorKV(msg string, keyvals ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Error.Println(line)
}

// CritKV logs a critical message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func CritKV(msg string, keyvals ...any) {
	if !isLevelEnabled(CritLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Crit.Println(line)
}

// AlertKV logs an alert message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func AlertKV(msg string, keyvals ...any) {
	if !isLevelEnabled(AlertLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Alert.Println(line)
}

// EmergKV logs an emergency message with structured key-value pairs.
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func EmergKV(msg string, keyvals ...any) {
	if !isLevelEnabled(EmergLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Emerg.Println(line)
}

// FatalKV logs a fatal message with structured key-value pairs and then calls os.Exit(1).
// Caller tagging is included when enabled in Init.
// Thread-safe for concurrent use.
func FatalKV(msg string, keyvals ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	fields := encodeFields(keyvals...)
	line := fmt.Sprintf("%s%s", msg, fields)
	line = formatWithCaller(2, line)
	Fatal.Println(line)
	os.Exit(1)
}

// --- API logging methods (HTTP status code based) ---

// Api logs an HTTP API call with automatic level selection based on status code.
// Status codes are mapped to levels: 2xx->INFO, 4xx->WARNING, 5xx->ERROR.
// Thread-safe for concurrent use.
//
// Example:
//
//	logger.Api(200, "api call successful")
//	logger.Api(404, "resource not found")
//	logger.Api(500, "internal server error")
func Api(statusCode int, msg string) {
	level := statusCodeToLevel(statusCode)
	if !isLevelEnabled(level) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	logMsg := fmt.Sprintf("[%d] %s", statusCode, msg)
	logMsg = formatWithCaller(2, logMsg)

	switch level {
	case InfoLevel:
		Info.Println(logMsg)
	case WarnLevel:
		Warning.Println(logMsg)
	case ErrorLevel:
		Error.Println(logMsg)
	}
}

// statusCodeToLevel maps HTTP status codes to log levels.
// 1xx, 2xx, 3xx -> INFO, 4xx -> WARNING, 5xx -> ERROR
func statusCodeToLevel(code int) Level {
	switch {
	case code >= 500:
		return ErrorLevel
	case code >= 400:
		return WarnLevel
	case code >= 300:
		return InfoLevel // 3xx redirects are informational, not warnings
	default:
		return InfoLevel // 1xx, 2xx
	}
}
