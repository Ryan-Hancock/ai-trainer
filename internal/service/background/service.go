package background

import (
	"ai-trainer/internal/core/plan"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Service represents the background service that runs periodic tasks
type Service struct {
	planService *plan.Service
	tasks       []Task
	interval    time.Duration
	mutex       sync.Mutex
	isRunning   bool
}

// Task represents a background task that can be scheduled
type Task struct {
	Name     string
	Interval time.Duration
	LastRun  time.Time
	Execute  func(ctx context.Context) error
}

// Config contains configuration for the background service
type Config struct {
	PlanService *plan.Service
	Interval    time.Duration
}

// NewService creates a new background service
func NewService(cfg Config) *Service {
	if cfg.Interval == 0 {
		// Default to 6 hours if not specified
		cfg.Interval = 6 * time.Hour
	}

	return &Service{
		planService: cfg.PlanService,
		interval:    cfg.Interval,
		tasks:       make([]Task, 0),
	}
}

// AddTask adds a new task to the service
func (s *Service) AddTask(task Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.tasks = append(s.tasks, task)
}

// Start begins the background service
func (s *Service) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.isRunning {
		s.mutex.Unlock()
		return fmt.Errorf("service is already running")
	}
	s.isRunning = true
	s.mutex.Unlock()

	log.Println("Starting background service...")

	// Run immediately on startup
	s.runTasks(ctx)

	// Then run on the schedule
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runTasks(ctx)
		case <-ctx.Done():
			log.Println("Background service shutting down...")
			s.mutex.Lock()
			s.isRunning = false
			s.mutex.Unlock()
			return nil
		}
	}
}

// runTasks executes all due tasks
func (s *Service) runTasks(ctx context.Context) {
	s.mutex.Lock()
	tasks := make([]Task, len(s.tasks))
	copy(tasks, s.tasks)
	s.mutex.Unlock()

	now := time.Now()
	for i, task := range tasks {
		// Check if the task is due to run
		if task.LastRun.IsZero() || now.Sub(task.LastRun) >= task.Interval {
			log.Printf("Running task: %s", task.Name)
			
			// Execute the task
			err := task.Execute(ctx)
			if err != nil {
				log.Printf("Error executing task %s: %v", task.Name, err)
			}
			
			// Update the last run time
			s.mutex.Lock()
			s.tasks[i].LastRun = now
			s.mutex.Unlock()
		}
	}
}

// Stop stops the background service
func (s *Service) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.isRunning = false
}

// IsRunning returns whether the service is currently running
func (s *Service) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isRunning
}
