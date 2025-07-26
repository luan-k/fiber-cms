package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) User {
	gofakeit.Seed(0)

	arg := CreateUserParams{
		Username:       gofakeit.Username(),
		Email:          gofakeit.Email(),
		FullName:       gofakeit.Name(),
		HashedPassword: gofakeit.Password(true, true, true, true, false, 32),
		Role:           "user",
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Role, user.Role)

	return user
}

func TestCreateUser(t *testing.T) {
	user := createTestUser(t)
	require.NotEmpty(t, user)
}

func TestDeleteUser(t *testing.T) {
	user := createTestUser(t)
	err := testQueries.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)

	user2, err := testQueries.GetUser(context.Background(), user.ID)
	require.Error(t, err)
	require.EqualError(t, err, "sql: no rows in result set")
	require.Empty(t, user2)
}

func TestGetUser(t *testing.T) {
	user1 := createTestUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Role, user2.Role)
}

func TestGetUserByEmail(t *testing.T) {
	user1 := createTestUser(t)
	user2, err := testQueries.GetUserByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Role, user2.Role)
}

func TestGetUserByUsername(t *testing.T) {
	user1 := createTestUser(t)
	user2, err := testQueries.GetUserByUsername(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Role, user2.Role)
}

func TestListUsers(t *testing.T) {
	for range 10 {
		createTestUser(t)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}
	users, err := testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, users, 5)

	for _, user := range users {
		require.NotEmpty(t, user)
		require.NotZero(t, user.ID)
		require.NotZero(t, user.CreatedAt)
		require.NotEmpty(t, user.Username)
		require.NotEmpty(t, user.Email)
		require.NotEmpty(t, user.FullName)
		require.NotEmpty(t, user.HashedPassword)
		require.NotEmpty(t, user.Role)
	}
}

func TestUpdateUser(t *testing.T) {
	user1 := createTestUser(t)
	require.NotEmpty(t, user1)

	newUsername := fmt.Sprintf("%s_%d", gofakeit.Username(), time.Now().UnixNano())
	newEmail := fmt.Sprintf("%d_%s", time.Now().UnixNano(), gofakeit.Email())

	arg := UpdateUserParams{
		ID:             user1.ID,
		Username:       newUsername,
		FullName:       gofakeit.Name(),
		Email:          newEmail,
		HashedPassword: gofakeit.Password(true, true, true, true, false, 32),
		Role:           "admin",
	}
	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, arg.ID, user2.ID)
	require.Equal(t, arg.Username, user2.Username)
	require.Equal(t, arg.FullName, user2.FullName)
	require.Equal(t, arg.Email, user2.Email)
	require.Equal(t, arg.HashedPassword, user2.HashedPassword)
	require.Equal(t, arg.Role, user2.Role)
}
