<div align="center">

[![Release](https://img.shields.io/github/v/release/mordilloSan/go-logger)](https://github.com/mordilloSan/go-logger/releases/latest)
![CI](https://github.com/mordilloSan/go-logger/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mordilloSan/go-logger)](https://goreportcard.com/report/github.com/mordilloSan/go-logger)
[![Go Reference](https://pkg.go.dev/badge/github.com/mordilloSan/go-logger/logger.svg)](https://pkg.go.dev/github.com/mordilloSan/go-logger/logger)

# go-logger

Simple Go logger with optional caller tagging (default off) and optional file output.
</div>

## Features

- **Configurable levels** - Enable/disable individual levels via `Config.Levels` or `LOGGER_LEVELS`
- **Optional colorized output** - ANSI colors per level when `Colorize` is enabled
- **Optional level prefix** - Include `[LEVEL]` when `IncludeLevelPrefix` is enabled (default off)
- **Plain stdout/stderr routing** - INFO/NOTICE/DEBUG to stdout; WARNING/ERROR/CRIT/ALERT/EMERG/FATAL to stderr
- **File logging** - Log to both console and file simultaneously
- **Optional caller tagging** - `[package.Function:line]` when `IncludeCallerTag` is enabled (default off)
- **Structured logging** - Key-value pairs for better debugging
- **API logging** - HTTP status code logging with automatic level mapping

> Note: This package uses only the Go standard library.

## Install

```bash
go get github.com/mordilloSan/go-logger@v1.0.1
```

## Quick Start

### Basic Usage

```go
package main

import (
    "time"
    logx "github.com/mordilloSan/go-logger/logger"
)

func main() {
    // Enable the levels you want to display
    logx.Init(logx.Config{
        Levels: logx.AllLevels(),
    })

    logx.Debugf("starting at %v", time.Now())
    logx.Infof("hello %s", "world")
    logx.Warnln("be careful")
    logx.Errorf("oops: %v", "something happened")

    // Structured logging with key-value pairs
    logx.InfoKV("request completed",
        "duration_ms", 42,
        "status", 200,
        "path", "/api/users")
}
```

### Colorized Console (Optional)

```go
logx.Init(logx.Config{
    Levels:   logx.AllLevels(),
    Colorize: true,
})
```

### File Logging

```go
// Log to both console and file simultaneously
// Console output can be colorized, file output is plain text
logx.Init(logx.Config{
    Levels:   logx.AllLevels(),
    Colorize: true,
    FilePath: "/var/log/myapp.log",
    IncludeLevelPrefix: true,
    IncludeCallerTag:   true,
})
defer logx.Close() // Don't forget to close the log file!

logx.Infof("application started")
// Console: [INFO] 2025/10/26 10:30:45 [main.main:15] application started (colored)
// File:    [INFO] 2025/10/26 10:30:45 [main.main:15] application started (plain text)
```

Behavior summary:

- **Console output:** Plain output to stdout/stderr with no timestamps when not logging to a file (INFO/NOTICE/DEBUG to stdout; WARNING/ERROR/CRIT/ALERT/EMERG/FATAL to stderr)
- **Colorized output:** Set `Colorize` to add ANSI colors (console only)
- **Level prefix:** Default off; set `IncludeLevelPrefix` to add `[LEVEL]`
- **Caller tagging:** Default off; set `IncludeCallerTag` to add `[package.Function:line]`
- **Systemd/journald:** When `JOURNAL_STREAM` is set and output is plain, log lines include syslog priority prefixes (e.g., `<7>` for DEBUG, `<6>` for INFO)
- **File logging:** Logs written to both console and file; ANSI color codes are stripped from file output

## API

### Initialization

- `Init(config Config)` - Setup logger with level selection, optional color, and optional file output
- `InitWithFile(config Config, filePath string)` - Setup logger with a file path override
- `Close() error` - Close the log file (call with `defer` after `Init` when FilePath is set)
- `AllLevels() []Level` - Convenience helper for enabling every level

Config fields:
- `Levels []Level` - Enable specific levels; nil uses `LOGGER_LEVELS` or defaults to all
- `Colorize bool` - Enable ANSI color output for console logs
- `FilePath string` - Log to file when set (logs also go to console)
- `IncludeLevelPrefix bool` - Add the `[LEVEL]` prefix in output
- `IncludeCallerTag bool` - Add the `[package.Function:line]` tag in messages

Defaults: `IncludeLevelPrefix=false`, `IncludeCallerTag=false`.

### Formatted Logging (with fmt.Sprintf)

- `Debugf(format string, v ...interface{})`
- `Infof(format string, v ...interface{})`
- `Noticef(format string, v ...interface{})`
- `Warnf(format string, v ...interface{})`
- `Errorf(format string, v ...interface{})`
- `Critf(format string, v ...interface{})`
- `Alertf(format string, v ...interface{})`
- `Emergf(format string, v ...interface{})`
- `Fatalf(format string, v ...interface{})` - Logs and calls `os.Exit(1)`

### Plain Logging (Println-style)

- `Debugln(v ...interface{})`
- `Infoln(v ...interface{})`
- `Noticeln(v ...interface{})`
- `Warnln(v ...interface{})`
- `Errorln(v ...interface{})`
- `Critln(v ...interface{})`
- `Alertln(v ...interface{})`
- `Emergln(v ...interface{})`
- `Fatalln(v ...interface{})` - Logs and calls `os.Exit(1)`

### Structured Logging (Key-Value Pairs)

- `DebugKV(msg string, keyvals ...any)`
- `InfoKV(msg string, keyvals ...any)`
- `NoticeKV(msg string, keyvals ...any)`
- `WarnKV(msg string, keyvals ...any)`
- `ErrorKV(msg string, keyvals ...any)`
- `CritKV(msg string, keyvals ...any)`
- `AlertKV(msg string, keyvals ...any)`
- `EmergKV(msg string, keyvals ...any)`
- `FatalKV(msg string, keyvals ...any)` - Logs and calls `os.Exit(1)`

Example:
```go
logx.InfoKV("user logged in",
    "user_id", 123,
    "ip", "192.168.1.1",
    "device", "mobile")
```

### API Logging (HTTP Status Code Based)

- `Api(statusCode int, msg string)` - Automatic level selection

Automatically selects log level based on HTTP status code:
- **1xx, 2xx, 3xx** → INFO (green when colorized) - Success and redirects
- **4xx** → WARNING (yellow when colorized) - Client errors
- **5xx** → ERROR (red when colorized) - Server errors

Example:
```go
logx.Api(200, "request successful")
logx.Api(404, "resource not found")
logx.Api(500, "internal server error")
```

## Level Filtering

Enable specific levels in code via `Config.Levels`, or leave it nil to honor the `LOGGER_LEVELS` environment variable:

```go
logx.Init(logx.Config{
    Levels: []logx.Level{logx.InfoLevel, logx.WarnLevel, logx.ErrorLevel},
})
```

Environment variable usage:

```bash
# Only log INFO and ERROR
LOGGER_LEVELS="INFO,ERROR" ./myapp

# Only log ERRORS
LOGGER_LEVELS="ERROR" ./myapp

# Log everything (default if not set)
./myapp
```

Valid level names: `DEBUG`, `INFO`, `NOTICE`, `WARNING`, `ERROR`, `CRIT`, `CRITICAL`, `ALERT`, `EMERG`, `EMERGENCY`, `FATAL`

## Output Examples

### Plain Console Output (with `IncludeLevelPrefix` and `IncludeCallerTag` enabled)

```
[INFO] [main.main:15] server starting on port 8080
[DEBUG] [main.initDB:23] connecting to database host=localhost port=5432
[INFO] [main.handleRequest:42] request completed duration_ms=42 status=200 path=/api/users
[ERROR] [main.processJob:67] job failed job_id=123 error="timeout exceeded"
```

## Use Cases

Perfect for:
- System utilities and daemons
- Web servers and APIs
- CLI applications
- System management dashboards
- Bridge processes requiring elevated privileges

Not ideal for:
- Cloud-native applications (use structured JSON loggers)
- Microservices sending logs to centralized systems

## Compatibility

- **Go:** 1.22+
- **OS:** Works anywhere stdout/stderr are available (ANSI colors shown when Colorize is enabled and terminal supports them)

## Testing

Run all tests:

```bash
make test              # Run all tests
go test ./...          # Or use go directly
go test -v ./...       # Verbose output
make test-concurrency  # Demo concurrency with live progress
```

### Test Coverage

**Concurrency Tests** - Prove thread-safety under extreme load:
- 10,000 goroutines × 100 messages × 4 levels = **4 million log operations**
- 100+ concurrent goroutines using all logging methods
- Real-time progress demo showing mutex effectiveness
- All tests verify **zero garbled output**

**Fatal Method Tests** - Verify logging before process exit:
- Confirms `Fatalf`, `Fatalln`, `FatalKV` write logs before `os.Exit(1)`
- Tests level filtering and output formatting
- Uses subprocess execution for proper testing

**Crash Scenario Tests** - Prove log flushing under failure:
- 5,000 rapid log operations all flushed correctly
- Panic recovery with proper log flushing
- Validates v1.1.0 claims about crash resilience

**Core Functionality Tests**:
- Stdout/stderr routing
- Colorized output (ANSI)
- Config-level filtering
- Environment-based level filtering
- Caller info tagging
- Structured logging (KV pairs)

Tests do not require external services.

### See It In Action

Watch the mutex prevent garbled output from 50 concurrent workers:

```bash
make test-concurrency
```

Output shows clean progress updates:
```
Starting concurrency test: 50 workers × 100 tasks = 5000 total operations
progress completed=1900 total=5000 percent=38.0% active_workers=50 tasks_per_sec=9500
progress completed=3800 total=5000 percent=76.0% active_workers=50 tasks_per_sec=9500
✓ CONCURRENCY TEST COMPLETE!
final stats: 5000 operations in 526ms = 9498 ops/sec - NO GARBLED OUTPUT
```

## Project Layout

```
go-logger/
├── cmd/
│   └── main.go          # Example app
├── logger/
│   ├── logger.go        # Core implementation
│   ├── doc.go          # Package documentation
│   └── *_test.go       # Tests
├── go.mod
└── README.md
```

Run the example app:

```bash
go run ./cmd            # console only
go run ./cmd ./app.log  # console + file
```

## Common Tasks

### Using Makefile (Recommended)

```bash
make                   # Run fmt, vet, and test (default)
make test              # Run all tests with verbose output
make test-concurrency  # Demo real-time concurrent logging (100 goroutines)
make fmt               # Format code
make vet               # Run static analysis
make pre-release       # Run all checks before creating a release
make clean             # Clean build cache
make help              # Show all available targets
```

**See the mutex in action:** Run `make test-concurrency` to watch 100 concurrent goroutines logging in real-time with no garbled lines!

### Using Go Commands Directly

```bash
go fmt ./...      # Format code
go vet ./...      # Lint
go test ./...     # Run tests
go test -v ./...  # Run tests with verbose output
```

## Why This Logger?

- **Simple:** Single `Init(Config)` call
- **Zero dependencies:** Just the Go standard library
- **Optional caller info:** Enable caller tagging when needed
- **Production-ready:** Plain stdout/stderr output plus optional file logging
- **Structured logging:** Key-value pairs for better debugging

## License

MIT. See `LICENSE`.
