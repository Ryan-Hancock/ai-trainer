run-service:
    set -a
    source .env
    go run cmd/ai-trainer/main.go