package main

import (
	"os"
	"time"

	"github.com/mordilloSan/go-logger/logger"
)

// Example demonstrating the simplified go-logger usage.
func main() {
	logFile := ""

	if len(os.Args) > 1 {
		logFile = os.Args[1]
	}

	// Initialize logger with optional file logging
	// Usage: ./go-logger [logfile]
	// Example: ./go-logger ./app.log
	logger.Init(logger.Config{
		Levels:   logger.AllLevels(), // Adjust this slice to enable/disable levels.
		FilePath: logFile,
	})
	if logFile != "" {
		defer logger.Close() // Don't forget to close the log file!
		logger.Infof("Logging to file: %s", logFile)
	} else {
		logger.Infof("Logging to console only (provide a log file path to enable file logging)")
	}

	// Formatted logging (classic)
	logger.Debugf("starting at %v", time.Now())
	logger.Infof("hello %s", "world")
	logger.Warnln("be careful")
	logger.Errorf("oops: %v", "something happened")

	// Structured logging with key-value pairs
	logger.InfoKV("request completed",
		"duration_ms", 42,
		"status", 200,
		"path", "/api/users",
		"method", "GET")

	logger.ErrorKV("database connection failed",
		"host", "localhost",
		"port", 5432,
		"retry_count", 3,
		"error", "connection timeout")

	logger.DebugKV("cache lookup",
		"key", "user:123",
		"hit", true,
		"ttl_seconds", 300)

	// API logging (automatic level selection based on HTTP status code)
	logger.Api(200, "request successful")
	logger.Api(301, "redirect to new location")
	logger.Api(404, "resource not found")
	logger.Api(500, "internal server error")

	// Uncomment to test Fatal methods (will exit the program):
	// logger.Fatalf("critical error: %v", "system failure")
	// logger.FatalKV("critical error", "component", "database", "action", "shutdown")
}
