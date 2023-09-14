package workerPool

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type WorkerPool struct {
	numWorkers int
	tasks      chan Task
	result     chan any
	error      chan error
	logger     *log.Logger
	// ensure the pool can only be started once
	start sync.Once
	// ensure the pool can only be stopped once
	stop sync.Once
	// close to signal the workers to stop working
	quit chan struct{}
}

func (wp *WorkerPool) Start() {
	wp.start.Do(func() {
		wp.logger.Println("starting worker pool")
		wp.startWorkers()
	})
}

func (wp *WorkerPool) startWorkers() {
	for i := 0; i < wp.numWorkers; i++ {
		go func(workerNum int) {
			wp.logger.Printf("starting worker %d", workerNum)
			client := newHttpClient()

			for {
				select {
				case <-wp.quit:
					wp.logger.Printf("stopping worker %d with quit channel\n", workerNum)
					return
				case task, ok := <-wp.tasks:
					if !ok {
						wp.logger.Printf("stopping worker %d with closed tasks channel\n", workerNum)
						return
					}

					wp.logger.Printf("Started task %s on worker %d", task.GetName(), workerNum)

					if err := task.Execute(client); err != nil {
						task.OnFailure(err)
					}
				}
			}
		}(i)
		time.Sleep(1 * time.Millisecond)
	}
}

func (wp *WorkerPool) Stop() {
	wp.stop.Do(func() {
		wp.logger.Println("stopping worker pool")
		close(wp.quit)
	})
}

func (wp *WorkerPool) AddWork(task Task) {
	select {
	case wp.tasks <- task:
	case <-wp.quit:
	}
}

var ErrNoWorkers = errors.New("cannot create a pool with no workers")
var ErrNegativeChannelSize = errors.New("cannot create a pool with a negative channel size")

func NewWorkerPool(numWorkers int, channelSize int, logger *log.Logger) (Pool, error) {
	if numWorkers <= 0 {
		return nil, ErrNoWorkers
	}
	if channelSize < 0 {
		return nil, ErrNegativeChannelSize
	}

	tasks := make(chan Task, channelSize)
	result := make(chan any)
	errs := make(chan error)
	return &WorkerPool{
		numWorkers: numWorkers,
		tasks:      tasks,
		logger:     logger,
		result:     result,
		error:      errs,
		start:      sync.Once{},
		stop:       sync.Once{},
		quit:       make(chan struct{}),
	}, nil
}

func (wp *WorkerPool) GetResultChan() chan any {
	return wp.result
}
func (wp *WorkerPool) GetErrorChan() chan error {
	return wp.error
}

func newHttpClient() *http.Client {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	transport := &http.Transport{
		Proxy:                 defaultTransport.Proxy,
		DialContext:           defaultTransport.DialContext,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		ResponseHeaderTimeout: defaultTransport.ResponseHeaderTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &http.Client{Transport: transport, Timeout: time.Duration(2) * time.Second}
}
