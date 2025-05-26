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
	DB *sql.DB
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

	return &Client{DB: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.DB.Close()
}