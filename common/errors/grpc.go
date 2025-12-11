package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err *AppError) error {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		code := mapErrorCodeToGRPC(appErr.Code)
		return status.Error(code, appErr.Message)
	}

	return status.Error(codes.Internal, err.Error())
}

func mapErrorCodeToGRPC(code string) codes.Code {
	switch code {
	case CodeNotFound:
		return codes.NotFound
	case CodeAlreadyExists:
		return codes.AlreadyExists
	case CodeInvalidInput:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden:
		return codes.PermissionDenied
	case CodeConflict:
		return codes.Aborted
	case CodeServiceUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func FromGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	code := mapGRPCToErrorCode(st.Code())
	return &AppError{
		Code:    code,
		Message: st.Message(),
		Err:     err,
	}
}

func mapGRPCToErrorCode(code codes.Code) string {
	switch code {
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeAlreadyExists
	case codes.InvalidArgument:
		return CodeInvalidInput
	case codes.Unauthenticated:
		return CodeUnauthorized
	case codes.PermissionDenied:
		return CodeForbidden
	case codes.Aborted:
		return CodeConflict
	case codes.Unavailable:
		return CodeServiceUnavailable
	case codes.FailedPrecondition:
		return CodeInvalidInput
	default:
		return CodeInternalServer
	}
}
