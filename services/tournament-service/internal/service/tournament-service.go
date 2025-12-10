package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/burakmert236/goodswipe-common/database"
	appErrors "github.com/burakmert236/goodswipe-common/errors"
	protogrpc "github.com/burakmert236/goodswipe-common/generated/v1/grpc"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-tournament-service/internal/events/publisher"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type TournamentService interface {
	CreateTournament(ctx context.Context, startsAt time.Time) (*models.Tournament, error)
	CreateCurrentTournament(ctx context.Context) (*models.Tournament, error)
	EnterTournament(ctx context.Context, userId string) error
	UpdateParticipationScore(ctx context.Context, userId string, levelIncrease int) error
	ClaimReward(ctx context.Context, userId, tournamentId string) error
}

type tournamentService struct {
	tournamentRepo    repository.TournamentRepository
	participationRepo repository.ParticipationRepository
	groupRepo         repository.GroupRepository
	transactionRepo   database.TransactionRepository
	userClient        protogrpc.UserServiceClient
	leaderboardClient protogrpc.LeaderboardServiceClient
	eventPublisher    *publisher.EventPublisher
	logger            *logger.Logger
}

func NewTournamentService(
	tournamentRepo repository.TournamentRepository,
	participationRepo repository.ParticipationRepository,
	groupRepo repository.GroupRepository,
	transactionRepo database.TransactionRepository,
	userClient protogrpc.UserServiceClient,
	leaderboardClient protogrpc.LeaderboardServiceClient,
	eventPublisher *publisher.EventPublisher,
	logger *logger.Logger,
) TournamentService {
	return &tournamentService{
		tournamentRepo:    tournamentRepo,
		participationRepo: participationRepo,
		groupRepo:         groupRepo,
		transactionRepo:   transactionRepo,
		userClient:        userClient,
		leaderboardClient: leaderboardClient,
		eventPublisher:    eventPublisher,
		logger:            logger,
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

	// Check existing participation
	existingParticipation, err := s.participationRepo.GetByUserAndTournament(ctx, userId, tournament.TournamentId)
	if err != nil {
		return fmt.Errorf("failed to get existing participation %w", err)
	}
	if existingParticipation != nil {
		return nil
	}

	// Fetch user
	userResponse, err := s.userClient.GetById(ctx, &protogrpc.GetUserByIdRequest{
		UserId: userId,
	})
	if err != nil {
		return fmt.Errorf("failed to get user grpc call: %s", err.Error())
	}

	// Handle validation and reservation
	if err := s.handleBeforeTournamentEntryOperations(ctx, userId, reservationId, int(userResponse.Level), tournament); err != nil {
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
		EndsAt:       tournament.EndsAt,
		RewardingMap: tournament.RewardingMap,
	}
	s.setDefaultValuesForParticipation(participation)
	putParticipationTransaction, err := s.participationRepo.GetTransactionForAddingParticipation(ctx, participation)
	if err != nil {
		return fmt.Errorf("failed to get transaction for adding new participation %w", err)
	}

	updateGroupTransaction := s.groupRepo.GetTransactionForAddingParticipant(ctx, group.GroupId, tournament.TournamentId)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddPut(putParticipationTransaction)
	transactionBuilder.AddUpdate(updateGroupTransaction)

	transactionErr := s.transactionRepo.Execute(ctx, transactionBuilder)

	if err = s.handleAfterTournamentEntryOperations(ctx, transactionErr, reservationId); err != nil {
		return fmt.Errorf("failed to handle after tournament entry operations %w", err)
	}

	if err = s.eventPublisher.PublishTournamentEntered(ctx, userId, userResponse.DisplayName, group.GroupId, tournament.TournamentId); err != nil {
		return fmt.Errorf("failed to publish tournament entered event %w", err)
	}

	return nil
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
	participation, err := s.participationRepo.UpdateParticipationScore(ctx, userId, tournament.TournamentId, scoreReward)
	if err != nil {
		return fmt.Errorf("failed to update participation score %w", err)
	}

	if participation != nil {
		s.logger.Info("Participation score updated. %d", participation)

		s.eventPublisher.PublishTournamentParticipationScoreUpdated(
			ctx,
			userId,
			participation.GroupId,
			participation.TournamentId,
			participation.Score,
		)
	}

	return nil
}

func (s *tournamentService) ClaimReward(ctx context.Context, userId, tournamentId string) error {
	participation, err := s.participationRepo.UpdateRewardProcessing(ctx, userId, tournamentId)
	if err != nil {
		return fmt.Errorf("failed to get tournament participation for user %w", err)
	}
	if participation == nil {
		return fmt.Errorf("there is no participation or reward is already claimed")
	}

	reward, err := s.handleRewardClaim(ctx, userId, participation)
	if err != nil {
		if _, err := s.participationRepo.UpdateRewardUnclaimed(ctx, userId, tournamentId); err != nil {
			return fmt.Errorf("failed to update back to unclaimed: %w", err)
		}
		return err
	}

	if reward <= 0 {
		if _, err := s.participationRepo.UpdateRewardClaimed(ctx, userId, tournamentId); err != nil {
			return fmt.Errorf("failed to mark participation claimed: %w", err)
		}
		return nil
	}

	addCoinResponse, err := s.userClient.CollectTournamentReward(ctx, &protogrpc.CollectTournamentRewardRequest{
		UserId:       userId,
		TournamentId: tournamentId,
		Coin:         int32(reward),
	})
	if addCoinResponse == nil || err != nil {
		if _, err := s.participationRepo.UpdateRewardUnclaimed(ctx, userId, tournamentId); err != nil {
			return fmt.Errorf("failed to update back to unclaimed: %w", err)
		}
		return err
	}

	if _, err := s.participationRepo.UpdateRewardClaimed(ctx, userId, tournamentId); err != nil {
		return fmt.Errorf("failed to mark participation claimed: %w", err)
	}
	return nil
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
	group.ParticipantCount = 0
}

func (s *tournamentService) getDefaultGroupSize() int {
	return 35
}

func (s *tournamentService) setDefaultValuesForParticipation(participation *models.Participation) {
	participation.RewardClaimStatus = models.Unclaimed
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
	level int,
	tournament *models.Tournament,
) error {
	if err := s.validateDate(tournament); err != nil {
		return err
	}

	if err := s.validateUserLevel(level, tournament); err != nil {
		return err
	}

	_, err := s.userClient.ReserveCoins(ctx, &protogrpc.ReserveCoinsRequest{
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

func (s *tournamentService) calculateReward(ctx context.Context, ranking int, rewardingMap map[string]int) (int, error) {
	if ranking < 1 {
		return 0, fmt.Errorf("invalid ranking: must be positive")
	}

	rankStr := strconv.Itoa(ranking)
	if reward, exists := rewardingMap[rankStr]; exists {
		return reward, nil
	}

	for key, reward := range rewardingMap {
		if strings.Contains(key, "-") {
			parts := strings.Split(key, "-")
			if len(parts) != 2 {
				continue
			}

			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			if err1 != nil || err2 != nil {
				continue
			}

			if ranking >= start && ranking <= end {
				return reward, nil
			}
		}
	}

	return 0, nil
}

func (s *tournamentService) handleRewardClaim(
	ctx context.Context,
	userId string,
	participation *models.Participation,
) (int, error) {
	if participation.EndsAt.Compare(time.Now().UTC()) > 0 {
		return 0, fmt.Errorf("tournament is not finished yet")
	}

	rankingResponse, err := s.leaderboardClient.GetTournamentRank(ctx, &protogrpc.GetTournamentRankRequest{
		UserId:       userId,
		TournamentId: participation.TournamentId,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get leaderboard grpc call: %s", err.Error())
	}

	reward, err := s.calculateReward(ctx, int(rankingResponse.Rank), participation.RewardingMap)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate reward for user %w", err)
	}

	return reward, nil
}
