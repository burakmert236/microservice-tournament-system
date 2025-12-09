package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/burakmert236/goodswipe-common/database"
	commonerrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-user-service/internal/events"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, displayName string) (*models.User, error)
	UpdateProgress(ctx context.Context, userId string, levelIncrease int) error
	GetById(ctx context.Context, userId string) (*models.User, error)

	// Reservation methods
	ReserveCoins(ctx context.Context, userId string, amount int, reservationId string) error
	ConfirmReservation(ctx context.Context, reservationId string) error
	RollbackReservation(ctx context.Context, reservationId string) error
}

type userService struct {
	userRepo        repository.UserRepository
	reservationRepo repository.ReservationRepository
	transactionRepo database.TransactionRepository
	publisher       *events.EventPublisher
	logger          *logger.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	reservationRepo repository.ReservationRepository,
	transactionRepo database.TransactionRepository,
	publisher *events.EventPublisher,
	logger *logger.Logger,
) UserService {
	return &userService{
		userRepo:        userRepo,
		reservationRepo: reservationRepo,
		transactionRepo: transactionRepo,
		publisher:       publisher,
		logger:          logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, displayName string) (*models.User, error) {
	if displayName == "" {
		return nil, commonerrors.NewAppError(
			commonerrors.ErrCodeInvalidInput,
			"invalid display name",
			nil,
		)
	}

	user := s.getDefaultUser()
	user.DisplayName = displayName

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created: %s", user.UserId)

	err := s.publisher.PublishUserCreated(ctx, user.UserId)

	return user, err
}

func (s *userService) UpdateProgress(ctx context.Context, userId string, levelIncrease int) error {
	reward := levelIncrease * s.getCoinRewardPerLevelUpgrade()
	newLevel, err := s.userRepo.UpdateLevelProgress(ctx, userId, levelIncrease, reward)

	if err != nil {
		return fmt.Errorf("failed to update level progress: %w", err)
	}

	return s.publisher.PublishUserLevelUp(ctx, userId, levelIncrease, newLevel)
}

func (s *userService) GetById(ctx context.Context, userId string) (*models.User, error) {
	user, err := s.userRepo.GetById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %s, %w", userId, err)
	}

	return user, nil
}

// Rservation methods

func (s *userService) ReserveCoins(ctx context.Context, userId string, amount int, reservationId string) error {
	existing, err := s.reservationRepo.GetById(ctx, reservationId)
	if err == nil && existing != nil {
		if existing.Status == models.ReservationStatusConfirmed || existing.Status == models.ReservationStatusReserved {
			return nil
		}
	}

	reservation := s.getDefaultReservation(userId, reservationId, amount)
	reservationItemPutTransaction, err := s.reservationRepo.GetCreateTransaction(ctx, &reservation)
	if err != nil {
		return fmt.Errorf("failed to get reservation create transaction: %w", err)
	}

	userCoinDeductionTransaction := s.userRepo.GetCoinDeductionTransaction(ctx, userId, amount)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddUpdate(userCoinDeductionTransaction)
	transactionBuilder.AddPut(reservationItemPutTransaction)

	err = s.transactionRepo.Execute(ctx, transactionBuilder)

	if err != nil {
		var txErr *types.TransactionCanceledException
		if errors.As(err, &txErr) {
			for _, reason := range txErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return commonerrors.NewAppError(
						commonerrors.ErrCodeInvalidInput,
						"insufficient coins or reservation already exists",
						nil,
					)
				}
			}
		}
		return fmt.Errorf("failed to reserve coins: %w", err)
	}

	return nil
}

func (s *userService) ConfirmReservation(ctx context.Context, reservationId string) error {
	return s.reservationRepo.UpdateStatus(ctx, reservationId, models.ReservationStatusConfirmed)
}

func (s *userService) RollbackReservation(ctx context.Context, reservationId string) error {
	reservation, err := s.reservationRepo.GetById(ctx, reservationId)
	if err != nil {
		return fmt.Errorf("failed to get reservation by id: %s, %w", reservationId, err)
	}

	if reservation.Status == models.ReservationStatusRolledBack {
		return nil
	}

	if reservation.Status != models.ReservationStatusReserved {
		return commonerrors.NewAppError(
			commonerrors.ErrCodeInvalidInput,
			"reservation cannot be rolled back",
			nil,
		)
	}

	coinAdditionTransaction := s.userRepo.GetCoinAdditionTransaction(ctx, reservation.UserId, int(reservation.Amount))
	updateReservationStatusTransaction := s.reservationRepo.GetUpdateStatusTransaction(ctx, reservationId, models.ReservationStatusRolledBack)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddUpdate(coinAdditionTransaction)
	transactionBuilder.AddUpdate(updateReservationStatusTransaction)

	return s.transactionRepo.Execute(ctx, transactionBuilder)
}

// Private methods

func (s *userService) getCoinRewardPerLevelUpgrade() int {
	return 100
}

func (s *userService) getDefaultUser() *models.User {
	return &models.User{
		UserId:     uuid.New().String(),
		Level:      1,
		Coin:       1000,
		TotalScore: 0,
	}
}

func (s *userService) getDefaultReservation(userId, reservationId string, amount int) models.Reservation {
	now := time.Now().UTC()

	reservation := &models.Reservation{
		ReservationId: reservationId,
		UserId:        userId,
		Amount:        int64(amount),
		Status:        models.ReservationStatusReserved,
		Purpose:       "TOURNAMENT_ENTRY",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return *reservation
}
