postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root golive_cms

dropdb:
	docker exec -it postgres12 dropdb --username=root golive_cms

new_migrate: ## copy this and add you own name
	migrate create -ext sql -dir db/migration -seq migration_name

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/golive_cms?sslmode=disable" --verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/golive_cms?sslmode=disable" --verbose down

sqlc:
	sqlc generate

server:
	go run main.go

test:
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/go-live-cms/go-live-cms/db/sqlc Store

# Development commands
devbuild:
	docker compose -f compose.dev.yaml up --build

dev:
	docker compose -f compose.dev.yaml up

devdown:
	docker compose -f compose.dev.yaml down

devlogs:
	docker compose -f compose.dev.yaml logs -f

devlogs-api:
	docker compose -f compose.dev.yaml logs -f api

devlogs-web:
	docker compose -f compose.dev.yaml logs -f web

devrebuild:
	docker compose -f compose.dev.yaml up --build --force-recreate

# Production commands
prod:
	docker compose -f compose.yaml up --build

proddown:
	docker compose -f compose.yaml down

prodlogs:
	docker compose -f compose.yaml logs -f

.PHONY: createdb dropdb postgres migrateup migratedown sqlc test mock dev devdown devlogs devlogs-api devlogs-web devrebuild prod proddown prodlogs