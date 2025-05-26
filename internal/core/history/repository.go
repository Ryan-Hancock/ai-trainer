package history

import (
	"ai-trainer/internal/adapter/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for interacting with workout history data
type Repository interface {
	SaveWorkout(ctx context.Context, workout Workout) error
	GetWorkoutByHevyID(ctx context.Context, hevyID string) (*Workout, error)
	GetRecentWorkouts(ctx context.Context, limit int) ([]Workout, error)
	GetAllWorkouts(ctx context.Context) ([]Workout, error)
	Migrate(ctx context.Context) error
}

// SQLiteRepository implements the Repository interface for SQLite
type SQLiteRepository struct {
	*database.Client
}

// NewSQLiteRepository creates a new SQLiteRepository
func NewSQLiteRepository(dbClient *database.Client) (*SQLiteRepository, error) {
	if dbClient == nil {
		return nil, fmt.Errorf("database client cannot be nil")
	}
	r := &SQLiteRepository{dbClient}
	if err := r.Migrate(context.Background()); err != nil { // Use context.Background() for migration
		return nil, fmt.Errorf("failed to migrate history tables: %w", err)
	}
	return r, nil
}

// SaveWorkout saves a workout to the database. If a workout with the same HevyID exists, it updates it.
func (r *SQLiteRepository) SaveWorkout(ctx context.Context, workout Workout) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	// Check if workout with HevyID already exists
	existingWorkout, err := r.GetWorkoutByHevyID(ctx, workout.HevyID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for existing workout: %w", err)
	}

	workoutJSON, err := json.Marshal(workout)
	if err != nil {
		return fmt.Errorf("failed to marshal workout to JSON: %w", err)
	}

	now := time.Now()
	if existingWorkout != nil {
		// Update existing workout
		workout.ID = existingWorkout.ID // Ensure ID is preserved
		workout.CreatedAt = existingWorkout.CreatedAt
		workout.UpdatedAt = now

		stmt, err := tx.PrepareContext(ctx, `
			UPDATE workouts SET
				workout_name = ?,
				date = ?,
				duration = ?,
				workout_data_json = ?, -- Use new column
				notes = ?,
				updated_at = ?
			WHERE hevy_id = ?
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare update statement: %w", err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx,
			workout.WorkoutName,
			workout.Date,
			workout.Duration,
			workoutJSON, // Store full workout JSON
			workout.Notes,
			workout.UpdatedAt,
			workout.HevyID,
		)
		if err != nil {
			return fmt.Errorf("failed to update workout: %w", err)
		}
	} else {
		// Insert new workout
		workout.ID = uuid.New().String()
		workout.CreatedAt = now
		workout.UpdatedAt = now

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO workouts (id, workout_name, date, duration, workout_data_json, notes, hevy_id, created_at, updated_at) -- Use new column
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(ctx,
			workout.ID,
			workout.WorkoutName,
			workout.Date,
			workout.Duration,
			workoutJSON, // Store full workout JSON
			workout.Notes,
			workout.HevyID,
			workout.CreatedAt,
			workout.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert workout: %w", err)
		}
	}

	return tx.Commit()
}

// GetWorkoutByHevyID retrieves a workout by its Hevy ID
func (r *SQLiteRepository) GetWorkoutByHevyID(ctx context.Context, hevyID string) (*Workout, error) {
	row := r.DB.QueryRowContext(ctx, ` 
		SELECT workout_data_json
		FROM workouts
		WHERE hevy_id = ?
	`, hevyID)

	var workoutJSON []byte
	err := row.Scan(&workoutJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan workout by Hevy ID: %w", err)
	}

	var workout Workout
	err = json.Unmarshal(workoutJSON, &workout)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workout JSON: %w", err)
	}

	return &workout, nil
}

// GetRecentWorkouts retrieves a limited number of recent workouts
func (r *SQLiteRepository) GetRecentWorkouts(ctx context.Context, limit int) ([]Workout, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT workout_data_json
		FROM workouts
		ORDER BY date DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent workouts: %w", err)
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var workoutJSON []byte
		err := rows.Scan(&workoutJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent workout: %w", err)
		}

		var workout Workout
		err = json.Unmarshal(workoutJSON, &workout)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal workout JSON for recent workout: %w", err)
		}
		workouts = append(workouts, workout)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recent workout rows: %w", err)
	}

	return workouts, nil
}

// GetAllWorkouts retrieves all workouts
func (r *SQLiteRepository) GetAllWorkouts(ctx context.Context) ([]Workout, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT workout_data_json
		FROM workouts
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all workouts: %w", err)
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var workoutJSON []byte
		err := rows.Scan(&workoutJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan all workout: %w", err)
		}

		var workout Workout
		err = json.Unmarshal(workoutJSON, &workout)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal workout JSON for all workout: %w", err)
		}
		workouts = append(workouts, workout)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating all workout rows: %w", err)
	}

	return workouts, nil
}

// Migrate creates the workouts table if it doesn't exist
func (r *SQLiteRepository) Migrate(ctx context.Context) error {
	// Drop the old table if it exists to apply schema changes
	dropQuery := `DROP TABLE IF EXISTS workouts;`
	_, err := r.DB.ExecContext(ctx, dropQuery)
	if err != nil {
		return fmt.Errorf("failed to drop old workouts table: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS workouts (
		id TEXT PRIMARY KEY,
		workout_name TEXT NOT NULL,
		date DATETIME NOT NULL,
		duration INTEGER,
		workout_data_json TEXT, -- New column to store full workout JSON
		notes TEXT,
		hevy_id TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = r.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to migrate workouts table: %w", err)
	}
	return nil
}
