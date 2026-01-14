package logger_test

import "github.com/mordilloSan/go-logger/logger"

// This example shows colorized output with all levels enabled.
func ExampleInit_colorized() {
	logger.Init(logger.Config{Levels: logger.AllLevels(), Colorize: true})
	logger.Debugf("debug is on")
	logger.Infof("hello %s", "world")
	logger.Warnln("be careful")
	logger.Errorf("oops: %v", "boom")
}

// This example shows a plain console setup with selective levels enabled.
func ExampleInit_plain() {
	logger.Init(logger.Config{Levels: []logger.Level{logger.InfoLevel, logger.WarnLevel, logger.ErrorLevel}})
	logger.Infof("ready")
}

// This example demonstrates structured logging with key-value pairs.
func ExampleInfoKV() {
	logger.Init(logger.Config{Levels: logger.AllLevels()})

	// Structured logging is great for debugging and analysis
	logger.InfoKV("request completed",
		"duration_ms", 42,
		"status", 200,
		"path", "/api/users",
		"method", "GET")

	logger.ErrorKV("database connection failed",
		"host", "localhost",
		"port", 5432,
		"retry_count", 3)
}

// This example shows how to filter log levels via environment variable.
func ExampleInit_levelFiltering() {
	// Set LOGGER_LEVELS="INFO,ERROR" before running to disable DEBUG and WARNING
	logger.Init(logger.Config{})

	logger.Debugf("this won't appear if DEBUG is filtered")
	logger.Infof("this will appear")
	logger.Warnf("this won't appear if WARNING is filtered")
	logger.Errorf("this will appear")
}
