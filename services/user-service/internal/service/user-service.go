package service

import (
	"context"
	"fmt"

	"github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-user-service/internal/events"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, displayName string) (*models.User, error)
	UpdateProgress(ctx context.Context, userId string, levelIncrease int) error
}

type userService struct {
	userRepo  repository.UserRepository
	publisher *events.EventPublisher
	logger    *logger.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	publisher *events.EventPublisher,
	logger *logger.Logger,
) UserService {
	return &userService{
		userRepo:  userRepo,
		publisher: publisher,
		logger:    logger,
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

	s.logger.Info("User created: %s", user.UserId)

	return user, nil
}

func (s *userService) UpdateProgress(ctx context.Context, userId string, levelIncrease int) error {
	reward := levelIncrease * s.getCoinRewardPerLevelUpgrade()
	newLevel, err := s.userRepo.UpdateLevelProgress(ctx, userId, levelIncrease, reward)

	if err != nil {
		s.logger.Error("failed to update level progress: %s", err.Error())
		return err
	}

	return s.publisher.PublishUserLevelUp(ctx, userId, levelIncrease, newLevel)
}

func (s *userService) getCoinRewardPerLevelUpgrade() int {
	return 100
}

func (s *userService) setDefaultValueForUser(user *models.User) {
	user.Level = 1
	user.Coin = 1000
	user.TotalScore = 0
}
