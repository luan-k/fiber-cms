postgres:
	 docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	 docker exec -it postgres12 createdb --username=root --owner=root golive_cms

dropdb:
	 docker exec -it postgres12 dropdb --username=root golive_cms

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

.PHONY: createdb dropdb postgres migrateup migratedown sqlc test mock