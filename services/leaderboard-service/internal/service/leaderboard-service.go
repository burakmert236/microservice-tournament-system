package service

import (
	"context"

	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/burakmert236/goodswipe-leaderboard-service/internal/repository"
)

type LeaderboardService interface {
	GetGlobalLeaderboard(ctx context.Context) error
	GetTournamentLeaderboard(ctx context.Context, userId, tournamentId string) error
	GetTournamentRank(ctx context.Context, userId, tournamentId string) error
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

func (s *leaderboardService) GetGlobalLeaderboard(ctx context.Context) error {
	return nil
}

func (s *leaderboardService) GetTournamentLeaderboard(ctx context.Context, userId, tournamentId string) error {
	return nil
}

func (s *leaderboardService) GetTournamentRank(ctx context.Context, userId, tournamentId string) error {
	return nil
}
