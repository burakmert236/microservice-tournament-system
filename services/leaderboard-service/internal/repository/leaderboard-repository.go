package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/burakmert236/goodswipe-common/cache"
	"github.com/burakmert236/goodswipe-common/logger"
	"github.com/redis/go-redis/v9"
)

const (
	GlobalLeaderboardLimit = 1000
	DefaultTTL             = 7 * 24 * time.Hour
)

type LeaderboardRepository struct {
	client *redis.Client
	logger *logger.Logger
}

func NewLeaderboardRepository(redisClient *cache.RedisClient, log *logger.Logger) *LeaderboardRepository {
	return &LeaderboardRepository{
		client: redisClient.GetClient(),
		logger: log.With("component", "LeaderboardRepository"),
	}
}

// Key Generation (Private Helpers)

func globalLeaderboardKey() string {
	return "leaderboard:global"
}

func groupLeaderboardKey(tournamentId, groupId string) string {
	return fmt.Sprintf("leaderboard:group:%s:%s", tournamentId, groupId)
}

// Write Operations

// AddUserToTournament adds user to tournament with 0 score
func (r *LeaderboardRepository) AddGlobalUser(ctx context.Context, userId string) error {
	pipe := r.client.Pipeline()

	globalLeaderboardKey := globalLeaderboardKey()

	member := redis.Z{
		Score:  0,
		Member: userId,
	}
	if err := pipe.ZAdd(ctx, globalLeaderboardKey, member).Err(); err != nil {
		return fmt.Errorf("failed to add user to global leaderboard: %w", err)
	}
	pipe.ZRemRangeByRank(ctx, globalLeaderboardKey, 0, -GlobalLeaderboardLimit-1)

	pipe.Expire(ctx, globalLeaderboardKey, DefaultTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to add global user",
			"error", err,
			"user_id", userId,
		)
		return fmt.Errorf("failed to update tournament score: %w", err)
	}

	return nil
}

func (r *LeaderboardRepository) AddUserToTournament(ctx context.Context, userId, groupId, tournamentId string) error {
	pipe := r.client.Pipeline()

	leaderboardKey := groupLeaderboardKey(tournamentId, groupId)

	member := redis.Z{
		Score:  0,
		Member: userId,
	}
	if err := pipe.ZAdd(ctx, leaderboardKey, member).Err(); err != nil {
		return fmt.Errorf("failed to add user to group leaderboard: %w", err)
	}

	pipe.Expire(ctx, leaderboardKey, DefaultTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to add global user",
			"error", err,
			"user_id", userId,
			"tournament_id", tournamentId,
		)
		return fmt.Errorf("failed to update tournament score: %w", err)
	}

	return nil
}

// UpdateTournamentScore updates score for a specific tournament (NOT cumulative)
func (r *LeaderboardRepository) UpdateTournamentScore(
	ctx context.Context,
	userId string,
	tournamentId string,
	groupId string,
	score float64,
) error {
	pipe := r.client.Pipeline()

	member := redis.Z{
		Score:  score,
		Member: userId,
	}

	pipe.ZAdd(ctx, groupLeaderboardKey(tournamentId, groupId), member)
	pipe.Expire(ctx, groupLeaderboardKey(tournamentId, groupId), DefaultTTL)

	pipe.ZAdd(ctx, globalLeaderboardKey(), member)
	pipe.Expire(ctx, globalLeaderboardKey(), DefaultTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to update tournament score",
			"error", err,
			"user_id", userId,
			"tournament_id", tournamentId,
		)
		return fmt.Errorf("failed to update tournament score: %w", err)
	}

	return nil
}

// Read Operations

// GetGlobalLeaderboard returns top N users from global leaderboard
func (r *LeaderboardRepository) GetGlobalLeaderboard(ctx context.Context) ([]redis.Z, error) {
	r.logger.Debug("Getting global leaderboard")

	result, err := r.client.ZRevRangeWithScores(ctx, globalLeaderboardKey(), 0, GlobalLeaderboardLimit-1).Result()
	if err != nil {
		r.logger.Error("Failed to get global leaderboard",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get global leaderboard: %w", err)
	}

	return result, nil
}

// GetGroupLeaderboard returns all users in a specific group
func (r *LeaderboardRepository) GetGroupLeaderboard(
	ctx context.Context,
	tournamentId string,
	groupId string,
) ([]redis.Z, error) {
	r.logger.Debug("Getting group leaderboard",
		"tournament_id", tournamentId,
		"group_id", groupId,
	)

	key := groupLeaderboardKey(tournamentId, groupId)

	result, err := r.client.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		r.logger.Error("Failed to get group leaderboard",
			"error", err,
			"tournament_id", tournamentId,
			"group_id", groupId,
		)
		return nil, fmt.Errorf("failed to get group leaderboard: %w", err)
	}

	return result, nil
}

// GetGroupRank returns user's rank within their group (1-based)
func (r *LeaderboardRepository) GetGroupRank(
	ctx context.Context,
	tournamentId string,
	groupId string,
	userId string,
) (int64, float64, error) {
	r.logger.Debug("Getting group rank",
		"tournament_id", tournamentId,
		"group_id", groupId,
		"user_id", userId,
	)

	key := groupLeaderboardKey(tournamentId, groupId)

	// Get rank
	rank, err := r.client.ZRevRank(ctx, key, userId).Result()
	if err == redis.Nil {
		return -1, 0, nil // User not in group
	} else if err != nil {
		r.logger.Error("Failed to get group rank",
			"error", err,
			"tournament_id", tournamentId,
			"group_id", groupId,
			"user_id", userId,
		)
		return -1, 0, fmt.Errorf("failed to get group rank: %w", err)
	}

	// Get score
	score, err := r.client.ZScore(ctx, key, userId).Result()
	if err != nil {
		return -1, 0, fmt.Errorf("failed to get score: %w", err)
	}

	return rank + 1, score, nil
}
