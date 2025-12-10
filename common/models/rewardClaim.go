package models

import (
	"fmt"
	"time"
)

type RewardClaim struct {
	UserId       string    `dynamodbav:"user_id"`
	TournamentId string    `dynamodbav:"tournament_id"`
	CreatedAt    time.Time `dynamodbav:"created_at"`
	UpdatedAt    time.Time `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

// Key handlers
func RewardClaimPK(userId string) string {
	return fmt.Sprintf("REWARDCLAIM#%s", userId)
}
