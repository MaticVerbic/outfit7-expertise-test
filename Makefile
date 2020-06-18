dc_run = docker-compose run --name api --rm api

api:
	$(dc_run) go run cmd/api/main.go

test:
	$(dc_run) go test -v ./...

up:
	docker network create traefik || true
	docker-compose up -d traefik
	docker-compose up -d redis
