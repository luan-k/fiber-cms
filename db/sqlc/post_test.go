package db

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func createPost(t *testing.T) Post {
	gofakeit.Seed(0)
	user := createTestUser(t)

	title := gofakeit.Sentence(3)
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))

	arg := CreatePostsParams{
		Name:        title,
		Content:     gofakeit.Paragraph(3, 5, 10, " "),
		Description: gofakeit.Sentence(10),
		UserID:      user.ID,
		Username:    user.Username,
		Url:         fmt.Sprintf("https://example.com/posts/%s", slug),
		Images:      []string{gofakeit.ImageURL(800, 600), gofakeit.ImageURL(800, 600)},
	}
	post, err := testQueries.CreatePosts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post)

	require.Equal(t, arg.Name, post.Name)
	require.Equal(t, arg.Content, post.Content)
	require.Equal(t, arg.Description, post.Description)
	require.Equal(t, arg.UserID, post.UserID)
	require.Equal(t, arg.Username, post.Username)
	require.Equal(t, arg.Url, post.Url)
	require.ElementsMatch(t, arg.Images, post.Images)
	require.NotZero(t, post.ID)
	require.NotZero(t, post.CreatedAt)

	return post
}

func TestCreatePosts(t *testing.T) {
	post := createPost(t)
	require.NotEmpty(t, post)
}

func TestGetPost(t *testing.T) {
	post1 := createPost(t)
	post2, err := testQueries.GetPost(context.Background(), post1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, post2)
	require.Equal(t, post1.ID, post2.ID)
	require.Equal(t, post1.Name, post2.Name)
	require.Equal(t, post1.Content, post2.Content)
	require.Equal(t, post1.Description, post2.Description)
	require.Equal(t, post1.UserID, post2.UserID)
	require.Equal(t, post1.Username, post2.Username)
	require.Equal(t, post1.Url, post2.Url)
	require.ElementsMatch(t, post1.Images, post2.Images)
	require.WithinDuration(t, post1.CreatedAt, post2.CreatedAt, 0)
}

func TestListPosts(t *testing.T) {
	for range 10 {
		createPost(t)
	}

	arg := ListPostsParams{
		Limit:  5,
		Offset: 5,
	}
	posts, err := testQueries.ListPosts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, posts, 5)

	for _, post := range posts {
		require.NotEmpty(t, post)
		require.NotZero(t, post.ID)
		require.NotZero(t, post.CreatedAt)
	}
}

func TestUpdatePost(t *testing.T) {
	post1 := createPost(t)
	require.NotEmpty(t, post1)

	gofakeit.Seed(1)
	newTitle := gofakeit.Sentence(3)

	arg := UpdatePostParams{
		ID:          post1.ID,
		Name:        newTitle,
		Content:     gofakeit.Paragraph(3, 5, 10, " "),
		Description: gofakeit.Sentence(10),
		UserID:      post1.UserID,
		Username:    post1.Username,
		Url:         post1.Url,
		Images:      []string{gofakeit.ImageURL(800, 600), gofakeit.ImageURL(800, 600)},
	}
	post2, err := testQueries.UpdatePost(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, post2)
	require.Equal(t, post1.ID, post2.ID)
	require.Equal(t, arg.Name, post2.Name)
	require.Equal(t, arg.Content, post2.Content)
	require.Equal(t, arg.Description, post2.Description)
	require.Equal(t, post1.UserID, post2.UserID)
	require.Equal(t, post1.Username, post2.Username)
	require.Equal(t, arg.Url, post2.Url)
	require.ElementsMatch(t, arg.Images, post2.Images)
	require.WithinDuration(t, post1.CreatedAt, post2.CreatedAt, 0)
	// TODO: not changing ChangedAt here, as it is not updated in the current implementation
	//require.NotEqual(t, post1.ChangedAt, post2.ChangedAt)
}

func TestDeletePost(t *testing.T) {
	post1 := createPost(t)
	err := testQueries.DeletePost(context.Background(), post1.ID)
	require.NoError(t, err)

	post2, err := testQueries.GetPost(context.Background(), post1.ID)
	require.Error(t, err)
	require.EqualError(t, err, "sql: no rows in result set")
	require.Empty(t, post2)
}
