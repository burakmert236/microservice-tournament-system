package models

import (
	"fmt"
	"time"
)

type ReservationStatus string

const (
	ReservationStatusReserved   ReservationStatus = "RESERVED"
	ReservationStatusConfirmed  ReservationStatus = "CONFIRMED"
	ReservationStatusRolledBack ReservationStatus = "ROLLED_BACK"
)

type Reservation struct {
	UserId       string            `dynamodbav:"user_id"`
	TournamentId string            `dynamodbav:"tournament_id"`
	Amount       int64             `dynamodbav:"amount"`
	Status       ReservationStatus `dynamodbav:"status"`
	Purpose      string            `dynamodbav:"purpose"`
	CreatedAt    time.Time         `dynamodbav:"created_at"`
	UpdatedAt    time.Time         `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

func ReservationPK(userId string) string {
	return fmt.Sprintf("RESERVATION#%s", userId)
}

func ReservationSK(tournamentId string) string {
	return fmt.Sprintf("TOURNAMENT#%s", tournamentId)
}
