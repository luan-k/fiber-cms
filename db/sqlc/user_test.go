package db

import (
	"context"
	"fmt"
	"strings"
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

func createTestUserWithPosts(t *testing.T) (User, CreatePostTxResult) {
	user := createTestUser(t)

	title := gofakeit.Sentence(3)
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))

	postArg := CreatePostTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       title,
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", slug),
			Images:      []string{gofakeit.ImageURL(800, 600)},
		},
		AuthorIDs: []int64{user.ID},
	}

	post, err := testStore.CreatePostTx(context.Background(), postArg)
	require.NoError(t, err)

	return user, post
}

func TestCreateUser(t *testing.T) {
	user := createTestUser(t)
	require.NotEmpty(t, user)
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

func TestDeleteUserWithTransferTx(t *testing.T) {
	user, post := createTestUserWithPosts(t)

	adminUser := createTestUser(t)

	err := testStore.DeleteUserWithTransferTx(context.Background(), DeleteUserWithTransferTxParams{
		UserID:       user.ID,
		TransferToID: adminUser.ID,
	})
	require.NoError(t, err)

	deletedUser, err := testQueries.GetUser(context.Background(), user.ID)
	require.Error(t, err)
	require.EqualError(t, err, "sql: no rows in result set")
	require.Empty(t, deletedUser)

	existingPost, err := testQueries.GetPost(context.Background(), post.Post.ID)
	require.NoError(t, err)
	require.NotEmpty(t, existingPost)
	require.Equal(t, adminUser.ID, existingPost.UserID)
	require.Equal(t, adminUser.Username, existingPost.Username)
}

func TestDeleteUserTx_NuclearOption(t *testing.T) {
	user, post := createTestUserWithPosts(t)

	err := testStore.DeleteUserTx(context.Background(), user.ID)
	require.NoError(t, err)

	deletedUser, err := testQueries.GetUser(context.Background(), user.ID)
	require.Error(t, err)
	require.Empty(t, deletedUser)

	deletedPost, err := testQueries.GetPost(context.Background(), post.Post.ID)
	require.Error(t, err)
	require.Empty(t, deletedPost)
}

func TestDeleteUser_WithoutTransaction_ShouldFail(t *testing.T) {
	user, _ := createTestUserWithPosts(t)

	err := testQueries.DeleteUser(context.Background(), user.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "foreign key constraint")
}

func TestUpdateUserTx(t *testing.T) {
	user := createTestUser(t)

	newUsername := fmt.Sprintf("updated_%s_%d", gofakeit.Username(), time.Now().UnixNano())
	newEmail := fmt.Sprintf("updated_%d_%s", time.Now().UnixNano(), gofakeit.Email())

	arg := UpdateUserTxParams{
		UpdateUserParams: UpdateUserParams{
			ID:             user.ID,
			Username:       newUsername,
			FullName:       gofakeit.Name(),
			Email:          newEmail,
			HashedPassword: gofakeit.Password(true, true, true, true, false, 32),
			Role:           "admin",
		},
		CheckUniqueness: true,
	}

	result, err := testStore.UpdateUserTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result.User)
	require.Equal(t, arg.Username, result.User.Username)
	require.Equal(t, arg.Email, result.User.Email)
	require.Equal(t, arg.Role, result.User.Role)
}

func TestUpdateUserTx_UniqueConstraintViolation(t *testing.T) {
	user1 := createTestUser(t)
	user2 := createTestUser(t)

	arg := UpdateUserTxParams{
		UpdateUserParams: UpdateUserParams{
			ID:       user2.ID,
			Username: user1.Username,
			Email:    user2.Email,
		},
		CheckUniqueness: true,
	}

	_, err := testStore.UpdateUserTx(context.Background(), arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}
