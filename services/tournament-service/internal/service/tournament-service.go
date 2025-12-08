package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/burakmert236/goodswipe-common/database"
	appErrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/google/uuid"
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
	logger          *logger.Logger
}

func NewTournamentService(
	tournamentRepo repository.TournamentRepository,
	participantRepo repository.ParticipationRepository,
	groupRepo repository.GroupRepository,
	transactionRepo database.TransactionRepository,
	logger *logger.Logger,
) TournamentService {
	return &tournamentService{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		groupRepo:       groupRepo,
		transactionRepo: transactionRepo,
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
		return nil, err
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
		return nil, err
	}

	return tournament, nil
}

func (s *tournamentService) EnterTournament(
	ctx context.Context,
	userId string,
) error {
	tournament, err := s.tournamentRepo.GetActiveTournament(ctx)
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf("current tournament is fetched: %s", tournament.TournamentId))

	if tournament.LastAllowedParticipationDate.Compare(time.Now()) > 0 {
		return fmt.Errorf("tournament last participation date is over:  %s",
			tournament.LastAllowedParticipationDate.Format(time.RFC3339))
	}

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
					return createGroupErr
				}
			}
		} else {
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	participation := &models.Participation{
		UserId:       userId,
		TournamentId: tournament.TournamentId,
		GroupId:      group.GroupId,
	}
	s.setDefaultValuesForParticipation(participation)
	putParticipationTransaction, err := s.participantRepo.GetTransactionForAddingParticipation(ctx, participation)
	if err != nil {
		return err
	}

	updateGroupTransaction := s.groupRepo.GetTransactionForAddingParticipant(ctx, group.GroupId, tournament.TournamentId)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddPut(putParticipationTransaction)
	transactionBuilder.AddUpdate(updateGroupTransaction)

	return s.transactionRepo.Execute(ctx, transactionBuilder)
}

func (s *tournamentService) UpdateParticipationScore(
	ctx context.Context,
	userId string,
	levelIncrease int,
) error {
	tournament, err := s.tournamentRepo.GetActiveTournament(ctx)
	if err != nil {
		return err
	}

	scoreReward := s.getLevelUpdateScoreReward(tournament, levelIncrease)
	return s.participantRepo.UpdateParticipationScore(ctx, userId, tournament.TournamentId, scoreReward)
}

func (s *tournamentService) setDefaultValuesForTournament(tournament *models.Tournament) {
	tournament.EndsAt = tournament.StartsAt.Add(24 * time.Hour).Add(-1 * time.Minute)
	tournament.LastAllowedParticipationDate = tournament.StartsAt.Add(12 * time.Hour)
	tournament.GroupSize = s.getDefaultGroupSize()
	tournament.ScoreRewardPerLevelUpgrade = 1
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
