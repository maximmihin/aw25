```makefile
# Makefile

up:
	docker compose up

down:
	docker compose down

test: # need docker for testcontainers https://golang.testcontainers.org/features/configuration/#docker-host-detection
	go test ./...

```