package database

import "database/sql"

func Ptr[T any](n sql.Null[T]) *T {
	if n.Valid {
		return &n.V
	}
	return nil
}
