package history

import (
	"ai-trainer/internal/adapter/hevy"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // Import uuid
)

// Service represents the history core service
type Service struct {
	repo Repository
	hevyClient *hevy.Client
}

// NewService creates a new history service
func NewService(repo Repository, hevyClient *hevy.Client) *Service {
	return &Service{
		repo: repo,
		hevyClient: hevyClient,
	}
}

// SyncRecentWorkouts fetches recent workouts from Hevy and syncs them to the database
func (s *Service) SyncRecentWorkouts(ctx context.Context) error {
	log.Println("Syncing recent Hevy workouts...")

	// Define a time window for "recent" workouts, e.g., last 7 days
	since := time.Now().AddDate(0, 0, -7) // Workouts from the last 7 days

	syncedCount := 0
	page := 1
	pageSize := 10 // Hevy API requires pageSize <= 10

	for {
		hevyWorkouts, err := s.hevyClient.FetchRecentWorkouts(page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to fetch recent workouts from Hevy (page %d): %w", page, err)
		}

		if len(hevyWorkouts) == 0 {
			break // No more workouts to fetch
		}

		for _, hw := range hevyWorkouts {
			// Only sync workouts that are within the 'since' time window
			if hw.StartTime.After(since) {
				workout := s.mapHevyWorkoutToHistoryWorkout(hw)
				err := s.repo.SaveWorkout(ctx, workout)
				if err != nil {
					log.Printf("Failed to save workout %s (Hevy ID: %s): %v", workout.WorkoutName, workout.HevyID, err)
					continue
				}
				syncedCount++
			}
		}

		if len(hevyWorkouts) < pageSize {
			break // Last page
		}
		page++
	}

	log.Printf("Successfully synced %d recent Hevy workouts.", syncedCount)
	return nil
}

// BackfillAllWorkouts fetches all historical workouts from Hevy and syncs them to the database
func (s *Service) BackfillAllWorkouts(ctx context.Context) error {
	log.Println("Starting full backfill of Hevy workout history...")

	syncedCount := 0
	page := 1
	pageSize := 10 // Hevy API requires pageSize <= 10

	// Define a very old start time and current end time for full backfill
	startTime := time.Time{} // Zero time, effectively fetching all history
	endTime := time.Now()

	for {
		hevyWorkouts, err := s.hevyClient.FetchWorkoutsByDateRange(page, pageSize, startTime, endTime)
		if err != nil {
			return fmt.Errorf("failed to fetch all historical workouts from Hevy (page %d): %w", page, err)
		}

		if len(hevyWorkouts) == 0 {
			break // No more workouts to fetch
		}

		for _, hw := range hevyWorkouts {
			workout := s.mapHevyWorkoutToHistoryWorkout(hw)
			err := s.repo.SaveWorkout(ctx, workout)
			if err != nil {
				log.Printf("Failed to save workout %s (Hevy ID: %s) during backfill: %v", workout.WorkoutName, workout.HevyID, err)
				continue
			}
			syncedCount++
		}

		if len(hevyWorkouts) < pageSize {
			break // Last page
		}
		page++
	}

	log.Printf("Successfully backfilled %d historical Hevy workouts.", syncedCount)
	return nil
}

// mapHevyWorkoutToHistoryWorkout converts a hevy.Workout to a history.Workout
func (s *Service) mapHevyWorkoutToHistoryWorkout(hw hevy.Workout) Workout {
	exercises := make([]Exercise, len(hw.Exercises))
	for i, he := range hw.Exercises {
		sets := make([]Set, len(he.Sets))
		for j, hs := range he.Sets {
			var weight float64
			if hs.WeightKg != nil {
				weight = *hs.WeightKg
			}
			var reps int
			if hs.Reps != nil {
				reps = *hs.Reps
			}
			var rpe float64
			if hs.RPE != nil {
				rpe = float64(*hs.RPE) // Convert int pointer to float64
			}

			sets[j] = Set{
				ID:     uuid.New().String(), // Generate new ID
				Weight: weight,
				Reps:   reps,
				RPE:    rpe,
			}
		}
		exercises[i] = Exercise{
			ID:           uuid.New().String(), // Generate new ID
			ExerciseName: he.Title,
			Sets:         sets,
			Notes:        he.Notes,
		}
	}

	duration := 0
	if !hw.EndTime.IsZero() && !hw.StartTime.IsZero() {
		duration = int(hw.EndTime.Sub(hw.StartTime).Seconds())
	}

	return Workout{
		WorkoutName: hw.Title,
		Date:        hw.StartTime,
		Duration:    duration,
		Exercises:   exercises,
		Notes:       hw.Description, // Map Hevy's Description to History's Notes
		HevyID:      hw.ID,
		// CreatedAt and UpdatedAt will be set by the repository
	}
}
