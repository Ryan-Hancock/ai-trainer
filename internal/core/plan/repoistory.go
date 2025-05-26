package plan

import (
	"ai-trainer/internal/adapter/database"
	"database/sql"
	"fmt"
	"log"
)

type Client struct {
	*database.Client
}

// NewClient creates a new SQLite database client
func NewClient(db *database.Client) (*Client, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}
	c := &Client{db}
	if err := createTables(c.DB); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	return c, nil
}

// createTables creates the necessary tables in the database
func createTables(db *sql.DB) error {
	// Create a table for storing AI-generated workout plans
	createWorkoutPlansTableSQL := `CREATE TABLE IF NOT EXISTS workout_plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		goal TEXT NOT NULL,
		plan TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createWorkoutPlansTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create workout_plans table: %w", err)
	}

	log.Println("Database tables checked/created successfully.")
	return nil
}

// SaveWorkoutPlan saves an AI-generated workout plan to the database
func (c *Client) SaveWorkoutPlan(goal string, planContent string) error {
	insertSQL := `INSERT INTO workout_plans(goal, plan) VALUES (?, ?)`
	_, err := c.DB.Exec(insertSQL, goal, planContent)
	if err != nil {
		return fmt.Errorf("failed to insert workout plan: %w", err)
	}
	log.Printf("Workout plan saved for goal: %s", goal)
	return nil
}

// GetRecentWorkoutPlans retrieves recent workout plans from the database
func (c *Client) GetRecentWorkoutPlans(limit int) ([]map[string]string, error) {
	selectSQL := `SELECT goal, plan, created_at FROM workout_plans ORDER BY created_at DESC LIMIT ?`
	rows, err := c.DB.Query(selectSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout plans: %w", err)
	}
	defer rows.Close()

	var plans []map[string]string
	for rows.Next() {
		var goal, planContent, createdAt string
		if err := rows.Scan(&goal, &planContent, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan workout plan row: %w", err)
		}
		plans = append(plans, map[string]string{
			"goal":       goal,
			"plan":       planContent,
			"created_at": createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning workout plan rows: %w", err)
	}

	return plans, nil
}

// Interface defines the database client interface
type Repoistory interface {
	SaveWorkoutPlan(goal string, planContent string) error
	GetRecentWorkoutPlans(limit int) ([]map[string]string, error)
}