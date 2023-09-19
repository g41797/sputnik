package sputnik

import "log"

const (
	labelLen   = 3
	infoLabel  = "[INF] "
	debugLabel = "[DBG] "
	warnLabel  = "[WRN] "
	errorLabel = "[ERR] "
	fatalLabel = "[FTL] "
	traceLabel = "[TRC] "
)

type LoggerFactory func() *Logger

type Logger struct {
	logger *log.Logger
}

func NewLogger(lg *log.Logger) *Logger {
	return &Logger{logger: lg}
}

// Noticef logs a notice statement
func (l *Logger) Noticef(format string, v ...interface{}) {
	l.logger.Printf(infoLabel+format, v...)
}

// Warnf logs a warning statement
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Printf(warnLabel+format, v...)
}

// Errorf logs an error statement
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Printf(errorLabel+format, v...)
}

// Fatalf logs a fatal error
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(fatalLabel+format, v...)
}

// Debugf logs a debug statement
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Printf(debugLabel+format, v...)

}

// Tracef logs a trace statement
func (l *Logger) Tracef(format string, v ...interface{}) {
	l.logger.Printf(traceLabel+format, v...)
}
