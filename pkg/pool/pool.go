package pool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Task struct {
	ID      string
	Process func() error
}

// Manage a pool of goroutine
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool
}

func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	log.Printf("All %d workers are now running", wp.workerCount)
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				// Channel closed, shutdown worker
				log.Printf("Worker %d: task queue closed", id)
				return
			}

			// Process the task
			startTime := time.Now()

			if err := task.Process(); err != nil {
				log.Printf("Worker %d: task %s failed: %v", id, task.ID, err)
			} else {
				duration := time.Since(startTime)
				log.Printf("Worker %d: task %s completed in %v", id, task.ID, duration)
			}
		}
	}
}

// Add a new task to the queue
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	case wp.taskQueue <- task:
		log.Printf("Task %s submitted to queue", task.ID)
		return nil
	}
}

// Submit async without blocking
func (wp *WorkerPool) SubmitAsync(task Task) error {
	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	case wp.taskQueue <- task:
		log.Printf("Task %s submitted to queue (async)", task.ID)
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown() {
	log.Println("Initial worker pool shutdown...")
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	log.Println("Worker pool shutdown completely")
}

// number of task waiting in queue
func (wp *WorkerPool) QueueLength() int {
	return len(wp.taskQueue)
}

// number of worker in pool
func (wp *WorkerPool) WorkerCount() int {
	return wp.workerCount
}
