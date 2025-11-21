.PHONY: build run test lint docker-up docker-down migrate-up migrate-down swagger load-test

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
	go test -v -race -coverprofile=coverage.out -coverpkg=./internal/... ./internal/test 2>&1 | \
	grep -E "(RUN|PASS|FAIL|coverage:)" | \
	sed 's/.*=== RUN   /Running: /' | \
	sed 's/--- PASS: /PASS: /' | \
	sed 's/--- FAIL: /FAIL: /' | \
	awk '/coverage:/ {print "Total coverage: " $$0; next} {print}'; \
	test $${PIPESTATUS[0]} -eq 0

lint:
	@if ! command -v golangci-lint > /dev/null 2>&1; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; fi
	@$$(command -v golangci-lint 2>/dev/null || echo $$(go env GOPATH)/bin/golangci-lint) run

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

migrate-up:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; docker run --rm -v $$(pwd)/migrations:/migrations --network host migrate/migrate:latest -path=/migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@$${DB_HOST:-localhost}:$${DB_PORT:-5435}/$${DB_NAME:-pr_reviewer}?sslmode=$${DB_SSLMODE:-disable}" up

migrate-down:
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; docker run --rm -v $$(pwd)/migrations:/migrations --network host migrate/migrate:latest -path=/migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@$${DB_HOST:-localhost}:$${DB_PORT:-5435}/$${DB_NAME:-pr_reviewer}?sslmode=$${DB_SSLMODE:-disable}" down

swagger:
	@if ! command -v swag > /dev/null 2>&1; then go install github.com/swaggo/swag/cmd/swag@latest; fi
	@$$(command -v swag 2>/dev/null || echo $$(go env GOPATH)/bin/swag) init -g cmd/server/main.go -o docs

load-test:
	@if ! command -v k6 > /dev/null 2>&1; then \
		echo "Installing k6..."; \
		if command -v dnf > /dev/null 2>&1; then \
			sudo dnf install -y https://github.com/grafana/k6/releases/download/v0.49.0/k6-v0.49.0-linux-amd64.rpm; \
		elif command -v apt > /dev/null 2>&1; then \
			sudo gpg -k || true; \
			sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69 || true; \
			echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list > /dev/null || true; \
			sudo apt-get update && sudo apt-get install -y k6; \
		else \
			echo "Please install k6 manually from https://k6.io/docs/getting-started/installation/"; \
			exit 1; \
		fi; \
	fi
	@k6 run k6/load_test.js
