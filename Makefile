.DEFAULT_GOAL := api

dc_run = docker-compose run --name api --rm api

api:
	$(dc_run_api) go run cmd/api/main.go

up:
	docker network create traefik || true
	docker-compose up -d traefik
	docker-compose up -d redis
