package logger

import (
	"os"
	"sync"
	"ygo/internal/block"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	log *zap.Logger
	// Once ensures the logger is initialized only once
	once sync.Once
)

// Init initializes the global logger
func Init() {
	logPath := "blockstore.log" // Default log path

	once.Do(func() {
		// Create encoder configuration
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    "function", // Enable function name
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Create a file core for writing to a specific log file
		var core zapcore.Core
		if logPath != "" {
			// Create log file if not exists
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				fileCore := zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					zapcore.AddSync(logFile),
					zap.NewAtomicLevelAt(zap.DebugLevel),
				)

				// Create a multi-core to write to both console and file
				core = zapcore.NewTee(
					fileCore,
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(encoderConfig),
						zapcore.AddSync(os.Stdout),
						zap.NewAtomicLevelAt(zap.DebugLevel),
					),
				)
			}
		}

		// If file creation failed or no path provided, log to console only
		if core == nil {
			core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				zap.NewAtomicLevelAt(zap.InfoLevel),
			)
		}

		// Create the global logger
		log = zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(1), // Skip the wrapper function
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	})
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	// Initialize with default configuration if not already initialized
	if log == nil {
		Init()
	}
	return log
}

// Debug logs a debug message
func Debug(msg string, block *block.Block, tlp *block.BlockTextListPosition, fields ...zap.Field) {
	if block != nil {
		fields = append(fields, zap.Any("Block", block.ID))
		if block.Left != nil {
			fields = append(fields, zap.Any("Block_left", block.Left.ID))
		}
		if block.Right != nil {
			fields = append(fields, zap.Any("Block_right", block.Right.ID))
		}
		fields = append(fields, zap.String("Block_content", block.Content))
	}

	if tlp != nil {
		if tlp.Left != nil {
			fields = append(fields, zap.Any("TLP_left", tlp.Left.ID))
		}
		if tlp.Right != nil {
			fields = append(fields, zap.Any("TLP_right", tlp.Right.ID))
		}
		fields = append(fields, zap.Int64("TLP_index", tlp.Index))
	}

	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return GetLogger().Sync()
}
