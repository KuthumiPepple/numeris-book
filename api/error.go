package api

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const ForeignKeyViolation = "23503"

var ErrRecordNotFound = pgx.ErrNoRows
var ErrForeignKeyViolation = &pgconn.PgError{Code: ForeignKeyViolation}

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
