package service

import (
	"context"

	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/repository"
)

type LeaderboardService interface {
	// Write Operations
	AddGlobalUser(ctx context.Context, userId, displayName string) *apperrors.AppError
	AddUserToTournament(ctx context.Context, userId, displayName, groupId, tournamentId string) *apperrors.AppError
	UpdateTournamentScore(ctx context.Context, userId, tournamentId string, score int) *apperrors.AppError

	// Read Operations
	GetGlobalLeaderboard(ctx context.Context) ([]repository.LeaderboardEntry, *apperrors.AppError)
	GetTournamentLeaderboard(ctx context.Context, userId, tournamentId string) ([]repository.LeaderboardEntry, *apperrors.AppError)
	GetTournamentRank(ctx context.Context, userId, tournamentId string) (int, *apperrors.AppError)
}

type leaderboardService struct {
	leaderboardRepo repository.LeaderboardRepository
	logger          *logger.Logger
}

func NewLeaderboardService(
	leaderboardRepo repository.LeaderboardRepository,
	logger *logger.Logger,
) LeaderboardService {
	return &leaderboardService{
		leaderboardRepo: leaderboardRepo,
		logger:          logger,
	}
}

// Write Operations

func (s *leaderboardService) AddGlobalUser(ctx context.Context, userId, displayName string) *apperrors.AppError {
	s.logger.Info("Adding global user")

	if err := s.leaderboardRepo.AddGlobalUser(ctx, userId, displayName); err != nil {
		return err
	}

	s.logger.Info("Global user added")
	return nil
}

func (s *leaderboardService) AddUserToTournament(
	ctx context.Context,
	userId, displayName, groupId, tournamentId string,
) *apperrors.AppError {
	s.logger.Info("Adding tournament user")

	if err := s.leaderboardRepo.AddUserToTournament(ctx, userId, displayName, groupId, tournamentId); err != nil {
		return nil
	}

	s.logger.Info("Tournament user added")
	return nil
}

func (s *leaderboardService) UpdateTournamentScore(
	ctx context.Context,
	userId, tournamentId string,
	score int,
) *apperrors.AppError {
	s.logger.Info("Updating tournament score")

	if err := s.leaderboardRepo.UpdateTournamentScore(ctx, userId, tournamentId, score); err != nil {
		return err
	}

	s.logger.Info("Tournament score updated")
	return nil
}

// Read Operations

func (s *leaderboardService) GetGlobalLeaderboard(ctx context.Context) ([]repository.LeaderboardEntry, *apperrors.AppError) {
	s.logger.Info("Getting global leaderboard")

	entries, err := s.leaderboardRepo.GetGlobalLeaderboard(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Global leaderboard retrieved", "count", len(entries))
	return entries, nil
}

func (s *leaderboardService) GetTournamentLeaderboard(
	ctx context.Context,
	userId, tournamentId string,
) ([]repository.LeaderboardEntry, *apperrors.AppError) {
	s.logger.Info("Getting tournament leaderboard",
		"user_id", userId,
		"tournament_id", tournamentId,
	)

	entries, err := s.leaderboardRepo.GetGroupLeaderboard(ctx, userId, tournamentId)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Tournament leaderboard retrieved",
		"user_id", userId,
		"tournament_id", tournamentId,
		"count", len(entries),
	)
	return entries, nil
}

func (s *leaderboardService) GetTournamentRank(
	ctx context.Context,
	userId, tournamentId string,
) (int, *apperrors.AppError) {
	s.logger.Info("Getting tournament rank",
		"user_id", userId,
		"tournament_id", tournamentId,
	)

	rank, err := s.leaderboardRepo.GetGroupRank(ctx, userId, tournamentId)
	if err != nil {
		return 0, err
	}

	s.logger.Info("Tournament rank retrieved")

	return int(rank + 1), nil
}
