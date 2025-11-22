package workers

import (
	"context"
	"sync"

	"log/slog"
)

type TaskType string

const (
	TaskTypeProcessSuccess TaskType = "process_success"
	TaskTypeProcessFailed  TaskType = "process_failed"
)

type Task struct {
	Type TaskType
	Data map[string]interface{}
}

type Pool struct {
	workers   int
	taskChan  chan *Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	processor TaskProcessor
}

type TaskProcessor interface {
	ProcessTask(ctx context.Context, task *Task) error
}

func NewPool(workers int, processor TaskProcessor) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workers:   workers,
		taskChan:  make(chan *Task, 100),
		ctx:       ctx,
		cancel:    cancel,
		processor: processor,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) Stop() {
	close(p.taskChan)
	p.cancel()
	p.wg.Wait()
}

func (p *Pool) Submit(task *Task) {
	select {
	case p.taskChan <- task:
	default:
		slog.Warn("Task queue full, dropping task", "type", task.Type)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	slog.Info("Worker started", "id", id)

	for {
		select {
		case <-p.ctx.Done():
			slog.Info("Worker stopping", "id", id)
			return
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			if err := p.processor.ProcessTask(p.ctx, task); err != nil {
				slog.Error("Task processing failed", "worker_id", id, "task_type", task.Type, "error", err)
			}
		}
	}
}
