package delaywheel

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
	Warn(format string, args ...any)
}

type standardLogger struct {
	l *log.Logger
}

func newStandardLogger() *standardLogger {
	return &standardLogger{
		l: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (st *standardLogger) Debug(format string, args ...any) {
	st.log("GDIT-DEBUG", format, args...)
}

func (st *standardLogger) Info(format string, args ...any) {
	st.log("GDIT-INFO", format, args...)
}

func (st *standardLogger) Warn(format string, args ...any) {
	st.log("GDIT-WARN", format, args...)
}

func (st *standardLogger) Error(format string, args ...any) {
	st.log("GDIT-ERROR", format, args...)
}

func (st *standardLogger) log(level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	st.l.Output(2, fmt.Sprintf("[%s] %s", level, msg))
}
