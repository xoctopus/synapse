package synapse

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/misc/must"
)

func New(ctx context.Context, appliers ...OptionApplier) *Synapse {
	x := &Synapse{inherited: ctx}
	for _, applier := range appliers {
		applier(&x.option)
	}

	if x.cancelCause == nil {
		x.cancelCause = context.Canceled
	}

	switch x.cancelMode {
	case withCancelTimeout:
		must.BeTrueF(x.cancelTimeout > 0, "invalid timeout")
		x.dispatched, x.cancel = context.WithTimeoutCause(ctx, x.cancelTimeout, x.cancelCause)
	case withCancelDeadline:
		x.dispatched, x.cancel = context.WithDeadlineCause(ctx, x.cancelDeadline, x.cancelCause)
	default:
		x.dispatched, x.cancel = context.WithCancelCause(ctx)
	}
	return x
}

type Synapse struct {
	option

	inherited  context.Context
	cancel     any
	dispatched context.Context

	mu       sync.Mutex
	wg       sync.WaitGroup
	canceled atomic.Bool
}

func (x *Synapse) Parent() context.Context {
	return x.inherited
}

func (x *Synapse) Value(key any) any {
	return x.inherited.Value(key)
}

func (x *Synapse) Spawn(f func(ctx context.Context)) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	if !x.canceled.Load() {
		x.wg.Go(func() {
			f(x.dispatched)
		})
		return nil
	}
	return codex.New(ERROR__SYNAPSE_CLOSED)
}

func (x *Synapse) Close(cause error) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.canceled.CompareAndSwap(false, true) {
		switch cancel := x.cancel.(type) {
		case context.CancelFunc:
			cancel()
		case context.CancelCauseFunc:
			cancel(cause)
		}

		if x.shutdownTimeout > 0 {
			sig := make(chan struct{})

			go func() {
				x.wg.Wait()
				close(sig)
			}()

			select {
			case <-sig:
				return nil
			case <-time.After(x.shutdownTimeout):
				return codex.New(ERROR__SYNAPSE_CLOSE_TIMEOUT)
			}
		}
		x.wg.Wait()
	}
	return nil
}
