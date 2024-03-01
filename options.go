package delaywheel

type Option func(*DelayWheel)

func WithCurTaskID(num uint64) Option {
	return func(dw *DelayWheel) {
		dw.curTaskID.Store(num)
	}
}

func WithAutoRun() Option {
	return func(dw *DelayWheel) {
		dw.autoRun = true
	}
}

func WithLogger(logger Logger) Option {
	return func(dw *DelayWheel) {
		dw.logger.Logger = logger
	}
}

func WithLogLevel(level LogLevel) Option {
	return func(dw *DelayWheel) {
		dw.logger.Level = level
	}
}
