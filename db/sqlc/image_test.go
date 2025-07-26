package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func createTestImage(t *testing.T) (User, Image) {
	user := createTestUser(t)

	gofakeit.Seed(0)
	arg := CreateImageParams{
		Name:        gofakeit.Word(),
		Description: gofakeit.Sentence(10),
		Alt:         gofakeit.Sentence(5),
		ImagePath:   fmt.Sprintf("/uploads/images/%s.jpg", gofakeit.UUID()),
		UserID:      user.ID,
	}

	image, err := testQueries.CreateImage(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, image)
	require.Equal(t, arg.Name, image.Name)
	require.Equal(t, arg.Description, image.Description)
	require.Equal(t, arg.Alt, image.Alt)
	require.Equal(t, arg.ImagePath, image.ImagePath)
	require.Equal(t, arg.UserID, image.UserID)
	require.NotZero(t, image.ID)

	return user, image
}

func TestCreateImage(t *testing.T) {
	_, image := createTestImage(t)
	require.NotEmpty(t, image)
}

func TestGetImage(t *testing.T) {
	_, image1 := createTestImage(t)
	image2, err := testQueries.GetImage(context.Background(), image1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, image2)
	require.Equal(t, image1.ID, image2.ID)
	require.Equal(t, image1.Name, image2.Name)
	require.Equal(t, image1.UserID, image2.UserID)
}

func TestListImages(t *testing.T) {
	for range 10 {
		createTestImage(t)
	}

	images, err := testQueries.ListImages(context.Background(), ListImagesParams{
		Limit:  5,
		Offset: 5,
	})
	require.NoError(t, err)
	require.Len(t, images, 5)

	for _, image := range images {
		require.NotEmpty(t, image)
		require.NotZero(t, image.ID)
		require.NotEmpty(t, image.Name)
		require.NotEmpty(t, image.UserID)
	}
}

func TestUpdateImage(t *testing.T) {
	_, image1 := createTestImage(t)

	newName := gofakeit.Word()
	newDescription := gofakeit.Sentence(15)

	arg := UpdateImageParams{
		ID:          image1.ID,
		Name:        newName,
		Description: newDescription,
		Alt:         image1.Alt,
		ImagePath:   image1.ImagePath,
	}

	image2, err := testQueries.UpdateImage(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, image2)
	require.Equal(t, image1.ID, image2.ID)
	require.Equal(t, newName, image2.Name)
	require.Equal(t, newDescription, image2.Description)
}

func TestDeleteImageTx(t *testing.T) {
	user, image := createTestImage(t)
	_, post := createTestUserWithPosts(t)

	_, err := testQueries.CreatePostImage(context.Background(), CreatePostImageParams{
		PostID:  post.Post.ID,
		ImageID: image.ID,
		Order:   0,
	})
	require.NoError(t, err)

	err = testStore.DeleteImageTx(context.Background(), DeleteImageTxParams{
		ImageID: image.ID,
		UserID:  user.ID,
	})
	require.NoError(t, err)

	deletedImage, err := testQueries.GetImage(context.Background(), image.ID)
	require.Error(t, err)
	require.Empty(t, deletedImage)

	postImages, err := testQueries.GetImagesByPost(context.Background(), post.Post.ID)
	require.NoError(t, err)
	require.Len(t, postImages, 0)
}

func TestCreatePostWithImagesTx(t *testing.T) {
	user := createTestUser(t)
	_, image1 := createTestImage(t)
	_, image2 := createTestImage(t)

	title := gofakeit.Sentence(3)

	arg := CreatePostWithImagesTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       title,
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", gofakeit.UUID()),
			Images:      []string{},
		},
		AuthorIDs: []int64{user.ID},
		ImageIDs:  []int64{image1.ID, image2.ID},
	}

	result, err := testStore.CreatePostWithImagesTx(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, result.Post)
	require.Len(t, result.UserPosts, 1)
	require.Len(t, result.PostImages, 2)

	postImages, err := testQueries.GetImagesByPost(context.Background(), result.Post.ID)
	require.NoError(t, err)
	require.Len(t, postImages, 2)
}

func TestImageAnalyticsQueries(t *testing.T) {
	timestamp := time.Now().Format("20060102150405")

	user := createTestUser(t)

	popular := createImageWithName(t, user.ID, fmt.Sprintf("popular_%s.jpg", timestamp), "Popular image")
	moderate := createImageWithName(t, user.ID, fmt.Sprintf("moderate_%s.jpg", timestamp), "Moderate image")
	createImageWithName(t, user.ID, fmt.Sprintf("unused_%s.jpg", timestamp), "Unused image")

	_, post1 := createTestUserWithPosts(t)
	_, post2 := createTestUserWithPosts(t)
	_, post3 := createTestUserWithPosts(t)

	for i, post := range []CreatePostTxResult{post1, post2, post3} {
		_, err := testQueries.CreatePostImage(context.Background(), CreatePostImageParams{
			PostID:  post.Post.ID,
			ImageID: popular.ID,
			Order:   int32(i),
		})
		require.NoError(t, err)
	}

	_, err := testQueries.CreatePostImage(context.Background(), CreatePostImageParams{
		PostID:  post1.Post.ID,
		ImageID: moderate.ID,
		Order:   1,
	})
	require.NoError(t, err)

	imagesWithCount, err := testQueries.ListImagesWithPostCount(context.Background(), ListImagesWithPostCountParams{
		Limit:  50,
		Offset: 0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(imagesWithCount), 3)

	popularImages, err := testQueries.GetPopularImages(context.Background(), 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(popularImages), 2)

	found := make(map[string]int64)
	for _, img := range popularImages {
		if img.Name == popular.Name {
			found["popular"] = img.PostCount
		}
		if img.Name == moderate.Name {
			found["moderate"] = img.PostCount
		}
	}

	require.Equal(t, int64(3), found["popular"])
	require.Equal(t, int64(1), found["moderate"])
}

func createImageWithName(t *testing.T, userID int64, name, description string) Image {
	arg := CreateImageParams{
		Name:        name,
		Description: description,
		Alt:         fmt.Sprintf("Alt text for %s", name),
		ImagePath:   fmt.Sprintf("/uploads/images/%s", name),
		UserID:      userID,
	}

	image, err := testQueries.CreateImage(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, image)
	require.NotZero(t, image.ID)

	return image
}
