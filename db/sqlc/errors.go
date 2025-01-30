package db

import "github.com/jackc/pgx/v5/pgconn"

const (
	UniqueViolation = "23505"
)

var ErrUniqueViolation = &pgconn.PgError{
	Code: UniqueViolation,
}
