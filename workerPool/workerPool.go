package workerPool

import (
	"errors"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"log"
	"sync"
)

type WorkerPool struct {
	numWorkers int
	tasks      chan Task

	logger *log.Logger
	// ensure the pool can only be started once
	start sync.Once
	// ensure the pool can only be stopped once
	stop sync.Once

	workerTasks *cmap.ConcurrentMap[string, []string]

	// close to signal the workers to stop working
	quit chan struct{}
}

func (wp *WorkerPool) Start() {
	wp.start.Do(func() {
		log.Print("starting worker pool")
		wp.startWorkers()
	})
}

func (wp *WorkerPool) startWorkers() {
	for i := 0; i < wp.numWorkers; i++ {
		go func(workerNum int) {
			log.Printf("starting worker %d", workerNum)

			for {
				select {
				case <-wp.quit:
					log.Printf("stopping worker %d with quit channel\n", workerNum)
					return
				case task, ok := <-wp.tasks:
					if !ok {
						log.Printf("stopping worker %d with closed tasks channel\n", workerNum)
						return
					}

					log.Printf("Started task %s on worker %d", task.GetName(), workerNum)

					workerKey := fmt.Sprintf("W%d", workerNum)
					workerTasks, ok := wp.workerTasks.Get(workerKey)
					if ok {
						wp.workerTasks.Set(workerKey, append(workerTasks, task.GetName()))
					} else {
						wp.workerTasks.Set(workerKey, append([]string{}, task.GetName()))
					}

					if err := task.Execute(); err != nil {
						task.OnFailure(err)
					}
				}
			}
		}(i)
	}
}

func (wp *WorkerPool) Stop() {
	wp.stop.Do(func() {
		log.Print("stopping worker pool")
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

func NewWorkerPool(numWorkers int, channelSize int, logger *log.Logger, wt *cmap.ConcurrentMap[string, []string]) (Pool, error) {
	if numWorkers <= 0 {
		return nil, ErrNoWorkers
	}
	if channelSize < 0 {
		return nil, ErrNegativeChannelSize
	}

	tasks := make(chan Task, channelSize)

	return &WorkerPool{
		numWorkers:  numWorkers,
		tasks:       tasks,
		logger:      logger,
		start:       sync.Once{},
		stop:        sync.Once{},
		workerTasks: wt,
		quit:        make(chan struct{}),
	}, nil
}
