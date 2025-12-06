package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const GlobalLimit = 1000

type LeaderboardRepo struct {
	client *redis.Client
	ctx    context.Context
}

func NewLeaderboardRepo(client *redis.Client) *LeaderboardRepo {
	return &LeaderboardRepo{
		client: client,
		ctx:    context.Background(),
	}
}

// Key Generation Helpers

func globalKey() string {
	return "leaderboard:global"
}

func tournamentKey(tournamentID string) string {
	return fmt.Sprintf("leaderboard:tournament:%s", tournamentID)
}

func groupKey(tournamentID string, groupID string) string {
	return fmt.Sprintf("leaderboard:group:%s:%s", tournamentID, groupID)
}

// Write Operations

// UpdateScore updates a user's score across all relevant leaderboards.
func (r *LeaderboardRepo) UpdateScore(
	user_id string,
	score float64,
	tournament_id string,
	group_id string) error {

	pipe := r.client.Pipeline()

	member := redis.Z{
		Score:  score,
		Member: user_id,
	}

	pipe.ZAdd(r.ctx, globalKey(), member)
	pipe.ZRemRangeByRank(r.ctx, globalKey(), 0, -GlobalLimit-1)

	pipe.ZAdd(r.ctx, tournamentKey(tournament_id), member)

	pipe.ZAdd(r.ctx, groupKey(tournament_id, group_id), member)

	_, err := pipe.Exec(r.ctx)
	return err
}

// Read Operations

// GetUserGroupRanking retrieves a user's 1-based ranking and score
// within their specific tournament group.
func (r *LeaderboardRepo) GetUserGroupRanking(
	user_id string,
	tournament_id string,
	group_id string) (int64, error) {

	key := groupKey(tournament_id, group_id)

	rank, err := r.client.ZRevRank(r.ctx, key, user_id).Result()

	if err == redis.Nil {
		return -1, nil
	} else if err != nil {
		return -1, err
	}

	return rank, nil
}

// GetGroupLeaderboard retrieves a range of users and scores for a specific group.
func (r *LeaderboardRepo) GetGroupLeaderboard(
	tournament_id string,
	group_id string) ([]redis.Z, error) {

	key := groupKey(tournament_id, group_id)

	return r.client.ZRevRangeWithScores(r.ctx, key, 0, -1).Result()
}
