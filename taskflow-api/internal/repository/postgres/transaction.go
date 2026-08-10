package postgres

import (
	"context"
	"database/sql"
)

// WithTransaction begins a transaction, runs fn with it, and commits if fn
// succeeds. err must be a named return: the deferred closure below assigns
// to it (e.g. err = tx.Commit()) after fn has already returned, and only a
// named return value lets that assignment reach the caller.
func WithTransaction(ctx context.Context, db *DB, fn func(tx *sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
