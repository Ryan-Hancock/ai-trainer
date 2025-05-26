package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Client represents the SQLite database client
type Client struct {
	db *sql.DB
}

// NewClient creates a new SQLite database client
func NewClient(databasePath string) (*Client, error) {
	// Check if the database file exists, if not, create it
	if _, err := os.Stat(databasePath); os.IsNotExist(err) {
		file, err := os.Create(databasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
		log.Printf("Database file created at: %s", databasePath)
	}

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
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
	_, err := c.db.Exec(insertSQL, goal, planContent)
	if err != nil {
		return fmt.Errorf("failed to insert workout plan: %w", err)
	}
	log.Printf("Workout plan saved for goal: %s", goal)
	return nil
}

// GetRecentWorkoutPlans retrieves recent workout plans from the database
func (c *Client) GetRecentWorkoutPlans(limit int) ([]map[string]string, error) {
	selectSQL := `SELECT goal, plan, created_at FROM workout_plans ORDER BY created_at DESC LIMIT ?`
	rows, err := c.db.Query(selectSQL, limit)
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
type Interface interface {
	Close() error
	SaveWorkoutPlan(goal string, planContent string) error
	GetRecentWorkoutPlans(limit int) ([]map[string]string, error)
}
