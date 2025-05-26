# AI Trainer Assistant

A personal trainer assistant that runs in the background and integrates with Hevy to track and analyze your workouts.

## Overview

This application serves as your personal AI trainer assistant. It runs in the background and periodically performs tasks such as:

- Generating workout plans based on your goals
- Analyzing your recent workouts from Hevy
- Providing nutrition tips
- Creating progress reports

## Setup

1. Clone this repository
2. Create a `.env` file in the root directory with the following content:
   ```
   OPENKEY=your_openai_api_key
   HEVYKEY=your_hevy_api_key
   ```
3. Replace the placeholder values with your actual API keys:
   - Get an OpenAI API key from [OpenAI](https://platform.openai.com/)
   - Get a Hevy API key from your Hevy account

## Running the Application

### Background Service

To run the application as a background service:

```bash
go run cmd/ai-trainer/main.go
```

This will start the AI Trainer Assistant in the background. It will run various tasks at scheduled intervals:
- Generate workout plans daily
- Analyze recent workouts twice daily
- Provide nutrition tips every 2 days
- Generate progress reports weekly

Press Ctrl+C to stop the service.

### Terminal Interface

For interactive use, you can run the terminal interface:

```bash
go run cmd/terminal/main.go
```

This provides a menu-driven interface where you can:
- Generate workout plans
- Check in
- Get nutrition tips

## Project Structure

- `cmd/` - Application entry points
  - `ai-trainer/` - Background service
  - `terminal/` - Interactive terminal interface
- `internal/` - Internal packages
  - `adapter/` - External service adapters
    - `hevy/` - Hevy API client
    - `llm/` - OpenAI client
  - `core/` - Core business logic
    - `plan/` - Workout planning
  - `service/` - Service implementations
    - `background/` - Background service
  - `utilities/` - Utility functions

## Development

For development and testing, you can modify the interval in `cmd/ai-trainer/main.go` to a shorter duration:

```go
bgService := background.NewService(background.Config{
    PlanService: planService,
    Interval:    1 * time.Minute, // Change to a shorter interval for testing
})
```

## Future Enhancements

- Log previous workouts from Hevy
- Add more sophisticated workout analysis
- Implement notifications
- Create a web or mobile interface
- Add more personalized recommendations
