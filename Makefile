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

protoc:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb \
    --grpc-gateway_opt paths=source_relative \
    proto/*.proto

mock:
	mockgen -destination db/mock/store.go github.com/valkyraycho/bank_project/db/sqlc Store

.PHONY: migrateup migrateup1 migratedown migratedown1 test protoc mock