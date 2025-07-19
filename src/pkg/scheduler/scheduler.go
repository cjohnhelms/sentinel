package scheduler

import (
	"context"
	"fmt"
	"github.com/cjohnhelms/sentinel/pkg/event"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type Task struct {
	ID         uuid.UUID
	TaskFunc   func(ctx context.Context, eventChan chan []event.Event)
	RunOnStart bool
	Scheduled  bool
	Cron       string
}

type Scheduler struct {
	wg           *sync.WaitGroup
	mu           sync.Mutex
	ctx          context.Context
	eventChan    chan []event.Event
	tasks        []Task
	runningTasks []uuid.UUID
}

func NewScheduler(ctx context.Context, eventChan chan []event.Event, tasks ...Task) *Scheduler {
	return &Scheduler{
		wg:           new(sync.WaitGroup),
		eventChan:    eventChan,
		tasks:        tasks,
		runningTasks: []uuid.UUID{},
		mu:           sync.Mutex{},
		ctx:          ctx,
	}
}

func (s *Scheduler) runner(task Task) {
	defer s.wg.Done()

	// initial run
	if task.RunOnStart {
		// check if running
		s.mu.Lock()
		for _, id := range s.runningTasks {
			if id == task.ID {
				slog.Warn(fmt.Sprintf("task already running, skipping %s", task.ID))
			} else {
				s.runningTasks = append(s.runningTasks, id)
			}
		}
		s.mu.Unlock()
		// run task
		task.TaskFunc(s.ctx, s.eventChan)

		s.mu.Lock()
		for i, id := range s.runningTasks {
			if id == task.ID {
				slices.Delete(s.runningTasks, i, i)
			}
		}
		s.mu.Unlock()
	}

	// scheduled tasks
	if task.Scheduled {
		for {
			nextRun := cronexpr.MustParse(task.Cron).Next(time.Now())
			timer := time.NewTimer(nextRun.Sub(time.Now()))
			select {
			case <-timer.C:
				// check if running
				s.mu.Lock()
				for _, id := range s.runningTasks {
					if id == task.ID {
						slog.Warn(fmt.Sprintf("task already running, skipping %s", task.ID))
						continue
					} else {
						s.runningTasks = append(s.runningTasks, id)
					}
				}
				s.mu.Unlock()

				// run task
				task.TaskFunc(s.ctx, s.eventChan)

				// remove from runningTasks
				s.mu.Lock()
				for i, id := range s.runningTasks {
					if id == task.ID {
						slices.Delete(s.runningTasks, i, i)
					}
				}
				s.mu.Unlock()
				timer.Stop()
			case <-s.ctx.Done():
				timer.Stop()
				slog.Info("task cancelled")
				return
			}
		}
	}

}

func (s *Scheduler) Run() {
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runner(task)
	}
	s.wg.Wait()
	slog.Debug("scheduler finished")
}
