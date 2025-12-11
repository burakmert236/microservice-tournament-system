package errors

import apperrors "github.com/burakmert236/goodswipe-common/errors"

func WrapCoinReservationError(err error) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeForbidden, "insufficient coin or reservation already exists")
}

func CoinReservationRollbackError() *apperrors.AppError {
	return apperrors.New(apperrors.CodeInternalServer, "reservation cannot be rolled back")
}
