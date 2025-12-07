package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/burakmert236/goodswipe-common/database"
	appErrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/models"
	"github.com/burakmert236/goodswipe-tournament-service/internal/repository"
	"github.com/google/uuid"
)

type TournamentService interface {
	CreateTournament(ctx context.Context, startsAt time.Time) (*models.Tournament, error)
	EnterTournament(ctx context.Context, userId, tournamentId string) error
	SubmitScore(ctx context.Context, userId string, score int) error
}

type tournamentService struct {
	tournamentRepo  repository.TournamentRepository
	participantRepo repository.ParticipationRepository
	groupRepo       repository.GroupRepository
	transactionRepo database.TransactionRepository
}

func NewTournamentService(
	tournamentRepo repository.TournamentRepository,
	participantRepo repository.ParticipationRepository,
	groupRepo repository.GroupRepository,
) TournamentService {
	return &tournamentService{
		tournamentRepo:  tournamentRepo,
		participantRepo: participantRepo,
		groupRepo:       groupRepo,
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

func (s *tournamentService) EnterTournament(
	ctx context.Context,
	userId, tournamentId string,
) error {
	group, err := s.groupRepo.FindAvailableGroup(ctx, tournamentId)

	if err != nil {
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			if appErr.Code == appErrors.ErrCodeNotFound {
				group = &models.Group{}
				s.setDefaultValueForGroup(group)
				err = s.groupRepo.CreateGroup(ctx, group)
			}
		}

		return fmt.Errorf("unexpected error: %w", err)
	}

	participation := &models.Participation{
		UserId:       userId,
		TournamentId: tournamentId,
		GroupId:      group.GroupId,
	}
	s.setDefaultValuesForParticipation(participation)
	putParticipationTransaction, err := s.participantRepo.GetTransactionForAddingParticipation(ctx, participation)
	if err != nil {
		return err
	}

	updateGroupTransaction := s.groupRepo.GetTransactionForAddingParticipant(ctx, group.GroupId, tournamentId)

	transactionBuilder := database.NewTransactionBuilder()
	transactionBuilder.AddPut(putParticipationTransaction)
	transactionBuilder.AddUpdate(updateGroupTransaction)

	return s.transactionRepo.Execute(ctx, transactionBuilder)
}

func (s *tournamentService) SubmitScore(
	ctx context.Context,
	userId string,
	score int,
) error {
	return nil
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
