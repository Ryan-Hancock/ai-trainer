run-service:
    bash -c "export $(grep -v '^#' .env | xargs) && go run cmd/ai-trainer/main.go"

run-terminal:
    bash -c "export $(grep -v '^#' .env | xargs) && go run cmd/terminal/main.go"

migrate-db:
    # Running the main ai-trainer service will trigger all necessary database migrations
    # for both plan and history modules upon startup.
    bash -c "export $(grep -v '^#' .env | xargs) && go run cmd/ai-trainer/main.go"
