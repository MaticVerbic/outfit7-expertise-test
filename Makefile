dc_run = docker-compose run --name api --rm api

api:
	$(dc_run) go run cmd/api/main.go

test:
	$(dc_run) go test -v ./...

integration:
	$(dc_run) go test -v --tags=integration ./...

warmup:
	$(dc_run) go run cmd/cache/main.go

up:
	docker network create traefik || true
	docker-compose up -d traefik
	docker-compose up -d redis
