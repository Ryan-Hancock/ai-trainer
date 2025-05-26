package hevy

import (
	"bytes" // Add bytes import
	"encoding/json"
	"fmt"
	"io/ioutil" // Add ioutil import
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: "https://api.hevyapp.com",
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doGET(path string, query url.Values) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)
	if query != nil {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("Error reading Hevy API response body for status %d: %v", resp.StatusCode, readErr)
		}
		// Restore the body for subsequent reads (e.g., by json.NewDecoder)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		log.Printf("Hevy API Error Response (Status: %d, URL: %s): %s", resp.StatusCode, fullURL, string(bodyBytes))
	}

	return resp, nil
}

type WorkoutResponse struct {
	Page      int       `json:"page"`
	PageCount int       `json:"page_count"`
	Workouts  []Workout `json:"workouts"`
}

type Workout struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Exercises   []Exercise `json:"exercises"`
}

type Exercise struct {
	Index              int    `json:"index"`
	Title              string `json:"title"`
	Notes              string `json:"notes"`
	ExerciseTemplateID string `json:"exercise_template_id"`
	SupersetsID        int    `json:"supersets_id"` // Corrected field name and type
	Sets               []Set  `json:"sets"`
}

type Set struct {
	Index           int      `json:"index"`
	Type            string   `json:"type"`
	WeightKg        *float64 `json:"weight_kg"`
	Reps            *int     `json:"reps"`
	DistanceMeters  *int     `json:"distance_meters"`
	DurationSeconds *int     `json:"duration_seconds"`
	RPE             *float64 `json:"rpe"` // Changed from *int to *float64
	CustomMetric    *string  `json:"custom_metric"`
}

func (c *Client) FetchRecentWorkouts(page, pageSize int) ([]Workout, error) {
	query := url.Values{}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))

	resp, err := c.doGET("/v1/workouts", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result WorkoutResponse // Use the defined WorkoutResponse struct
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Workouts, err
}

// FetchWorkoutsByDateRange fetches workouts within a specific date range, with pagination.
// The Hevy API might use 'start_time' and 'end_time' for filtering.
func (c *Client) FetchWorkoutsByDateRange(page, pageSize int, startTime, endTime time.Time) ([]Workout, error) {
	query := url.Values{}
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	// Assuming the Hevy API accepts these parameters for date filtering
	if !startTime.IsZero() {
		query.Set("start_time", startTime.Format(time.RFC3339))
	}
	if !endTime.IsZero() {
		query.Set("end_time", endTime.Format(time.RFC3339))
	}

	resp, err := c.doGET("/v1/workouts", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var result WorkoutResponse // Use the defined WorkoutResponse struct
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Workouts, err
}

func (c *Client) FetchWorkoutByID(id string) (*Workout, error) {
	resp, err := c.doGET(fmt.Sprintf("/v1/workouts/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var workout Workout
	err = json.NewDecoder(resp.Body).Decode(&workout)
	return &workout, err
}

type Interface interface {
	FetchRecentWorkouts(page, pageSize int) ([]Workout, error)
	FetchWorkoutsByDateRange(page, pageSize int, startTime, endTime time.Time) ([]Workout, error) // Add to interface
	FetchWorkoutByID(id string) (*Workout, error)
}
