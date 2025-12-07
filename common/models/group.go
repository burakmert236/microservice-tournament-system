package models

import "fmt"

type Group struct {
	GroupId          string `dynamodbav:"group_id"`
	TournamentId     string `dynamodbav:"tournament_id"`
	GroupSize        int    `dynamodbav:"group_size"`
	ParticipantCount int    `dynamodbav:"participant_count"`
	IsFull           bool   `dynamodbav:"is_full"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

// Key handlers

func GroupSK(groupID string) string {
	return fmt.Sprintf("GROUP#%s", groupID)
}
