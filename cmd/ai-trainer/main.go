package main

import (
	"ai-trainer/internal/adapter/database"
	"ai-trainer/internal/adapter/hevy"
	"ai-trainer/internal/adapter/llm"
	"ai-trainer/internal/core/history" // Import history package
	"ai-trainer/internal/core/plan"
	"ai-trainer/internal/service/background"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting AI Trainer Assistant...")

	// Load environment variables
	hevyAPIKey := os.Getenv("HEVYKEY")
	openAIKey := os.Getenv("OPENKEY")
	databasePath := os.Getenv("DATABASE_PATH")

	if hevyAPIKey == "" || openAIKey == "" {
		log.Fatal("Missing required environment variables. Please set HEVYKEY and OPENKEY.")
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
		log.Fatalf("Failed to create database client: %v", err)
	}
	defer dbAdapter.Close()

	// Initialize plan service
	planService, err := plan.NewService(plan.Config{
		LLM:  llmAdapter,
		Hevy: hevyAdapter,
		DB:   dbAdapter,
	})
	if err != nil {
		log.Fatalf("Failed to create plan service: %v", err)
	}

	// Initialize history repository and service
	historyRepo, err := history.NewSQLiteRepository(dbAdapter)
	if err != nil {
		log.Fatalf("Failed to create history repository: %v", err)
	}
	historyService := history.NewService(historyRepo, hevyAdapter)

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and configure the background service
	// For development/testing, you might want a shorter interval
	bgService := background.NewService(background.Config{
	    PlanService: planService,
	    Interval:    1 * time.Minute,
	})
	// bgService := background.NewService(background.Config{
	// 	PlanService: planService,
	// 	Interval:    6 * time.Hour,
	// })

	// Add tasks to the background service
	bgService.AddTask(background.CreateWorkoutPlanTask(planService, "maintain fitness"))
	bgService.AddTask(background.CreateHevySyncTask(historyService)) // Add Hevy sync task
	bgService.AddTask(background.CreateWorkoutAnalysisTask(planService))
	bgService.AddTask(background.CreateNutritionTipTask(planService))
	bgService.AddTask(background.CreateProgressReportTask(planService))

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the background service
	go func() {
		if err := bgService.Start(ctx); err != nil {
			log.Fatalf("Background service error: %v", err)
		}
	}()

	log.Println("AI Trainer Assistant is running in the background. Press Ctrl+C to exit.")

	// Wait for termination signal
	sig := <-signalChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	// Give tasks a moment to complete
	time.Sleep(1 * time.Second)
	log.Println("AI Trainer Assistant shutdown complete.")
}
