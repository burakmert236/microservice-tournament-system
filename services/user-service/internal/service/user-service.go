package service

import (
	"context"
	"fmt"

	"github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, displayName string) (*models.User, error)
	UpdateProgress(ctx context.Context, userId string, levelIncrease int) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, displayName string) (*models.User, error) {
	if displayName == "" {
		return nil, errors.NewAppError(
			errors.ErrCodeInvalidInput,
			"invalid display name",
			nil,
		)
	}

	user := &models.User{
		UserId:      uuid.New().String(),
		DisplayName: displayName,
	}
	s.setDefaultValueForUser(user)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateProgress(ctx context.Context, userId string, levelIncrease int) error {
	reward := levelIncrease * s.getCoinRewardPerLevelUpgrade()
	return s.userRepo.UpdateLevelProgress(ctx, userId, levelIncrease, reward)
}

func (s *userService) getCoinRewardPerLevelUpgrade() int {
	return 100
}

func (s *userService) setDefaultValueForUser(user *models.User) {
	user.Level = 1
	user.Coin = 1000
	user.TotalScore = 0
}
