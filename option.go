package synapse

import (
	"time"
)

type option struct {
	cancelMode      cancelMode
	cancelCause     error
	shutdownTimeout time.Duration
	cancelTimeout   time.Duration
	cancelDeadline  time.Time
}

type OptionApplier func(*option)

func WithCancelCause(cause error) OptionApplier {
	return func(o *option) {
		o.cancelMode = withCancel
		o.cancelCause = cause
	}
}

func WithTimeout(timeout time.Duration) OptionApplier {
	return func(o *option) {
		o.cancelMode = withCancelTimeout
		o.cancelTimeout = timeout
	}
}

func WithCancelDeadline(deadline time.Time) OptionApplier {
	return func(o *option) {
		o.cancelMode = withCancelDeadline
		o.cancelDeadline = deadline
	}
}

func WithShutdownTimeout(timeout time.Duration) OptionApplier {
	return func(o *option) {
		o.shutdownTimeout = timeout
	}
}

type cancelMode int8

const (
	withCancel cancelMode = iota
	withCancelTimeout
	withCancelDeadline
)
