package service

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/burakmert236/goodswipe-common/database"
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	usererrors "github.com/burakmert236/goodswipe-user-service/internal/errors"
	"github.com/burakmert236/goodswipe-user-service/internal/events"
	"github.com/burakmert236/goodswipe-user-service/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, displayName string) (*models.User, *apperrors.AppError)
	GetById(ctx context.Context, userId string) (*models.User, *apperrors.AppError)
	UpdateProgress(ctx context.Context, userId string, levelIncrease int) *apperrors.AppError
	CollectTournamentReward(ctx context.Context, userId, tournamentId string, coin int) *apperrors.AppError

	// Reservation methods
	ReserveCoins(ctx context.Context, userId string, amount int, reservationId string) *apperrors.AppError
	ConfirmReservation(ctx context.Context, reservationId string) *apperrors.AppError
	RollbackReservation(ctx context.Context, reservationId string) *apperrors.AppError
}

type userService struct {
	userRepo              repository.UserRepository
	reservationRepo       repository.ReservationRepository
	rewardClaimRepository repository.RewardClaimRepository
	transactionRepo       database.TransactionRepository
	publisher             *events.EventPublisher
	logger                *logger.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	reservationRepo repository.ReservationRepository,
	rewardClaimRepository repository.RewardClaimRepository,
	transactionRepo database.TransactionRepository,
	publisher *events.EventPublisher,
	logger *logger.Logger,
) UserService {
	return &userService{
		userRepo:              userRepo,
		reservationRepo:       reservationRepo,
		rewardClaimRepository: rewardClaimRepository,
		transactionRepo:       transactionRepo,
		publisher:             publisher,
		logger:                logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, displayName string) (*models.User, *apperrors.AppError) {
	if displayName == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "display name is required")
	}

	user := s.getDefaultUser()
	user.DisplayName = displayName

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("User created: %s", user.UserId)

	err := s.publisher.PublishUserCreated(ctx, user.UserId, user.DisplayName)

	return user, err
}

func (s *userService) GetById(ctx context.Context, userId string) (*models.User, *apperrors.AppError) {
	user, err := s.userRepo.GetById(ctx, userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateProgress(ctx context.Context, userId string, levelIncrease int) *apperrors.AppError {
	reward := levelIncrease * s.getCoinRewardPerLevelUpgrade()
	newLevel, err := s.userRepo.UpdateLevelProgress(ctx, userId, levelIncrease, reward)
	if err != nil {
		return err
	}

	return s.publisher.PublishUserLevelUp(ctx, userId, levelIncrease, newLevel)
}

func (s *userService) CollectTournamentReward(
	ctx context.Context,
	userId, tournamentId string,
	coin int,
) *apperrors.AppError {
	rewardClaim, err := s.rewardClaimRepository.GetByIdempotency(ctx, userId, tournamentId)
	if err != nil {
		return err
	}

	if rewardClaim != nil {
		return nil
	}

	addCoinErr := s.userRepo.AddCoin(ctx, userId, coin)
	if addCoinErr != nil {
		if err := s.rewardClaimRepository.Delete(ctx, userId, tournamentId); err != nil {
			return err
		}
		return addCoinErr
	}

	if err = s.rewardClaimRepository.Create(ctx, userId, tournamentId); err != nil {
		return err
	}

	return nil
}

// Reservation methods

func (s *userService) ReserveCoins(ctx context.Context, userId string, amount int, reservationId string) *apperrors.AppError {
	existing, err := s.reservationRepo.GetById(ctx, reservationId)
	if err == nil && existing != nil {
		if existing.Status == models.ReservationStatusConfirmed || existing.Status == models.ReservationStatusReserved {
			return nil
		}
	}

	reservation := s.getDefaultReservation(userId, reservationId, amount)
	reservationItemPutTransaction, err := s.reservationRepo.GetCreateTransaction(ctx, &reservation)
	if err != nil {
		return err
	}

	userCoinDeductionTransaction := s.userRepo.GetCoinDeductionTransaction(ctx, userId, amount)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddUpdate(userCoinDeductionTransaction)
	transactionBuilder.AddPut(reservationItemPutTransaction)

	transactionErr := s.transactionRepo.Execute(ctx, transactionBuilder)

	if transactionErr != nil {
		var txErr *types.TransactionCanceledException
		if errors.As(transactionErr, &txErr) {
			for _, reason := range txErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return usererrors.WrapCoinReservationError(txErr)
				}
			}
		}
		return transactionErr
	}

	return nil
}

func (s *userService) ConfirmReservation(ctx context.Context, reservationId string) *apperrors.AppError {
	return s.reservationRepo.UpdateStatus(ctx, reservationId, models.ReservationStatusConfirmed)
}

func (s *userService) RollbackReservation(ctx context.Context, reservationId string) *apperrors.AppError {
	reservation, err := s.reservationRepo.GetById(ctx, reservationId)
	if err != nil {
		return err
	}

	if reservation.Status == models.ReservationStatusRolledBack {
		return nil
	}

	if reservation.Status != models.ReservationStatusReserved {
		return usererrors.CoinReservationRollbackError()
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
		UserId: uuid.New().String(),
		Level:  1,
		Coin:   1000,
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
