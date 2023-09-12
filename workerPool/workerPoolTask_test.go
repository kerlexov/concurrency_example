package workerPool

import (
	"fmt"
	"sync"
	"time"
)

type testTask struct {
	id          int
	name        string
	executeFunc func() error

	shouldErr bool
	wg        *sync.WaitGroup

	mFailure       *sync.Mutex
	failureHandled bool
}

func newTestTask(executeFunc func() error, shouldErr bool, wg *sync.WaitGroup) *testTask {
	return &testTask{
		executeFunc: executeFunc,
		shouldErr:   shouldErr,
		wg:          wg,
		mFailure:    &sync.Mutex{},
	}
}

func (t *testTask) GetName() string {
	return fmt.Sprintf("%s - %d", t.name, t.id)
}
func (t *testTask) Execute() error {
	if t.wg != nil {
		defer t.wg.Done()
	}

	if t.executeFunc != nil {
		return t.executeFunc()
	}

	// if no function provided, just wait and error if told to do so
	time.Sleep(50 * time.Millisecond)
	if t.shouldErr {
		return fmt.Errorf("planned Execute() error")
	}
	return nil
}

func (t *testTask) OnFailure(e error) {
	t.mFailure.Lock()
	defer t.mFailure.Unlock()

	t.failureHandled = true
}

func (t *testTask) hitFailureCase() bool {
	t.mFailure.Lock()
	defer t.mFailure.Unlock()

	return t.failureHandled
}
