package database

import "database/sql"

func Ptr[T any](n sql.Null[T]) *T {
	if n.Valid {
		return &n.V
	}
	return nil
}

func Null[T any](ptr *T) sql.Null[T] {
	if ptr == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{
		V:     *ptr,
		Valid: true,
	}
}
