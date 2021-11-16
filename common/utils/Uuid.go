package utils

import "github.com/google/uuid"

// Creates a new uuid as a string.
func NewUuid() string {
	return uuid.New().String()
}
