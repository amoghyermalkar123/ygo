package blockstore

import (
	"os"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLogger configures a zap logger writing to a specific file with function name display
func setupLogger(logPath string) (*zap.Logger, error) {
	// Create the log file if it doesn't exist, append to it if it does
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create encoder configuration with function display
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function", // This enables function name display
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create the core for writing to the file
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	// Create the logger with function caller
	logger := zap.New(
		core,
		zap.AddCaller(),      // Adds the calling function info
		zap.AddCallerSkip(0), // Adjust this if you wrap logger calls
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}

// addFunctionName is a helper to manually add function name to the log context
func addFunctionName() zap.Field {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return zap.String("function", "unknown")
	}
	fn := runtime.FuncForPC(pc)
	return zap.String("function", fn.Name())
}
