package database

import "fmt"

type ErrorCode int32

const (
	INTERNAL_ERROR         ErrorCode = 0
	NAMESPACE_NOT_FOUND    ErrorCode = 1
	ID_NOT_FOUND           ErrorCode = 2
	UNABLE_TO_CREATE_TABLE ErrorCode = 3
	FILESYSTEM_ERROR       ErrorCode = 4
)

type DbError struct {
	ErrorCode ErrorCode
	Message   string
}

func (r *DbError) Error() string {
	return fmt.Sprintf("%v (error_code: %v)", r.Message, r.ErrorCode)
}
