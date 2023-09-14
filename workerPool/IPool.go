package workerPool

type Pool interface {
	// Start gets the worker pool ready to process jobs, and should only be called once
	Start()
	// Stop stops the worker pool, tears down any required resources,
	// and should only be called once
	Stop()
	// AddWork adds a task for the worker pool to process. It is only valid after
	// Start() has been called and before Stop() has been called.
	AddWork(Task)
	// GetResultChan returns the channel that results are sent on
	GetResultChan() chan any
}

type Task interface {
	// Execute performs the work
	Execute() error
	// OnFailure handles any error returned from Execute()
	OnFailure(error)
	// GetName returns the name of the task
	GetName() string
}
