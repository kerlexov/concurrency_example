package workerPool

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"log"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool_BlockedAddWorkReleaseAfterStop(t *testing.T) {
	logger := &log.Logger{}
	cMap := cmap.New[[]string]()

	p, err := NewWorkerPool(1, 0, logger, &cMap)
	if err != nil {
		t.Fatal("error making worker pool:", err)
	}

	p.Start()

	wg := &sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		// the first should start processing right away, the second two should hang
		wg.Add(1)
		go func() {
			p.AddWork(newTestTask(func() error {
				time.Sleep(20 * time.Second)
				return nil
			}, false, nil))
			wg.Done()
		}()
	}

	done := make(chan struct{})
	p.Stop()
	go func() {
		// wait on our AddWork calls to complete, then signal on the done channel
		wg.Wait()
		done <- struct{}{}
	}()

	// wait until either we hit our timeout, or we're told the AddWork calls completed
	select {
	case <-time.After(1 * time.Second):
		t.Fatal("failed because still hanging on AddWork")
	case <-done:
		// this is the success case
	}
}
