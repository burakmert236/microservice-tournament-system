package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/burakmert236/goodswipe-common/cache"
	apperrors "github.com/burakmert236/goodswipe-common/errors"
	"github.com/burakmert236/goodswipe-common/logger"
	leaderboarderrors "github.com/burakmert236/goodswipe-leaderboard-service/internal/errors"
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

func usernamesHashKey() string {
	return "usernames"
}

func userGroupMappingsHashKey() string {
	return "user:group"
}

func groupLeaderboardKey(tournamentId, groupId string) string {
	return fmt.Sprintf("leaderboard:group:%s:%s", tournamentId, groupId)
}

func userTournamentField(userId, tournamentId string) string {
	return fmt.Sprintf("%s:%s", userId, tournamentId)
}

// Write Operations

// AddUserToTournament adds user to tournament with 0 score
func (r *LeaderboardRepository) AddGlobalUser(ctx context.Context, userId, displayName string) *apperrors.AppError {
	pipe := r.client.Pipeline()

	pipe.HSet(ctx, usernamesHashKey(), userId, displayName)

	globalLeaderboardKey := globalLeaderboardKey()

	member := redis.Z{
		Score:  0,
		Member: userId,
	}
	if err := pipe.ZAdd(ctx, globalLeaderboardKey, member).Err(); err != nil {
		return leaderboarderrors.UserNotExistsInAnyGroup()
	}
	pipe.ZRemRangeByRank(ctx, globalLeaderboardKey, 0, -GlobalLeaderboardLimit-1)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to add global user",
			"error", err,
			"user_id", userId,
		)
		return apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to update tournament score")
	}

	return nil
}

func (r *LeaderboardRepository) AddUserToTournament(
	ctx context.Context,
	userId, displayName, groupId, tournamentId string,
) *apperrors.AppError {
	pipe := r.client.Pipeline()

	pipe.HSet(ctx, usernamesHashKey(), userId, displayName)
	pipe.HSet(ctx, userGroupMappingsHashKey(), userTournamentField(userId, tournamentId), groupId)
	pipe.Expire(ctx, userGroupMappingsHashKey(), DefaultTTL)

	leaderboardKey := groupLeaderboardKey(tournamentId, groupId)

	member := redis.Z{
		Score:  0,
		Member: userId,
	}
	if err := pipe.ZAdd(ctx, leaderboardKey, member).Err(); err != nil {
		return apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to add user to group leaderboard")
	}
	pipe.Expire(ctx, leaderboardKey, DefaultTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		r.logger.Error("Failed to add global user",
			"error", err,
			"user_id", userId,
			"tournament_id", tournamentId,
		)
		return apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to update tournament score")
	}

	return nil
}

// UpdateTournamentScore updates score for a specific tournament (NOT cumulative)
func (r *LeaderboardRepository) UpdateTournamentScore(
	ctx context.Context,
	userId, tournamentId string,
	score int,
) *apperrors.AppError {
	groupId, err := r.client.HGet(ctx, userGroupMappingsHashKey(), userTournamentField(userId, tournamentId)).Result()
	if err == redis.Nil {
		return leaderboarderrors.UserNotExistsInAnyGroup()
	} else if err != nil {
		return apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get user group")
	}

	pipe := r.client.Pipeline()

	member := redis.Z{
		Score:  float64(score),
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
		return apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to update tournament score")
	}

	return nil
}

// Read Operations

type LeaderboardEntry struct {
	UserId      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Score       float64 `json:"score"`
	Rank        int64   `json:"rank"`
}

func (r *LeaderboardRepository) generateLeaderboardEntryList(ctx context.Context, result []redis.Z) []LeaderboardEntry {
	entries := make([]LeaderboardEntry, len(result))

	for i, z := range result {
		userId := z.Member.(string)
		displayName, err := r.client.HGet(ctx, usernamesHashKey(), userId).Result()
		if err != nil {
			r.logger.Error("Failed to get username from hash",
				"userId", userId,
			)
			displayName = ""
		}

		entries[i] = LeaderboardEntry{
			UserId:      userId,
			DisplayName: displayName,
			Score:       z.Score,
			Rank:        int64(i + 1),
		}
	}

	return entries
}

// GetGlobalLeaderboard returns top N users from global leaderboard
func (r *LeaderboardRepository) GetGlobalLeaderboard(ctx context.Context) ([]LeaderboardEntry, *apperrors.AppError) {
	r.logger.Debug("Getting global leaderboard")

	result, err := r.client.ZRevRangeWithScores(ctx, globalLeaderboardKey(), 0, GlobalLeaderboardLimit-1).Result()
	if err != nil {
		r.logger.Error("Failed to get global leaderboard",
			"error", err,
		)
		return nil, apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get global leaderboard")
	}

	return r.generateLeaderboardEntryList(ctx, result), nil
}

// GetGroupLeaderboard returns all users in a specific group
func (r *LeaderboardRepository) GetGroupLeaderboard(
	ctx context.Context,
	userId, tournamentId string,
) ([]LeaderboardEntry, *apperrors.AppError) {
	r.logger.Debug("Getting group leaderboard",
		"tournament_id", tournamentId,
		"user_id", userId,
	)

	groupId, err := r.client.HGet(ctx, userGroupMappingsHashKey(), userTournamentField(userId, tournamentId)).Result()
	if err == redis.Nil {
		return nil, leaderboarderrors.UserNotExistsInAnyGroup()
	} else if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get user group")
	}

	key := groupLeaderboardKey(tournamentId, groupId)

	result, err := r.client.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		r.logger.Error("Failed to get group leaderboard",
			"error", err,
			"tournament_id", tournamentId,
			"group_id", groupId,
		)
		return nil, apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get group leaderboard")
	}

	return r.generateLeaderboardEntryList(ctx, result), nil
}

// GetGroupRank returns user's rank within their group (1-based)
func (r *LeaderboardRepository) GetGroupRank(
	ctx context.Context,
	userId, tournamentId string,
) (int64, *apperrors.AppError) {
	r.logger.Debug("Getting group rank",
		"tournament_id", tournamentId,
		"user_id", userId,
	)

	groupId, err := r.client.HGet(ctx, userGroupMappingsHashKey(), userTournamentField(userId, tournamentId)).Result()
	if err == redis.Nil {
		return 0, leaderboarderrors.UserNotExistsInAnyGroup()
	} else if err != nil {
		return 0, apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get user group")
	}

	key := groupLeaderboardKey(tournamentId, groupId)

	rank, err := r.client.ZRevRank(ctx, key, userId).Result()
	if err == redis.Nil {
		return -1, nil
	} else if err != nil {
		r.logger.Error("Failed to get group rank",
			"error", err,
			"tournament_id", tournamentId,
			"group_id", groupId,
			"user_id", userId,
		)
		return -1, apperrors.Wrap(err, apperrors.CodeRedisOperationError, "failed to get group rank")
	}

	return rank, nil
}
