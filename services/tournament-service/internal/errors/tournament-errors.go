package errors

import (
	"fmt"
	"time"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
)

func ClaimRewardError() *apperrors.AppError {
	return apperrors.New(apperrors.CodeForbidden,
		"there is no participation or reward is already claimed")
}

func TournamentDateError(date time.Time) *apperrors.AppError {
	return apperrors.New(apperrors.CodeForbidden,
		fmt.Sprintf("tournament last participation date is over: %s", date.Format(time.RFC3339)))
}

func UserLevelLimitError(limit int) *apperrors.AppError {
	return apperrors.New(apperrors.CodeForbidden,
		fmt.Sprintf("user level must be at least %d", limit))
}

func InvalidRankingError() *apperrors.AppError {
	return apperrors.New(apperrors.CodeInternalServer, "ranking must be a positive number")
}

func TournamentNotFinishedError() *apperrors.AppError {
	return apperrors.New(apperrors.CodeForbidden, "tournament is not finished yet")
}
