package models

import (
	"fmt"
	"time"
)

type Participation struct {
	UserId         string    `dynamodbav:"user_id"`
	TournamentId   string    `dynamodbav:"tournament_id"`
	GroupId        string    `dynamodbav:"group_id"`
	Score          int       `dynamodbav:"score"`
	RewardsClaimed bool      `dynamodbav:"rewards_claimed"`
	CreatedAt      time.Time `dynamodbav:"created_at"`
	UpdatedAt      time.Time `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`

	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
}

func UserGSI1PK(userID string) string {
	return fmt.Sprintf("USER#%s", userID)
}

func TournamentJoinedGSI1SK(tournamentID, joinedAt string) string {
	return fmt.Sprintf("TOURNAMENT#%s#JOINED#%s", tournamentID, joinedAt)
}
