package delaywheel

type Option func(*DelayWheel)

// WithCurTaskID sets the initial task ID for the DelayWheel.
// This can be used to start the task IDs from a specific number.
func WithCurTaskID(num uint64) Option {
	return func(dw *DelayWheel) {
		dw.curTaskID.Store(num)
	}
}

// WithAutoRun enables the automatic execution of tasks when they are due.
// When enabled, tasks will be executed in their own goroutines as soon as they are scheduled.
func WithAutoRun() Option {
	return func(dw *DelayWheel) {
		dw.autoRun = true
	}
}

// WithLogger sets the logger to be used by the DelayWheel.
// This allows for custom logging implementations to be integrated into the DelayWheel,
// facilitating logging of internal events according to the user's preferences.
func WithLogger(logger Logger) Option {
	return func(dw *DelayWheel) {
		dw.logger.Logger = logger
	}
}

// WithLogLevel sets the logging level for the DelayWheel.
// This controls the verbosity of the logs produced by the DelayWheel,
// allowing for finer control over what is logged.
func WithLogLevel(level LogLevel) Option {
	return func(dw *DelayWheel) {
		dw.logger.Level = level
	}
}

// WithPendingBufferSize sets the size of the buffer for the channel that holds pending tasks.
// This can be used to control the maximum number of tasks that can be held in the queue before being processed,
// which can help in managing memory usage and controlling how tasks are batched for execution.
func WithPendingBufferSize(size int) Option {
	return func(dw *DelayWheel) {
		dw.pendingTaskCh = make(chan func(), size)
	}
}
