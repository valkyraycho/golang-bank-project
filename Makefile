migrateup:
	migrate -path db/migration -database postgres://postgres:secret@localhost:5432/bank_project?sslmode=disable up

migrateup1:
	migrate -path db/migration -database postgres://postgres:secret@localhost:5432/bank_project?sslmode=disable up 1

migratedown:
	migrate -path db/migration -database postgres://postgres:secret@localhost:5432/bank_project?sslmode=disable down

migratedown1:
	migrate -path db/migration -database postgres://postgres:secret@localhost:5432/bank_project?sslmode=disable down 1

test:
	go test -v -cover ./...

.PHONY: migrateup migrateup1 migratedown migratedown1 test