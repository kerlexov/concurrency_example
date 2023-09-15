package workerPool

import (
	"log"
	"sync"
	"testing"
)

func TestWorkerPool_Work(t *testing.T) {
	var tasks []*testTask
	wg := &sync.WaitGroup{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		tasks = append(tasks, newTestTask(i, nil, false, wg))
	}

	logger := log.New(log.Writer(), "", log.LstdFlags)

	p, err := NewWorkerPool(5, len(tasks), logger)
	if err != nil {
		t.Fatal("error making worker pool:", err)
	}
	p.Start()

	for _, j := range tasks {
		p.AddWork(j)
	}

	// we'll get a timeout failure if the tasks weren't processed
	wg.Wait()

	for taskNum, task := range tasks {
		if task.hitFailureCase() {
			t.Fatalf("error function called on task %d when it shouldn't be", taskNum)
		}
	}
}
