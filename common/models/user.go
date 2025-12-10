package models

import (
	"fmt"
	"time"
)

type User struct {
	UserId      string    `dynamodbav:"user_id"`
	DisplayName string    `dynamodbav:"display_name"`
	Level       int       `dynamodbav:"level"`
	Coin        int       `dynamodbav:"coin"`
	CreatedAt   time.Time `dynamodbav:"created_at"`
	UpdatedAt   time.Time `dynamodbav:"updated_at"`

	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

// Key handlers
func UserPK(userId string) string {
	return fmt.Sprintf("USER#%s", userId)
}

func ProfileSK() string {
	return "PROFILE"
}

func UserSK(userId string) string {
	return fmt.Sprintf("USER#%s", userId)
}

func ExtractUserID(pk string) (string, error) {
	if len(pk) < 6 || pk[:5] != "USER#" {
		return "", fmt.Errorf("invalid user PK format: %s", pk)
	}
	return pk[5:], nil
}
