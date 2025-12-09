package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/burakmert236/goodswipe-common/database"
	appErrors "github.com/burakmert236/goodswipe-common/errors"
	protogrpc "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type TournamentService interface {
	CreateTournament(ctx context.Context, startsAt time.Time) (*models.Tournament, error)
	CreateCurrentTournament(ctx context.Context) (*models.Tournament, error)
	EnterTournament(ctx context.Context, userId string) error
	UpdateParticipationScore(ctx context.Context, userId string, levelIncrease int) error
}

type tournamentService struct {
	tournamentRepo  repository.TournamentRepository
	participantRepo repository.ParticipationRepository
	groupRepo       repository.GroupRepository
	transactionRepo database.TransactionRepository
	userClient      protogrpc.UserServiceClient
	logger          *logger.Logger
}

func NewTournamentService(
	tournamentRepo repository.TournamentRepository,
	participantRepo repository.ParticipationRepository,
	groupRepo repository.GroupRepository,
	transactionRepo database.TransactionRepository,
	userClient protogrpc.UserServiceClient,
	logger *logger.Logger,
) TournamentService {
	return &tournamentService{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		groupRepo:       groupRepo,
		transactionRepo: transactionRepo,
		userClient:      userClient,
		logger:          logger,
	}
}

func (s *tournamentService) CreateTournament(ctx context.Context, startsAt time.Time) (*models.Tournament, error) {
	tournament := &models.Tournament{
		TournamentId: uuid.New().String(),
		StartsAt:     startsAt,
	}
	s.setDefaultValuesForTournament(tournament)

	if err := s.tournamentRepo.Create(ctx, tournament); err != nil {
		return nil, fmt.Errorf("failed to create tournament %w", err)
	}

	return tournament, nil
}

func (s *tournamentService) CreateCurrentTournament(ctx context.Context) (*models.Tournament, error) {
	currentTournament, _ := s.tournamentRepo.GetActiveTournament(ctx)
	if currentTournament != nil {
		return currentTournament, nil
	}

	now := time.Now().UTC()
	startsAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tournament := &models.Tournament{
		TournamentId: uuid.New().String(),
		StartsAt:     startsAt,
	}
	s.setDefaultValuesForTournament(tournament)

	if err := s.tournamentRepo.Create(ctx, tournament); err != nil {
		return nil, fmt.Errorf("failed to create tournament %w", err)
	}

	return tournament, nil
}

func (s *tournamentService) EnterTournament(
	ctx context.Context,
	userId string,
) error {
	reservationId := uuid.New().String()

	// Get active tournament
	tournament, err := s.tournamentRepo.GetActiveTournament(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tournament %w", err)
	}
	s.logger.Info(fmt.Sprintf("current tournament is fetched: %s", tournament.TournamentId))

	// Check existing particpation
	existingParticipation, err := s.participantRepo.GetByUserAndTournament(ctx, userId, tournament.TournamentId)
	if err != nil {
		return fmt.Errorf("failed to get existing participation %w", err)
	}
	if existingParticipation != nil {
		return nil
	}

	// Handle validation and reservation
	if err := s.handleBeforeTournamentEntryOperations(ctx, userId, reservationId, tournament); err != nil {
		return err
	}

	// Get available group
	group, err := s.findOrCreateAvailableGroup(ctx, tournament)
	if err != nil {
		return fmt.Errorf("failed to get available group %w", err)
	}

	// Build transaction for participation
	participation := &models.Participation{
		UserId:       userId,
		TournamentId: tournament.TournamentId,
		GroupId:      group.GroupId,
	}
	s.setDefaultValuesForParticipation(participation)
	putParticipationTransaction, err := s.participantRepo.GetTransactionForAddingParticipation(ctx, participation)
	if err != nil {
		return fmt.Errorf("failed to get transaction for adding new participation %w", err)
	}

	updateGroupTransaction := s.groupRepo.GetTransactionForAddingParticipant(ctx, group.GroupId, tournament.TournamentId)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddPut(putParticipationTransaction)
	transactionBuilder.AddUpdate(updateGroupTransaction)

	transactionErr := s.transactionRepo.Execute(ctx, transactionBuilder)
	return s.handleAfterTournamentEntryOperations(ctx, transactionErr, reservationId)
}

func (s *tournamentService) UpdateParticipationScore(
	ctx context.Context,
	userId string,
	levelIncrease int,
) error {
	tournament, err := s.tournamentRepo.GetActiveTournament(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tournament %w", err)
	}

	scoreReward := s.getLevelUpdateScoreReward(tournament, levelIncrease)
	return s.participantRepo.UpdateParticipationScore(ctx, userId, tournament.TournamentId, scoreReward)
}

// Private methods

func (s *tournamentService) setDefaultValuesForTournament(tournament *models.Tournament) {
	tournament.EndsAt = tournament.StartsAt.Add(24 * time.Hour).Add(-1 * time.Minute)
	tournament.LastAllowedParticipationDate = tournament.StartsAt.Add(12 * time.Hour)
	tournament.UserLevelLimit = 10
	tournament.GroupSize = s.getDefaultGroupSize()
	tournament.ScoreRewardPerLevelUpgrade = 1
	tournament.EnteranceFee = 500
	tournament.RewardingMap = map[string]int{
		"1":    5000,
		"2":    3000,
		"3":    2000,
		"4-10": 1000,
	}
}

func (s *tournamentService) setDefaultValueForGroup(group *models.Group) {
	group.GroupSize = s.getDefaultGroupSize()
	group.IsFull = false
}

func (s *tournamentService) getDefaultGroupSize() int {
	return 35
}

func (s *tournamentService) setDefaultValuesForParticipation(participation *models.Participation) {
	participation.RewardsClaimed = false
	participation.Score = 0
}

func (s *tournamentService) getLevelUpdateScoreReward(
	tournament *models.Tournament,
	levelIncrease int,
) int {
	return levelIncrease * tournament.ScoreRewardPerLevelUpgrade
}

func (s *tournamentService) validateDate(tournament *models.Tournament) error {
	if tournament.LastAllowedParticipationDate.Compare(time.Now()) > 0 {
		return fmt.Errorf("tournament last participation date is over:  %s",
			tournament.LastAllowedParticipationDate.Format(time.RFC3339))
	}

	return nil
}

func (s *tournamentService) validateUserLevel(userLevel int, tournament *models.Tournament) error {
	if userLevel < tournament.UserLevelLimit {
		return fmt.Errorf("user level must be at least: %d", tournament.UserLevelLimit)
	}

	return nil
}

func (s *tournamentService) findOrCreateAvailableGroup(ctx context.Context, tournament *models.Tournament) (*models.Group, error) {
	group, err := s.groupRepo.FindAvailableGroup(ctx, tournament.TournamentId)

	if err != nil {
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == appErrors.ErrCodeNotFound {
				group = &models.Group{
					GroupId:      uuid.New().String(),
					TournamentId: tournament.TournamentId,
					GroupSize:    tournament.GroupSize,
				}
				s.setDefaultValueForGroup(group)
				if createGroupErr := s.groupRepo.CreateGroup(ctx, group); createGroupErr != nil {
					return nil, fmt.Errorf("failed to create group %w", createGroupErr)
				}
			}
		} else {
			return nil, fmt.Errorf("unexpected error: %w", err)
		}
	}

	return group, nil
}

func (s *tournamentService) handleBeforeTournamentEntryOperations(
	ctx context.Context,
	userId, reservationId string,
	tournament *models.Tournament,
) error {
	if err := s.validateDate(tournament); err != nil {
		return err
	}

	userResponse, err := s.userClient.GetById(ctx, &protogrpc.GetUserByIdRequest{
		UserId: userId,
	})

	if err := s.validateUserLevel(int(userResponse.Level), tournament); err != nil {
		return err
	}

	_, err = s.userClient.ReserveCoins(ctx, &protogrpc.ReserveCoinsRequest{
		UserId:        userId,
		Amount:        int64(tournament.EnteranceFee),
		ReservationId: reservationId,
	})

	if err != nil {
		st, _ := status.FromError(err)
		return fmt.Errorf("failed to reserve coins: %s", st.Message())
	}

	return nil
}

func (s *tournamentService) handleAfterTournamentEntryOperations(ctx context.Context, trsansactionErr error, reservationId string) error {
	if trsansactionErr != nil {
		s.logger.Warn("Failed to add participant, rolling back reservation %s", reservationId)

		_, rollbackErr := s.userClient.RollbackReservation(ctx, &protogrpc.RollbackReservationRequest{
			ReservationId: reservationId,
		})

		if rollbackErr != nil {
			s.logger.Warn("CRITICAL: Failed to rollback reservation %s: %v", reservationId, rollbackErr)
		}

		return fmt.Errorf("failed to join tournament: %w", trsansactionErr)
	}

	s.logger.Info("Confirming reservation %s", reservationId)
	_, err := s.userClient.ConfirmReservation(ctx, &protogrpc.ConfirmReservationRequest{
		ReservationId: reservationId,
	})

	if err != nil {
		s.logger.Warn("Failed to confirm reservation %s: %v", reservationId, err)
	}

	return err
}
