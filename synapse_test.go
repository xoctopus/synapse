package synapse_test

import (
	"context"
	"fmt"
	"time"

	syn "github.com/xoctopus/synapse"
)

func ExampleSynapse() {
	// Close
	run := func(worker func(context.Context, time.Duration, int), workerCost, shutdownTimeout, closeWait time.Duration) {
		// 1. Initialize Synapse with a 2-second shutdown timeout
		ctx := context.Background()
		s := syn.New(ctx, syn.WithShutdownTimeout(shutdownTimeout))

		// 2. Spawn workers
		for i := 1; i <= 3; i++ {
			workerID := i
			err := s.Spawn(func(ctx context.Context) {
				worker(ctx, workerCost, workerID)
			})
			if err != nil {
				return
			}
		}
		time.Sleep(closeWait)
		// 3. Close the Synapse to trigger graceful shutdown
		// This blocks until all workers finish or the 2s timeout is reached.
		if err := s.Close(nil); err != nil {
			fmt.Printf("Close failed: %v\n", err)
		}
	}

	// worker1 do task controlled by context
	worker1 := func(ctx context.Context, cost time.Duration, wid int) {
		select {
		case <-time.After(cost):
			fmt.Printf("Worker %d finished task\n", wid)
		case <-ctx.Done():
			fmt.Printf("Worker %d canceled\n", wid)
		}
	}

	// worker1 do task without context controlling
	worker2 := func(ctx context.Context, cost time.Duration, wid int) {
		select {
		case <-time.After(cost):
			fmt.Printf("Worker %d finished task\n", wid)
		}
	}

	// scheduled before close
	run(worker1, 100*time.Millisecond, 2*time.Second, 150*time.Millisecond)
	// not scheduled before close
	run(worker1, 100*time.Millisecond, 2*time.Second, time.Millisecond)
	// close timeout
	run(worker2, 10*time.Second, 2*time.Second, time.Millisecond)

	// Unordered output:
	// Worker 1 finished task
	// Worker 2 finished task
	// Worker 3 finished task
	// Worker 1 canceled
	// Worker 2 canceled
	// Worker 3 canceled
	// Close failed: [syn.Error:2] SYNAPSE_CLOSE_TIMEOUT
}
