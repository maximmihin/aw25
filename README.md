```makefile
# Makefile

up:
	docker compose up

down:
	docker compose down

test:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep "total:" | awk '{print $3}'

```