package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

var Log *Logger

// InitializeGlobalLogger creates and configures the global logger instance
func InitializeGlobalLogger(level, env, serviceName string) {
	zerolog.TimeFieldFormat = time.RFC3339

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zeroLogger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", serviceName).
		Str("env", env).
		Logger()

	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "fatal":
		logLevel = zerolog.FatalLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zeroLogger = zeroLogger.Level(logLevel)

	Log = &Logger{
		logger: zeroLogger,
	}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Info().Msg(msg)
}

func (l *Logger) Infof(msg string, fields ...interface{}) {
	l.logger.Info().Msgf(msg, fields...)
}

func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	l.logger.Error().Err(err).Msg(msg)
}

func (l *Logger) Fatal(msg string, err error) {
	l.logger.Fatal().Err(err).Msg(msg)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Warnf(msg string, fields ...interface{}) {
	l.logger.Warn().Msgf(msg, fields...)
}
