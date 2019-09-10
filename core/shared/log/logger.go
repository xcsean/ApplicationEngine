package log

import (
	"fmt"

	"github.com/rs/zerolog"
)

var (
	mainflag bool
	mainlog  zerolog.Logger
)

func init() {
	mainflag = false
}

// SetupMainLogger setup the main logger settings
func SetupMainLogger(dir, filename, level string) error {
	if mainflag {
		return fmt.Errorf("main log already init")
	}
	rot := NewRotatorByHour(dir, filename)
	if rot == nil {
		return fmt.Errorf("new rotator by hour failed")
	}
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000 MST"
	mainlog = zerolog.New(rot).With().Timestamp().Logger()
	mainflag = true
	return nil
}

// ChangeLevel change the level of logger
func ChangeLevel(newLevel string) string {
	return setLogLevel(newLevel)
}

// Debug output debug level log
func Debug(format string, v ...interface{}) {
	mainlog.Debug().Msg(fmt.Sprintf(format, v...))
}

// Info output info level log
func Info(format string, v ...interface{}) {
	mainlog.Info().Msg(fmt.Sprintf(format, v...))
}

// Warn output warning level log
func Warn(format string, v ...interface{}) {
	mainlog.Warn().Msg(fmt.Sprintf(format, v...))
}

// Error output error level log
func Error(format string, v ...interface{}) {
	mainlog.Error().Msg(fmt.Sprintf(format, v...))
}

// Fatal output fatal level log
func Fatal(format string, v ...interface{}) {
	mainlog.Fatal().Msg(fmt.Sprintf(format, v...))
}

// Panic output panic level log
func Panic(format string, v ...interface{}) {
	mainlog.Panic().Msg(fmt.Sprintf(format, v...))
}

// internal functions
func getLogLevel(logLevel string) string {
	switch logLevel {
	case "debug", "info", "warn", "error", "fatal", "panic":
		return logLevel
	default:
		return "info"
	}
}

func setLogLevel(logLevel string) string {
	level := getLogLevel(logLevel)
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}
	return level
}
