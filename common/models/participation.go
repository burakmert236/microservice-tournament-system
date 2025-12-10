package models

import (
	"fmt"
	"time"
)

type RewardClaimStatus string

const (
	Unclaimed  RewardClaimStatus = "UNCLAIMED"
	Processing RewardClaimStatus = "PROCESSING"
	Claimed    RewardClaimStatus = "CLAIMED"
)

type Participation struct {
	UserId            string            `dynamodbav:"user_id"`
	TournamentId      string            `dynamodbav:"tournament_id"`
	GroupId           string            `dynamodbav:"group_id"`
	Score             int               `dynamodbav:"score"`
	RewardClaimStatus RewardClaimStatus `dynamodbav:"reward_claim_status"`
	EndsAt            time.Time         `dynamodbav:"ends_at"`
	RewardingMap      map[string]int    `dynamodbav:"rewarding_map"`
	CreatedAt         time.Time         `dynamodbav:"created_at"`
	UpdatedAt         time.Time         `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

func UserGSI1PK(userId string) string {
	return fmt.Sprintf("USER#%s", userId)
}

func TournamentJoinedGSI1SK(tournamentId, joinedAt string) string {
	return fmt.Sprintf("TOURNAMENT#%s#JOINED#%s", tournamentId, joinedAt)
}
