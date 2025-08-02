package token

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = fmt.Errorf("token is invalid")
	ErrExpiredToken = fmt.Errorf("token has expired")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	TokenType string    `json:"token_type"`
}

func NewPayload(userID int64, username string, duration time.Duration, tokenType string) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		UserID:    userID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		TokenType: tokenType,
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	// TODO: Implement token blacklisting when Redis/database is available
	/* if isTokenBlacklisted(payload.ID) {
	    return ErrInvalidToken
	} */

	return nil
}

// TODO: Implement this function with Redis or database
/* func isTokenBlacklisted(tokenID uuid.UUID) bool {
	// Check Redis or database for blacklisted tokens
	return false
} */
