#!/bin/bash
set -e


# Environment Variables for Testing
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable"
export JWT_ENABLED=true
export JWT_SECRET="default-secret-for-testing-only-at-least-32-chars"
export PROVIDER_BASE_URL="http://localhost:8080"
# Kill anything on port 8080
lsof -ti:8080 | xargs kill -9 || true
# Ensure we have a clean state if possible, or at least that previous runs didn't leave a zombie
pkill -f "make run" || true
pkill -f "go run ./cmd/api" || true


# Create test_db if not exists
echo "Creating test_db..."
docker compose exec postgres createdb -U postgres test_db || true

echo "Running migrations..."
make migrate-up

make run > provider.log 2>&1 &
PID=$!

echo "Provider PID: $PID"

# Wait for health check
echo "Waiting for provider to be ready..."
READY=false
for i in {1..30}; do
  if curl -s http://localhost:8080/healthz > /dev/null; then
    echo "Provider is ready!"
    READY=true
    break
  fi
  sleep 1
done

if [ "$READY" = false ]; then
    echo "Provider failed to start"
    cat provider.log
    kill $PID || true
    exit 1
fi

echo "Checking /healthz headers..."
curl -v http://localhost:8080/healthz > health_body.txt 2> health_headers.txt

echo "Running provider verification..."
# Run the test and capture exit code
set +e
make test-contract-provider
TEST_EXIT=$?
set -e

echo "Stopping provider..."
kill $PID || true

if [ $TEST_EXIT -eq 0 ]; then
    echo "Verification SUCCESS"
else
    echo "Verification FAILED"
    # Show logs if failure
    echo "--- Provider Logs ---"
    tail -n 50 provider.log
    echo "--- Health Check Headers ---"
    cat health_headers.txt
fi

exit $TEST_EXIT
