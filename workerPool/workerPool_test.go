package workerPool

import (
	"errors"
	cmap "github.com/orcaman/concurrent-map/v2"
	"log"
	"testing"
)

func TestWorkerPool_NewPool(t *testing.T) {
	logger := &log.Logger{}
	cMap := cmap.New[[]string]()

	if _, err := NewWorkerPool(0, 0, logger, &cMap); !errors.Is(err, ErrNoWorkers) {
		t.Fatalf("expected error when creating pool with 0 workers, got: %v", err)
	}
	if _, err := NewWorkerPool(-1, 0, logger, &cMap); !errors.Is(err, ErrNoWorkers) {
		t.Fatalf("expected error when creating pool with -1 workers, got: %v", err)
	}
	if _, err := NewWorkerPool(1, -1, logger, &cMap); !errors.Is(err, ErrNegativeChannelSize) {
		t.Fatalf("expected error when creating pool with -1 channel size, got: %v", err)
	}

	p, err := NewWorkerPool(5, 0, logger, &cMap)
	if err != nil {
		t.Fatalf("expected no error creating pool, got: %v", err)
	}
	if p == nil {
		t.Fatal("NewSimplePool returned nil Pool for valid input")
	}
}

func TestWorkerPool_MultipleStartStopDontPanic(t *testing.T) {
	logger := &log.Logger{}
	cMap := cmap.New[[]string]()

	p, err := NewWorkerPool(5, 0, logger, &cMap)
	if err != nil {
		t.Fatal("error creating pool:", err)
	}

	// We're just checking to make sure multiple calls to start or stop
	// don't cause a panic
	p.Start()
	p.Start()

	p.Stop()
	p.Stop()
}
