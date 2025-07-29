package util

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := gofakeit.Password(true, true, true, true, false, 10)

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	require.NotEqual(t, password, hashedPassword)

	require.Contains(t, hashedPassword, "$2a$")
}

func TestCheckPassword(t *testing.T) {
	password := gofakeit.Password(true, true, true, true, false, 10)

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = CheckPassword(password, hashedPassword)
	require.NoError(t, err)

	wrongPassword := gofakeit.Password(true, true, true, true, false, 8)
	err = CheckPassword(wrongPassword, hashedPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}

func TestHashPassword_MinLength(t *testing.T) {

	shortPassword := "123"
	_, err := HashPassword(shortPassword)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least 6 characters")

	validPassword := "123456"
	hashedPassword, err := HashPassword(validPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
}

func TestHashPassword_Uniqueness(t *testing.T) {
	password := gofakeit.Password(true, true, true, true, false, 10)

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)

	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)

	require.NotEqual(t, hashedPassword1, hashedPassword2)

	require.NoError(t, CheckPassword(password, hashedPassword1))
	require.NoError(t, CheckPassword(password, hashedPassword2))
}

func TestCheckPasswordMatch(t *testing.T) {
	password := gofakeit.Password(true, true, true, true, false, 10)
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	require.True(t, CheckPasswordMatch(password, hashedPassword))

	wrongPassword := gofakeit.Password(true, true, true, true, false, 8)
	require.False(t, CheckPasswordMatch(wrongPassword, hashedPassword))
}

func TestPassword_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "very short password",
			password: "12345",
			wantErr:  true,
		},
		{
			name:     "minimum valid password",
			password: "123456",
			wantErr:  false,
		},
		{
			name:     "normal password",
			password: "mySecurePassword123",
			wantErr:  false,
		},
		{
			name:     "maximum valid password (72 chars)",
			password: "a" + gofakeit.Password(true, true, true, true, false, 71),
			wantErr:  false,
		},
		{
			name:     "password too long (73 chars)",
			password: "a" + gofakeit.Password(true, true, true, true, false, 72),
			wantErr:  true,
		},
		{
			name:     "password with special chars",
			password: "p@ssw0rd!#$%",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tc.password)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, hashedPassword)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, hashedPassword)

				require.NoError(t, CheckPassword(tc.password, hashedPassword))
			}
		})
	}
}

func TestHashPassword_Cost(t *testing.T) {
	password := "testPassword123"

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)

	err = CheckPassword(password, hashedPassword)
	require.NoError(t, err)
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123"

	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	password := "benchmarkPassword123"
	hashedPassword, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := CheckPassword(password, hashedPassword)
		if err != nil {
			b.Fatal(err)
		}
	}
}
