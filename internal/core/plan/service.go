package plan

import (
	"ai-trainer/internal/adapter/hevy"
	"ai-trainer/internal/adapter/llm"
	prompts "ai-trainer/internal/utilities/prompt"
	"context"
	"encoding/json"
	"fmt"
)

type Service struct {
	LLM  llm.Interface
	Hevy hevy.Interface
}

type Config struct {
	LLM  llm.Interface
	Hevy hevy.Interface
}

func NewService(cfg Config) (*Service, error) {
	err := prompts.LoadPromptsFromFile("internal/core/plan/workout.prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	return &Service{
		LLM:  cfg.LLM,
		Hevy: cfg.Hevy,
	}, nil
}

// GenerateWorkoutPlan generates a workout plan based on the user's goal
func (s *Service) GenerateWorkoutPlan(goal string) (Workout, error) {
	workouts, err := s.fetchRecentWorkouts(1, 1)
	if err != nil {
		return Workout{}, fmt.Errorf("failed to fetch recent workouts: %w", err)
	}

	promptText, err := prompts.GetPrompt("GenerateWorkoutPlan", map[string]string{
		"Goal":              goal,
		"PreviousExercises": workouts,
	})
	if err != nil {
		return Workout{}, err
	}

	resp, err := s.LLM.Complete(context.Background(), llm.CompletionRequest{
		Prompt: promptText,
		System: "You are a JSON-generating AI trainer.",
	})
	if err != nil {
		return Workout{}, err
	}

	var workout Workout
	err = json.Unmarshal([]byte(resp.Text), &workout)
	if err != nil {
		return Workout{}, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	return workout, nil
}

func (s *Service) fetchRecentWorkouts(page, pageSize int) (string, error) {
	workouts, err := s.Hevy.FetchRecentWorkouts(page, pageSize)
	if err != nil {
		return "", fmt.Errorf("failed to fetch recent workouts: %w", err)
	}
	if len(workouts) == 0 {
		return "", fmt.Errorf("no recent workouts found")
	}

	js, err := json.Marshal(workouts)
	return string(js), err
}

type Interface interface {
	GenerateWorkoutPlan(goal string) (Workout, error)
}
