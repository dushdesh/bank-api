postgres:
	docker run --name bank-api-pg -e POSTGRES_PASSWORD=bank-api-db -p 5432:5432 -d postgres

mongo:
	docker run --name bank-api-mongo -p 27017:27017 -d mongo

db:	postgres mongo

pgstart:
	docker start bank-api-pg

mongostart:
	docker start bank-api-mongo

dbstart: pgstart mongostart

createdb:
	docker exec -it bank-api-pg createdb --username=postgres --owner=postgres bank-api

dropdb:
	docker exec -it bank-api-pg dropdb --username=postgres bank-api

migrateup:
	./bin/migrate -path db/migration/ -database "postgresql://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable" -verbose up

migratedown:
	./bin/migrate -path db/migration/ -database "postgresql://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable" -verbose down

sqlc:
	./bin/sqlc generate

build:
	@go build -o bin/bank

run: build
	@./bin/bank

test:
	@go test -v -cover ./...

.PHONY: postgres mongo db pgstart mongostart dbstart createdb dropdb migrateup migratedown sqlc build run test
