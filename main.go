package main

import (
	"concurrency_example/workerPool"
	"context"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"io"
	"log"
	"net/http"
	"strings"
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
	_, err = io.ReadAll(res.Body)
	if err != nil {
		t.logger.Fatalln("Failed to read response body", err)
		return err
	}
	//t.logger.Println("Response status code:", res.StatusCode)
	//t.logger.Println("Response size:", len(data))

	return nil
}

func (t *ExampleTask) OnFailure(err error) {
	t.logger.Fatalln("Failed to execute task", t.id, t.name, err)
	panic(err)
}

func main() {
	ctx := context.Background()
	logger := &log.Logger{}
	numWorkers := 5
	channelSize := 25

	cMap := cmap.New[[]string]()
	pool, err := workerPool.NewWorkerPool(numWorkers, channelSize, logger, &cMap)
	if err != nil {
		panic(err)
	}
	pool.Start()

	wg := &sync.WaitGroup{}

	for i := 0; i < channelSize; i++ {
		wg.Add(1)
		timeoutCtx, _ := context.WithTimeout(ctx, 10*time.Second)
		task := &ExampleTask{
			id:         i,
			name:       "request",
			url:        "https://google.com",
			wg:         wg,
			logger:     logger,
			timeoutCtx: timeoutCtx,
		}

		pool.AddWork(task)
	}
	wg.Wait()

	pool.Stop()

	time.Sleep(500 * time.Millisecond)
	cMap.IterCb(func(key string, value []string) {
		fmt.Printf("Worker: %s, Tasks(%d): %s\n", key, len(value), strings.Join(value, ", "))
	})
}
