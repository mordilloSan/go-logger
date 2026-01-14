// Package logger provides a simple leveled logger with
// optional caller function tagging and optional file output.
//
// # Console Output
//
// Plain output is used by default. Set Config.Colorize to enable ANSI colors.
//
// # Features
//
//   - Global package-level functions (no dependency injection needed)
//   - Optional caller tagging [package.Function:line]
//   - Structured logging with key-value pairs
//   - Level filtering via Config.Levels or LOGGER_LEVELS environment variable
//   - Extended syslog-compatible levels: NOTICE, CRIT, ALERT, EMERG
//   - Optional file logging with color stripping for files
//   - Journald priority prefixes for plain output when JOURNAL_STREAM is set
//   - Optional [LEVEL] prefix via Config.IncludeLevelPrefix
//
// # Usage
//
// Initialize once at startup:
//
//	logger.Init(logger.Config{Levels: logger.AllLevels()})
//	logger.Init(logger.Config{Levels: []logger.Level{logger.InfoLevel, logger.WarnLevel}})
//
// Use formatted logging:
//
//	logger.Infof("server started on port %d", 8080)
//	logger.Errorf("failed to connect: %v", err)
//
// Use structured logging with key-value pairs:
//
//	logger.InfoKV("request completed",
//	    "duration_ms", 42,
//	    "status", 200,
//	    "path", "/api/users")
//
// Fatal logging (logs and exits):
//
//	logger.Fatalf("critical error: %v", err)
//	logger.FatalKV("shutdown required", "reason", "out of memory")
//
// # Level Filtering
//
// Configure levels in code via Config.Levels, or leave it nil to honor the environment variable:
//
//	LOGGER_LEVELS="INFO,ERROR" ./myapp
//
// This package is lightweight and has no external dependencies.
package logger
