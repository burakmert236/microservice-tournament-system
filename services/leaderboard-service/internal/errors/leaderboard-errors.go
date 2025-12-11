package errors

import apperrors "github.com/burakmert236/goodswipe-common/errors"

func UserNotExistsInAnyGroup() *apperrors.AppError {
	return apperrors.New(apperrors.CodeNotFound, "user doesn't exists in any group og this tournament")
}
