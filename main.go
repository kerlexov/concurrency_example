package main

import (
	"concurrency_example/workerPool"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type ExampleTask struct {
	id         int
	name       string
	url        string
	wg         *sync.WaitGroup
	timeoutCtx context.Context
	result     chan any
	errors     chan error
}

func (t *ExampleTask) GetName() string {
	return fmt.Sprintf("%s - %d", t.name, t.id)
}

func (t *ExampleTask) Execute(client *http.Client) error {
	if t.wg != nil {
		defer t.wg.Done()
	}

	req, err := http.NewRequestWithContext(t.timeoutCtx, "GET", t.url, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	t.result <- data
	return nil
}

func (t *ExampleTask) OnFailure(err error) {
	t.errors <- err
}

func main() {
	ctx := context.Background()
	logger := log.New(log.Writer(), "", log.LstdFlags)
	result := make([][]byte, 0)
	errs := make([]error, 0)
	numWorkers := 500
	channelSize := 1500

	pool, err := workerPool.NewWorkerPool(numWorkers, channelSize, logger)
	if err != nil {
		panic(err)
	}
	pool.Start()

	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < channelSize; i++ {
		wg.Add(1)
		timeoutCtx, _ := context.WithTimeout(ctx, 2*time.Second)
		task := &ExampleTask{
			id:         i,
			name:       "request",
			url:        "https://example.com",
			wg:         wg,
			timeoutCtx: timeoutCtx,
			result:     pool.GetResultChan(),
			errors:     pool.GetErrorChan(),
		}

		pool.AddWork(task)
		time.Sleep(5 * time.Millisecond)
	}

	go func() {
		for {
			select {
			case res := <-pool.GetResultChan():
				result = append(result, res.([]byte))
			case errori := <-pool.GetErrorChan():
				errs = append(errs, errori)
			}
		}
	}()
	wg.Wait()
	totalTime := time.Since(start)
	pool.Stop()
	time.Sleep(500 * time.Millisecond)

	logger.Printf("Took %s\n", totalTime)
	logger.Printf("Result: %d, Expected %d\n", len(result), channelSize)
	logger.Printf("Errors  %d\n", len(errs))
}
