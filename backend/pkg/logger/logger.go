package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger and provides sugared API
type Logger struct {
	core *zap.Logger
	*zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger(level string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: zapLevel == zapcore.DebugLevel,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{core: zapLogger, SugaredLogger: zapLogger.Sugar()}, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	if l == nil || l.core == nil {
		return nil
	}
	return l.core.Sync()
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(keysAndValues ...interface{}) *Logger {
	if l == nil {
		return l
	}
	return &Logger{core: l.core, SugaredLogger: l.SugaredLogger.With(keysAndValues...)}
}
