package testcontainer

import (
	"context"
	"fmt"

	"go-service-template/internal/app/infrastructure"
)

// Logger - logger for using inside testcontainer engine
type Logger struct{}

// Printf write log message
func (l *Logger) Printf(format string, v ...interface{}) {
	log := infrastructure.GetBaseLogger(context.Background())
	log.Debug().Msg("[TestContainer]" + fmt.Sprintf(format, v...))
}
