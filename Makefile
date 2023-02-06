#!make
include .env

deploy: lint docker-compose.up migrate.up
	@echo "----- deploy -----"

DB_CONNECTION="host=$(DBHOST) port=$(DBPORT) user=$(DBUSERNAME) password=$(DBPASSWORD) dbname=$(DBNAME) sslmode=$(DBSSLMODE)"
MIGRATIONS_FOLDER="pkg/database/schema"

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


lint: 
	@echo "----- lint programm -----"
	@golangci-lint run