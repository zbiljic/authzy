package database

import "errors"

var (
	// ErrNotFound will throw if the item does not exists.
	ErrNotFound = errors.New("item not found")

	// ErrAlreadyExists is used when an item with same key already exists.
	ErrAlreadyExists = errors.New("item already exists")
)
