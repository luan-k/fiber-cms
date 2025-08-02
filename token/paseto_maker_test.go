package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	userID := int64(123)
	username := "testuser"
	duration := time.Minute

	t.Run("CreateAccessToken", func(t *testing.T) {
		token, err := maker.CreateToken(userID, username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, payload)

		require.Equal(t, userID, payload.UserID)
		require.Equal(t, username, payload.Username)
		require.Equal(t, "access", payload.TokenType)
		require.WithinDuration(t, time.Now(), payload.IssuedAt, time.Second)
		require.WithinDuration(t, time.Now().Add(duration), payload.ExpiredAt, time.Second)
	})

	t.Run("CreateRefreshToken", func(t *testing.T) {
		refreshDuration := time.Hour * 24 * 7
		token, err := maker.CreateRefreshToken(userID, username, refreshDuration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, payload)

		require.Equal(t, userID, payload.UserID)
		require.Equal(t, username, payload.Username)
		require.Equal(t, "refresh", payload.TokenType)
		require.WithinDuration(t, time.Now(), payload.IssuedAt, time.Second)
		require.WithinDuration(t, time.Now().Add(refreshDuration), payload.ExpiredAt, time.Second)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		token, err := maker.CreateToken(userID, username, -time.Minute)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.Error(t, err)
		require.EqualError(t, err, ErrExpiredToken.Error())
		require.Nil(t, payload)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		payload, err := maker.VerifyToken("invalid.token.here")
		require.Error(t, err)
		require.EqualError(t, err, ErrInvalidToken.Error())
		require.Nil(t, payload)
	})

	t.Run("TokenTypes", func(t *testing.T) {
		accessToken, err := maker.CreateToken(userID, username, duration)
		require.NoError(t, err)

		refreshToken, err := maker.CreateRefreshToken(userID, username, duration)
		require.NoError(t, err)

		require.NotEqual(t, accessToken, refreshToken)

		accessPayload, err := maker.VerifyToken(accessToken)
		require.NoError(t, err)
		require.Equal(t, "access", accessPayload.TokenType)

		refreshPayload, err := maker.VerifyToken(refreshToken)
		require.NoError(t, err)
		require.Equal(t, "refresh", refreshPayload.TokenType)
	})
}

func TestInvalidSymmetricKeySize(t *testing.T) {
	maker, err := NewPasetoMaker("invalid-key")
	require.Error(t, err)
	require.Nil(t, maker)
}

func TestPayloadValidation(t *testing.T) {

	payload := &Payload{
		UserID:    123,
		Username:  "testuser",
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(time.Minute),
		TokenType: "access",
	}

	err := payload.Valid()
	require.NoError(t, err)

	expiredPayload := &Payload{
		UserID:    123,
		Username:  "testuser",
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(-time.Minute),
		TokenType: "access",
	}

	err = expiredPayload.Valid()
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
}

func TestTokenSecurity(t *testing.T) {
	maker1, err := NewPasetoMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	maker2, err := NewPasetoMaker("98765432109876543210987654321098")
	require.NoError(t, err)

	userID := int64(123)
	username := "testuser"
	duration := time.Minute

	token, err := maker1.CreateToken(userID, username, duration)
	require.NoError(t, err)

	payload, err := maker2.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}

func BenchmarkPasetoMaker_CreateToken(b *testing.B) {
	maker, err := NewPasetoMaker("12345678901234567890123456789012")
	require.NoError(b, err)

	userID := int64(123)
	username := "testuser"
	duration := time.Minute

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := maker.CreateToken(userID, username, duration)
		require.NoError(b, err)
	}
}

func BenchmarkPasetoMaker_VerifyToken(b *testing.B) {
	maker, err := NewPasetoMaker("12345678901234567890123456789012")
	require.NoError(b, err)

	userID := int64(123)
	username := "testuser"
	duration := time.Minute

	token, err := maker.CreateToken(userID, username, duration)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := maker.VerifyToken(token)
		require.NoError(b, err)
	}
}
