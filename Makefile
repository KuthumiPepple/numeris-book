DB_URL=postgresql://root:secret@localhost:5432/numerisbookdb?sslmode=disable

postgres:
	docker compose up

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

db_start:
	docker compose start

db_stop:
	docker compose stop

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres new_migration migrateup migratedown db_start db_stop test server