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
	ReservationId string            `dynamodbav:"reservation_id"`
	UserId        string            `dynamodbav:"user_id"`
	Amount        int64             `dynamodbav:"amount"`
	Status        ReservationStatus `dynamodbav:"status"`
	Purpose       string            `dynamodbav:"purpose"`
	CreatedAt     time.Time         `dynamodbav:"created_at"`
	UpdatedAt     time.Time         `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

func ReservationPK(reservationId string) string {
	return fmt.Sprintf("RESERVATION#%s", reservationId)
}

func ReservationSK() string {
	return "META"
}
