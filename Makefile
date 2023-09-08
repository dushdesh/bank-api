postgres:
	docker run --name bank-api-pg -e POSTGRES_PASSWORD=bank-api-db -p 5432:5432 -d postgres

mongo:
	docker run --name bank-api-mongo -p 27017:27017 -d mongo

db:	postgres mongo

pgstart:
	docker start bank-api-pg

pgconnect:
	psql postgres://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable

mongostart:
	docker start bank-api-mongo

dbstart: pgstart mongostart

createdb:
	docker exec -it bank-api-pg createdb --username=postgres --owner=postgres bank-api

dropdb:
	docker exec -it bank-api-pg dropdb --username=postgres bank-api

migratecreate:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration/ -database "postgresql://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration/ -database "postgresql://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration/ -database "postgresql://postgres:bank-api-db@localhost:5432/bank-api?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

build:
	@go build -o bin/bank

run: build
	@./bin/bank

test:
	@go test -v -cover ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go bank/db/sqlc Store

server:
	go run main.go

docker:
	sudo service docker start

portcheck:
	lsof -i $(port)

.PHONY: postgres mongo db pgstart pgconnect mongostart dbstart createdb dropdb migratecreate migrateup migratedown sqlc build run test server docker mock
