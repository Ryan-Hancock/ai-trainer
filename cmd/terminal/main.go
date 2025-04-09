package main

import (
	"ai-trainer/internal/adapter/hevy"
	"ai-trainer/internal/adapter/llm"
	"ai-trainer/internal/core/plan"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

func main() {
	hevyAdapter := hevy.NewClient(os.Getenv("HEVYKEY"))
	llmAdapter := llm.NewClient(os.Getenv("OPENKEY"))
	planService, err := plan.NewService(plan.Config{
		LLM:  llmAdapter,
		Hevy: hevyAdapter,
	})
	if err != nil {
		fmt.Printf("Failed to create plan service: %v\n", err)
		return
	}

	for {
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{"Generate Workout Plan", "Check-In", "Nutrition Tip", "Quit"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Generate Workout Plan":
			handleGeneratePlan(planService)
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
