package main

import (
	"ai-trainer/internal/adapter/database" // Import database
	"ai-trainer/internal/adapter/hevy"
	"ai-trainer/internal/adapter/llm"
	"ai-trainer/internal/core/history" // Import history service
	"ai-trainer/internal/core/plan"
	"context" // Import context
	"fmt"
	"log" // Import log for backfill output
	"os"

	"github.com/manifoldco/promptui"
)

func main() {
	hevyAPIKey := os.Getenv("HEVYKEY")
	openAIKey := os.Getenv("OPENKEY")
	databasePath := os.Getenv("DATABASE_PATH")

	if hevyAPIKey == "" || openAIKey == "" {
		fmt.Println("Missing required environment variables. Please set HEVYKEY and OPENKEY.")
		os.Exit(1)
	}

	if databasePath == "" {
		databasePath = "./ai_trainer.db" // Default database path
		log.Printf("DATABASE_PATH not set, using default: %s", databasePath)
	}

	// Initialize adapters
	hevyAdapter := hevy.NewClient(hevyAPIKey)
	llmAdapter := llm.NewClient(openAIKey)
	dbAdapter, err := database.NewClient(databasePath)
	if err != nil {
		fmt.Printf("Failed to create database client: %v\n", err)
		os.Exit(1)
	}
	defer dbAdapter.Close()

	// Initialize plan service
	planService, err := plan.NewService(plan.Config{
		LLM:  llmAdapter,
		Hevy: hevyAdapter,
		DB:   dbAdapter,
	})
	if err != nil {
		fmt.Printf("Failed to create plan service: %v\n", err)
		os.Exit(1)
	}

	// Initialize history repository and service
	historyRepo, err := history.NewSQLiteRepository(dbAdapter)
	if err != nil {
		fmt.Printf("Failed to create history repository: %v\n", err)
		os.Exit(1)
	}
	historyService := history.NewService(historyRepo, hevyAdapter)

	for {
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{"Generate Workout Plan", "Backfill Hevy Workouts", "Check-In", "Nutrition Tip", "Quit"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Generate Workout Plan":
			handleGeneratePlan(planService)
		case "Backfill Hevy Workouts":
			handleBackfillHevyWorkouts(historyService)
		case "Check-In":
			handleCheckIn()
		case "Nutrition Tip":
			handleNutritionTip()
		case "Quit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}

func handleGeneratePlan(planService *plan.Service) {
	goalPrompt := promptui.Prompt{
		Label: "What is your primary goal? (e.g., build muscle, lose fat)",
	}
	goal, _ := goalPrompt.Run()

	fmt.Printf("Generating a workout plan for goal: %s \n", goal)
	workout, err := planService.GenerateWorkoutPlan(goal)
	if err != nil {
		fmt.Printf("Error generating workout plan: %v\n", err)
		return
	}
	fmt.Println("Generated Workout Plan:")
	for _, exercise := range workout.Exercises {
		fmt.Printf("- %s: %s\n", exercise.Name, exercise.Description)
		fmt.Println("  Reps:", exercise.Reps)
		fmt.Println("  Weight:", exercise.Weight)
		fmt.Println("  Body Part:", exercise.BodyPart)
		fmt.Println("  ----------------------")
	}
}

func handleBackfillHevyWorkouts(historyService *history.Service) {
	fmt.Println("Starting Hevy workout history backfill...")
	ctx := context.Background() // Use a background context for the backfill
	err := historyService.BackfillAllWorkouts(ctx)
	if err != nil {
		fmt.Printf("Error during Hevy workout backfill: %v\n", err)
		return
	}
	fmt.Println("Hevy workout history backfill completed.")
}

func handleCheckIn() {
	moodPrompt := promptui.Prompt{
		Label: "How was your energy today? (1-10)",
	}
	mood, _ := moodPrompt.Run()

	fmt.Printf("Thanks for checking in. You rated your energy: %s\n", mood)
	// Could trigger a log entry or LLM response
}

func handleNutritionTip() {
	fmt.Println("Fetching a nutrition tip...")
	// Call LLM or predefined tips
	fmt.Println("💡 Tip: Prioritize protein at every meal to support muscle recovery.")
}
