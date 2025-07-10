run:
	@if ! docker compose ps | grep -q "Up"; then \
		echo "Starting Docker containers..."; \
		docker compose up -d; \
	else \
		echo "Docker containers already running."; \
	fi
	air

test:
	go test -v ./...