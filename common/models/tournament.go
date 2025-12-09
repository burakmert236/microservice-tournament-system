package models

import (
	"fmt"
	"time"
)

type Tournament struct {
	TournamentId                 string         `dynamodbav:"tournament_id"`
	StartsAt                     time.Time      `dynamodbav:"starts_at"`
	EndsAt                       time.Time      `dynamodbav:"ends_at"`
	LastAllowedParticipationDate time.Time      `dynamodbav:"last_allowed_participation_date"`
	ScoreRewardPerLevelUpgrade   int            `dynamodbav:"score_reward_per_level_upgrade"`
	GroupSize                    int            `dynamodbav:"group_size"`
	UserLevelLimit               int            `dynamodbav:"user_level_limit"`
	EnteranceFee                 int            `dynamodbav:"enterance_fee"`
	RewardingMap                 map[string]int `dynamodbav:"rewarding_map"`
	CreatedAt                    time.Time      `dynamodbav:"created_at"`
	UpdatedAt                    time.Time      `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`

	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
}

// Key handlers
func TournamentPK(tournamentId string) string {
	return fmt.Sprintf("TOURNAMENT#%s", tournamentId)
}

func MetaSK() string {
	return "META"
}

func TournamentGSI1PK() string {
	return "CURRENT_TOURNAMENT"
}

func StartTimeGSI1SK(startTime string) string {
	return fmt.Sprintf("START#%s", startTime)
}

func ExtractTournamentID(pk string) (string, error) {
	if len(pk) < 12 || pk[:11] != "TOURNAMENT#" {
		return "", fmt.Errorf("invalid tournament PK format: %s", pk)
	}
	return pk[11:], nil
}
