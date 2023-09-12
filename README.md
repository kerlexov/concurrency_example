# concurrency_example

This is a simple example of how to use the `worker pool` and `concurrency` patterns in Go

Worker pool is used to limit the number of goroutines that can run concurrently. This is useful when you have a lot of tasks to run concurrently and you don't want to overload the system.

Concurrency is used to run multiple tasks concurrently. This is useful when you have a lot of tasks to run concurrently and you want to make sure that all of them are executed.


## Usage
Run the following command to run the example:

```go run main.go```

Race condition can be detected by running the following command:


```go run main.go -race```

Run the following command to run the tests:

```go test -v ./...```

