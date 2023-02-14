package infrastructure

import (
	"fmt"
	"os"

	"github.com/lithammer/shortuuid/v4"
	zero "github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/net/context"
)

var baseLog *Logger

type (
	// Logger - alias for zero.Logger type
	Logger = zero.Logger

	// SugarLogger wrapper for Logger
	SugarLogger struct{}

	// SaramaLogger wrapper for using inside sarama library
	SaramaLogger struct{}
)

// LogPanic - writes message with panic level
func (l *SugarLogger) LogPanic(ctx context.Context, msg string, err error) *Logger {
	log := GetBaseLogger(ctx)
	log.Panic().Stack().Err(err).Msg(msg)
	return log
}

// LogFatal - writes message with fatal level
// Attention! log.Fatal call os.exit(1).
func (l *SugarLogger) LogFatal(ctx context.Context, msg string, err error) *Logger {
	log := GetBaseLogger(ctx)
	log.Fatal().Stack().Err(err).Msg(msg)
	return log
}

// LogError - writes message with error level
func (l *SugarLogger) LogError(ctx context.Context, msg string, err error) *Logger {
	log := GetBaseLogger(ctx)
	log.Error().Err(err).Msg(msg)
	return log
}

// LogWarn - writes message with warn level
func (l *SugarLogger) LogWarn(ctx context.Context, msg string) *Logger {
	log := GetBaseLogger(ctx)
	log.Warn().Msg(msg)
	return log
}

// LogInfo - writes message with info level
func (l *SugarLogger) LogInfo(ctx context.Context, msg string) *Logger {
	log := GetBaseLogger(ctx)
	log.Info().Msg(msg)
	return log
}

// LogDebug - writes message with debug level
func (l *SugarLogger) LogDebug(ctx context.Context, msg string) *Logger {
	log := GetBaseLogger(ctx)
	log.Debug().Msg(msg)
	return log
}

// Log - return reference to base logger
func (l *SugarLogger) Log(ctx context.Context) *Logger {
	return GetBaseLogger(ctx)
}

// InitGlobalLogger - initialize main logger
func InitGlobalLogger(logLevel string, serviceName string, serviceInstance string) {
	zero.ErrorStackMarshaler = pkgerrors.MarshalStack //nolint:reassign
	l := zero.New(os.Stdout).With().
		Caller().
		Timestamp().
		Str("service", serviceName).
		Str("instance", serviceInstance).
		Logger()
	switch logLevel {
	case "warning":
		zero.SetGlobalLevel(zero.WarnLevel)
	case "info":
		zero.SetGlobalLevel(zero.InfoLevel)
	case "debug":
		zero.SetGlobalLevel(zero.DebugLevel)
	case "trace":
		zero.SetGlobalLevel(zero.TraceLevel)
	default:
		zero.SetGlobalLevel(zero.InfoLevel)
	}
	baseLog = &l
}

//------------------------------Sarama logger--------------------------------------------------//

// Print  implementation for SaramaLogger
func (l *SaramaLogger) Print(v ...interface{}) {
	log := GetBaseLogger(context.Background())
	log.Debug().Msg("[SARAMA]" + fmt.Sprintf("%v", v))
}

// Printf implementation for SaramaLogger
func (l *SaramaLogger) Printf(format string, v ...interface{}) {
	log := GetBaseLogger(context.Background())
	log.Debug().Msg("[SARAMA]" + fmt.Sprintf(format, v...))
}

// Println  implementation for SaramaLogger
func (l *SaramaLogger) Println(v ...interface{}) {
	log := GetBaseLogger(context.Background())
	log.Debug().Msg("[SARAMA]" + fmt.Sprintf("%v", v))
}

// GetSaramaLogger return new instance SaramaLogger
func GetSaramaLogger(_ context.Context) *SaramaLogger {
	return &SaramaLogger{}
}

//-----------------------------------------------------------------------------------------------------//

// GenerateID - function for short uid generation
func GenerateID() string {
	return shortuuid.New()
}

// GetBaseLogger - returns logger from context
func GetBaseLogger(ctx context.Context) *Logger {
	l, ok := ctx.Value(CtxKeyLogger{}).(*Logger)
	if !ok || l == nil {
		// looking for logger in context. If not found return baseLog
		l = baseLog
	}
	return l
}
