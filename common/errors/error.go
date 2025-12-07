package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

type ErrorCode string

const (
	ErrCodeNotFound             ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists        ErrorCode = "ALREADY_EXISTS"
	ErrCodeInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrCodeInternal             ErrorCode = "INTERNAL"
	ErrCodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrCodeTournamentFull       ErrorCode = "TOURNAMENT_FULL"
	ErrCodeAlreadyParticipating ErrorCode = "ALREADY_PARTICIPATING"
)

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Convert AppError to gRPC status
func (e *AppError) ToGRPCStatus() error {
	var grpcCode codes.Code

	switch e.Code {
	case ErrCodeNotFound:
		grpcCode = codes.NotFound
	case ErrCodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	case ErrCodeInvalidInput:
		grpcCode = codes.InvalidArgument
	case ErrCodeUnauthorized:
		grpcCode = codes.Unauthenticated
	case ErrCodeTournamentFull:
		grpcCode = codes.FailedPrecondition
	default:
		grpcCode = codes.Internal
	}

	return status.Error(grpcCode, e.Message)
}
