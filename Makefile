#!make
include .env
LINTER=golangci-lint

deploy: lint docker-compose.up migrate.up
	@echo "----- deploy -----"

DB_CONNECTION="host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USERNAME) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSL_MODE)"
MIGRATIONS_FOLDER="pkg/database/migrations"
SQLC_FOLDER="pkg/database"

docker-compose.up: 
	@echo "----- deploy by docker -----"
	@docker-compose up -d


docker-compose.down: 
	docker-compose down

migrate.up:
	@echo "----- running migrations up -----"
	@cd $(MIGRATIONS_FOLDER);\
	goose postgres ${DB_CONNECTION} up


migrate.down:
	cd $(MIGRATIONS_FOLDER);\
	goose postgres ${DB_CONNECTION} down


migrate.create:
	cd $(MIGRATIONS_FOLDER);\
	goose create $(name) sql

sqlc.generate:
	cd $(SQLC_FOLDER);\
	sqlc generate

lint: 
	$(LINTER) run

docker.build:
	docker build -t dyleme/apod .

docker.push: docker.build
	docker push dyleme/apod
