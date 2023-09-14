package main

import (
	"concurrency_example/workerPool"
	"context"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
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
	logger     *log.Logger
	timeoutCtx context.Context
	result     chan any
}

func (t *ExampleTask) GetName() string {
	return fmt.Sprintf("%s - %d", t.name, t.id)
}

func (t *ExampleTask) Execute() error {
	//t.logger.Println("Executing task", t.GetName())
	if t.wg != nil {
		defer t.wg.Done()
	}

	req, err := http.NewRequestWithContext(t.timeoutCtx, "GET", t.url, nil)
	if err != nil {
		t.logger.Fatalln("Failed to create request", err)
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.logger.Fatalln("Failed to execute request", err)
		return err
	}

	defer res.Body.Close()
	//data, err := io.ReadAll(res.Body)
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.logger.Fatalln("Failed to read response body", err)
		return err
	}
	//t.logger.Println("Response status code:", res.StatusCode)
	//t.logger.Println("Response size:", len(data))

	t.result <- data
	return nil
}

func (t *ExampleTask) OnFailure(err error) {
	t.logger.Fatalln("Failed to execute task", t.id, t.name, err)
}

func main() {
	ctx := context.Background()
	logger := &log.Logger{}
	var result [][]byte
	numWorkers := 5
	channelSize := 150

	cMap := cmap.New[[]string]()
	pool, err := workerPool.NewWorkerPool(numWorkers, channelSize, logger, &cMap)
	if err != nil {
		panic(err)
	}
	pool.Start()

	wg := &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < channelSize; i++ {
		wg.Add(1)
		timeoutCtx, _ := context.WithTimeout(ctx, 5*time.Second)
		task := &ExampleTask{
			id:         i,
			name:       "request",
			url:        "https://google.com",
			wg:         wg,
			logger:     logger,
			timeoutCtx: timeoutCtx,
			result:     pool.GetResultChan(),
		}

		pool.AddWork(task)
	}

	go func() {
		for {
			select {
			case res := <-pool.GetResultChan():
				result = append(result, res.([]byte))
			}
		}
	}()
	wg.Wait()

	pool.Stop()
	fmt.Printf("Took %s\n", time.Since(start))

	time.Sleep(500 * time.Millisecond)

	fmt.Printf("Result: %d, Expected %d\n", len(result), channelSize)
	fmt.Println("Tasks per worker:")
	cMap.IterCb(func(key string, value []string) {
		fmt.Printf("Worker: %s, Tasks(%d)\n", key, len(value))
	})
}
