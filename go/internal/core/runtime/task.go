package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/acore2026/ueransim-go/internal/core/logging"
)

type MessageType string

const (
	MessageTypeTick MessageType = "tick"
)

type Message struct {
	Type    MessageType
	Payload any
}

type TaskHandler interface {
	OnStart(context.Context, *Task) error
	OnMessage(context.Context, Message) error
	OnStop(context.Context) error
}

type Task struct {
	name    string
	logger  logging.Logger
	handler TaskHandler
	inbox   chan Message
}

func NewTask(name string, logger logging.Logger, handler TaskHandler, buffer int) *Task {
	if buffer <= 0 {
		buffer = 16
	}

	return &Task{
		name:    name,
		logger:  logger.With("task", name),
		handler: handler,
		inbox:   make(chan Message, buffer),
	}
}

func (t *Task) Send(msg Message) error {
	select {
	case t.inbox <- msg:
		return nil
	default:
		return fmt.Errorf("task %s inbox full", t.name)
	}
}

func (t *Task) SetHandler(handler TaskHandler) {
	t.handler = handler
}

func (t *Task) Run(ctx context.Context) error {
	t.logger.Info("starting")
	defer t.logger.Info("stopped")

	if err := t.handler.OnStart(ctx, t); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			stopErr := t.handler.OnStop(context.WithoutCancel(ctx))
			if stopErr != nil && !errors.Is(stopErr, context.Canceled) {
				return stopErr
			}
			return nil
		case msg := <-t.inbox:
			if err := t.handler.OnMessage(ctx, msg); err != nil {
				return err
			}
		}
	}
}

type PeriodicTask struct {
	interval time.Duration
	logger   logging.Logger
}

func NewPeriodicTask(interval time.Duration, logger logging.Logger) PeriodicTask {
	return PeriodicTask{interval: interval, logger: logger}
}

func (p PeriodicTask) Start(ctx context.Context, task *Task) {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := task.Send(Message{Type: MessageTypeTick}); err != nil {
					p.logger.Debug("drop tick", "error", err)
				}
			}
		}
	}()
}

type Group struct {
	logger logging.Logger
	tasks  []*Task
}

func NewGroup(logger logging.Logger, tasks ...*Task) *Group {
	return &Group{logger: logger, tasks: tasks}
}

func (g *Group) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(g.tasks))
	var wg sync.WaitGroup

	for _, task := range g.tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task.Run(ctx); err != nil {
				errCh <- fmt.Errorf("%s: %w", task.name, err)
				cancel()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case err := <-errCh:
		<-done
		g.logger.Error("task group failed", "error", err)
		return err
	}
}
