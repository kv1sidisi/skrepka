.PHONY: migrate-create migrate-up

DB_USER ?= skrepka_user
DB_PASSWORD ?= Murakava3213
DB_NAME ?= skrepka_db
DB_HOST ?= localhost
DB_PORT ?= 5432

DATABASE_URL=postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

migrate-up:
	export DATABASE_URL=$(DATABASE_URL); migrate -database "$$DATABASE_URL" -path ./migrations up

