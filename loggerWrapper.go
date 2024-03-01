package delaywheel

type LogLevel int

const (
	LOG_DEBUG LogLevel = iota
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

type loggerWrapper struct {
	Logger Logger
	Level  LogLevel
}

func (lo *loggerWrapper) ShouldLog(logLevel LogLevel) bool {
	return logLevel >= lo.Level
}

func (lo *loggerWrapper) Debug(format string, args ...any) {
	if lo.ShouldLog(LOG_DEBUG) {
		lo.Logger.Debug(format, args...)
	}
}

func (lo *loggerWrapper) Info(format string, args ...any) {
	if lo.ShouldLog(LOG_INFO) {
		lo.Logger.Info(format, args...)
	}
}

func (lo *loggerWrapper) Warn(format string, args ...any) {
	if lo.ShouldLog(LOG_WARN) {
		lo.Logger.Warn(format, args...)
	}
}

func (lo *loggerWrapper) Error(format string, args ...any) {
	if lo.ShouldLog(LOG_ERROR) {
		lo.Logger.Error(format, args...)
	}
}
