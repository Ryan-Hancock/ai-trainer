package background

import (
	"ai-trainer/internal/core/history" // Import history service
	"ai-trainer/internal/core/plan"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// CreateWorkoutPlanTask creates a task that generates workout plans and saves them to the database
func CreateWorkoutPlanTask(planService *plan.Service, goal string) Task {
	return Task{
		Name:     "Generate Workout Plan",
		Interval: 24 * time.Hour, // Generate a new plan daily
		Execute: func(ctx context.Context) error {
			log.Println("Generating workout plan...")

			workout, err := planService.GenerateWorkoutPlan(goal)
			if err != nil {
				return fmt.Errorf("failed to generate workout plan: %w", err)
			}

			// The planService now saves the workout plan to the database,
			// so we just need to log it here.
			fmt.Println("Generated Workout Plan:")
			for _, exercise := range workout.Exercises {
				fmt.Printf("- %s: %s\n", exercise.Name, exercise.Description)
				fmt.Println("  Reps:", exercise.Reps)
				fmt.Println("  Weight:", exercise.Weight)
				fmt.Println("  Body Part:", exercise.BodyPart)
				fmt.Println("  ----------------------")
			}

			return nil
		},
	}
}

// CreateHevySyncTask creates a task that syncs recent Hevy workouts
func CreateHevySyncTask(historyService *history.Service) Task {
	return Task{
		Name:     "Sync Recent Hevy Workouts",
		Interval: 6 * time.Hour, // Sync every 6 hours
		Execute: func(ctx context.Context) error {
			return historyService.SyncRecentWorkouts(ctx)
		},
	}
}

// CreateWorkoutAnalysisTask creates a task that analyzes recent workouts
// This is a placeholder for future implementation
func CreateWorkoutAnalysisTask(planService *plan.Service) Task {
	return Task{
		Name:     "Analyze Recent Workouts",
		Interval: 12 * time.Hour, // Analyze workouts twice daily
		Execute: func(ctx context.Context) error {
			log.Println("Analyzing recent workouts...")

			// TODO: Implement workout analysis logic
			// 1. Fetch recent workouts from Hevy
			// 2. Analyze workout patterns, progress, etc.
			// 3. Generate insights and recommendations

			log.Println("Workout analysis completed")
			return nil
		},
	}
}

// CreateNutritionTipTask creates a task that provides nutrition tips
// This is a placeholder for future implementation
func CreateNutritionTipTask(planService *plan.Service) Task {
	return Task{
		Name:     "Provide Nutrition Tip",
		Interval: 48 * time.Hour, // Provide tips every 2 days
		Execute: func(ctx context.Context) error {
			log.Println("Generating nutrition tip...")

			// TODO: Implement nutrition tip generation
			// This could use the LLM to generate personalized nutrition tips
			// based on workout history and goals

			log.Println("Nutrition tip generated")
			return nil
		},
	}
}

// CreateProgressReportTask creates a task that generates progress reports
// This is a placeholder for future implementation
func CreateProgressReportTask(planService *plan.Service) Task {
	return Task{
		Name:     "Generate Progress Report",
		Interval: 7 * 24 * time.Hour, // Weekly progress report
		Execute: func(ctx context.Context) error {
			log.Println("Generating progress report...")

			// TODO: Implement progress report generation
			// 1. Fetch workout history for the past week/month
			// 2. Analyze progress towards goals
			// 3. Generate a report with insights and recommendations

			log.Println("Progress report generated")
			return nil
		},
	}
}

// CreateRetrieveWorkoutPlansTask creates a task that retrieves recent workout plans from the database
func CreateRetrieveWorkoutPlansTask(planService *plan.Service, limit int) Task {
	return Task{
		Name:     "Retrieve Recent Workout Plans",
		Interval: 1 * time.Hour, // Retrieve recent plans hourly (example interval)
		Execute: func(ctx context.Context) error {
			log.Printf("Retrieving recent workout plans (limit: %d)...", limit)

			plans, err := planService.DB.GetRecentWorkoutPlans(limit)
			if err != nil {
				return fmt.Errorf("failed to retrieve recent workout plans: %w", err)
			}

			if len(plans) == 0 {
				log.Println("No recent workout plans found in the database.")
				return nil
			}

			log.Println("Recent Workout Plans:")
			for _, planData := range plans {
				fmt.Printf("Goal: %s\n", planData["goal"])
				fmt.Printf("Created At: %s\n", planData["created_at"])

				// Unmarshal and format the plan JSON for better readability
				var workout plan.Workout
				err := json.Unmarshal([]byte(planData["plan"]), &workout)
				if err != nil {
					log.Printf("Failed to unmarshal workout plan JSON: %v", err)
					fmt.Printf("Plan (Raw JSON): %s\n", planData["plan"])
				} else {
					fmt.Println("Plan:")
					for _, exercise := range workout.Exercises {
						fmt.Printf("  - %s: %s\n", exercise.Name, exercise.Description)
						fmt.Printf("    Reps: %d\n", exercise.Reps)
						fmt.Printf("    Weight: %.2f\n", exercise.Weight)
						fmt.Printf("    Body Part: %s\n", exercise.BodyPart)
						fmt.Println("    ----------------------")
					}
				}
				fmt.Println("----------------------")
			}

			return nil
		},
	}
}
