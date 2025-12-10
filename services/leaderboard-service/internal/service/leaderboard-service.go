package service

import (
	"context"
	"fmt"

	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/repository"
)

type LeaderboardService interface {
	// Write Operations
	AddGlobalUser(ctx context.Context, userId, displayName string) error
	AddUserToTournament(ctx context.Context, userId, displayName, groupId, tournamentId string) error
	UpdateTournamentScore(ctx context.Context, userId, displayName, tournamentId string, score int) error

	// Read Operations
	GetGlobalLeaderboard(ctx context.Context) ([]repository.LeaderboardEntry, error)
	GetTournamentLeaderboard(ctx context.Context, userId, tournamentId string) ([]repository.LeaderboardEntry, error)
	GetTournamentRank(ctx context.Context, userId, tournamentId string) (int, error)
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

func (s *leaderboardService) AddGlobalUser(ctx context.Context, userId, displayName string) error {
	s.logger.Info("Adding global user")

	err := s.leaderboardRepo.AddGlobalUser(ctx, userId, displayName)
	if err != nil {
		s.logger.Error("Failed to add global user", "error", err)
		return fmt.Errorf("failed to add global user: %w", err)
	}

	s.logger.Info("Global user added")
	return nil
}

func (s *leaderboardService) AddUserToTournament(
	ctx context.Context,
	userId, displayName, groupId, tournamentId string,
) error {
	s.logger.Info("Adding tournament user")

	err := s.leaderboardRepo.AddUserToTournament(ctx, userId, displayName, groupId, tournamentId)
	if err != nil {
		s.logger.Error("Failed to add tournament user", "error", err)
		return fmt.Errorf("failed to add tournament user: %w", err)
	}

	s.logger.Info("Tournament user added")
	return nil
}

func (s *leaderboardService) UpdateTournamentScore(
	ctx context.Context,
	userId, displayName, tournamentId string,
	score int,
) error {
	s.logger.Info("Updating tournament score")

	err := s.leaderboardRepo.UpdateTournamentScore(ctx, userId, displayName, tournamentId, score)
	if err != nil {
		s.logger.Error("Failed to update tournament score", "error", err)
		return fmt.Errorf("failed to update tournament score: %w", err)
	}

	s.logger.Info("Tournament score updated")
	return nil
}

// Read Operations

func (s *leaderboardService) GetGlobalLeaderboard(ctx context.Context) ([]repository.LeaderboardEntry, error) {
	s.logger.Info("Getting global leaderboard")

	entries, err := s.leaderboardRepo.GetGlobalLeaderboard(ctx)
	if err != nil {
		s.logger.Error("Failed to get global leaderboard", "error", err)
		return nil, fmt.Errorf("failed to get global leaderboard: %w", err)
	}

	s.logger.Info("Global leaderboard retrieved", "count", len(entries))
	return entries, nil
}

func (s *leaderboardService) GetTournamentLeaderboard(
	ctx context.Context,
	userId, tournamentId string,
) ([]repository.LeaderboardEntry, error) {
	s.logger.Info("Getting tournament leaderboard",
		"user_id", userId,
		"tournament_id", tournamentId,
	)

	entries, err := s.leaderboardRepo.GetGroupLeaderboard(ctx, userId, tournamentId)
	if err != nil {
		s.logger.Error("Failed to get tournament leaderboard",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get tournament leaderboard: %w", err)
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
) (int, error) {
	s.logger.Info("Getting tournament rank",
		"user_id", userId,
		"tournament_id", tournamentId,
	)

	rank, err := s.leaderboardRepo.GetGroupRank(ctx, userId, tournamentId)
	if err != nil {
		s.logger.Error("Failed to get tournament rank",
			"error", err,
			"user_id", userId,
			"tournament_id", tournamentId,
		)
		return 0, fmt.Errorf("failed to get tournament rank: %w", err)
	}

	s.logger.Info("Tournament rank retrieved")

	return int(rank), nil
}
