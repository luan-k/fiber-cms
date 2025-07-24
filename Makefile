postgres:
	 docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	 docker exec -it postgres12 createdb --username=root --owner=root fiber_cms

dropdb:
	 docker exec -it postgres12 dropdb --username=root fiber_cms

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/fiber_cms?sslmode=disable" --verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/fiber_cms?sslmode=disable" --verbose down

.PHONY: createdb dropdb postgres migrateup migratedown