package models

import (
	"fmt"
	"time"
)

type Tournament struct {
	TournamentId                 string           `dynamodbav:"tournament_id"`
	Name                         string           `dynamodbav:"name"`
	Status                       TournamentStatus `dynamodbav:"status"`
	StartsAt                     time.Time        `dynamodbav:"starts_at"`
	EndsAt                       time.Time        `dynamodbav:"ends_at"`
	LastAllowedParticipationDate time.Time        `dynamodbav:"last_allowed_participation_date"`
	ScoreRewardPerLevelUpgrade   int              `dynamodbav:"score_reward_per_level_upgrade"`
	GroupSize                    int              `dynamodbav:"group_size"`
	RewardingMap                 map[string]int   `dynamodbav:"rewarding_map"`
	CreatedAt                    time.Time        `dynamodbav:"created_at"`
	UpdatedAt                    time.Time        `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`

	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
}

type TournamentStatus int

const (
	Ongoing TournamentStatus = iota
	Completed
)

var tournamentStatusNames = map[TournamentStatus]string{
	Ongoing:   "Ongoing",
	Completed: "Completed",
}

func (s TournamentStatus) String() string {
	return tournamentStatusNames[s]
}

// Key handlers
func TournamentPK(tournamentID string) string {
	return fmt.Sprintf("TOURNAMENT#%s", tournamentID)
}

func MetaSK() string {
	return "META"
}

func StatusGSI1PK(status string) string {
	return fmt.Sprintf("STATUS#%s", status)
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
