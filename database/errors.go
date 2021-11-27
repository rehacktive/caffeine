package database

import "fmt"

type ErrorCode int32

const (
	INTERNAL_ERROR = 0;
	NAMESPACE_NOT_FOUND = 1;
	ID_NOT_FOUND        = 2;
	UNABLE_TO_CREATE_TABLE = 3;
	FILESYSTEM_ERROR = 4;
)

type DbError struct {
	ErrorCode int32
	Message string
}

func (r *DbError) Error() string {
	return fmt.Sprintf("%v (error_code: %v)", r.Message, r.ErrorCode)
}
