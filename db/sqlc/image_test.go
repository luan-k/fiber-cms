package db

import (
	"context"
	"database/sql"
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
	unused := createImageWithName(t, user.ID, fmt.Sprintf("unused_%s.jpg", timestamp), "Unused image")

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

	imageCountMap := make(map[string]int64)
	for _, img := range imagesWithCount {
		if img.Name == popular.Name {
			imageCountMap["popular"] = img.PostCount
		}
		if img.Name == moderate.Name {
			imageCountMap["moderate"] = img.PostCount
		}
		if img.Name == unused.Name {
			imageCountMap["unused"] = img.PostCount
		}
	}

	require.Equal(t, int64(3), imageCountMap["popular"], "Popular image should have 3 posts in ListImagesWithPostCount")
	require.Equal(t, int64(1), imageCountMap["moderate"], "Moderate image should have 1 post in ListImagesWithPostCount")
	require.Equal(t, int64(0), imageCountMap["unused"], "Unused image should have 0 posts in ListImagesWithPostCount")

	popularImages, err := testQueries.GetPopularImages(context.Background(), 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(popularImages), 2)

	popularFound := make(map[string]int64)
	for _, img := range popularImages {
		if img.Name == popular.Name {
			popularFound["popular"] = img.PostCount
		}
		if img.Name == moderate.Name {
			popularFound["moderate"] = img.PostCount
		}
		if img.Name == unused.Name {
			popularFound["unused"] = img.PostCount
		}
	}

	require.Equal(t, int64(3), popularFound["popular"], "Popular image should have 3 posts in GetPopularImages")
	require.Equal(t, int64(1), popularFound["moderate"], "Moderate image should have 1 post in GetPopularImages")

	require.Equal(t, int64(0), popularFound["unused"], "Unused image should not be in GetPopularImages")

	t.Logf("Popular (%s): %d posts", popular.Name, popularFound["popular"])
	t.Logf("Moderate (%s): %d posts", moderate.Name, popularFound["moderate"])
	t.Logf("Unused (%s): not in popular results (expected)", unused.Name)
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

func TestGetPostWithImages(t *testing.T) {
	user := createTestUser(t)
	_, image1 := createTestImage(t)
	_, image2 := createTestImage(t)

	arg := CreatePostWithImagesTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       gofakeit.Sentence(3),
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", gofakeit.UUID()),
		},
		AuthorIDs: []int64{user.ID},
		ImageIDs:  []int64{image1.ID, image2.ID},
	}

	result, err := testStore.CreatePostWithImagesTx(context.Background(), arg)
	require.NoError(t, err)

	postWithImages, err := testQueries.GetPostWithImages(context.Background(), result.Post.ID)
	require.NoError(t, err)
	require.NotEmpty(t, postWithImages)

	require.Equal(t, result.Post.ID, postWithImages.ID)
	require.Equal(t, result.Post.Title, postWithImages.Title)

	require.NotEmpty(t, postWithImages.Images)

	t.Logf("Post with images: %+v", postWithImages)
}

func TestListPostsWithImages(t *testing.T) {
	user := createTestUser(t)
	_, image1 := createTestImage(t)
	_, image2 := createTestImage(t)

	arg := CreatePostWithImagesTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       gofakeit.Sentence(3),
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", gofakeit.UUID()),
		},
		AuthorIDs: []int64{user.ID},
		ImageIDs:  []int64{image1.ID, image2.ID},
	}

	_, err := testStore.CreatePostWithImagesTx(context.Background(), arg)
	require.NoError(t, err)

	_, err = testStore.CreatePostTx(context.Background(), CreatePostTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       gofakeit.Sentence(3),
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", gofakeit.UUID()),
		},
		AuthorIDs: []int64{user.ID},
	})
	require.NoError(t, err)

	postsWithImages, err := testQueries.ListPostsWithImages(context.Background(), ListPostsWithImagesParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(postsWithImages), 2)

	for _, post := range postsWithImages {
		require.NotNil(t, post.Images)
		t.Logf("Post %d has images: %v", post.ID, post.Images)
	}
}

func TestGetPostsByUserWithImages(t *testing.T) {
	user := createTestUser(t)
	_, image1 := createTestImage(t)

	arg := CreatePostWithImagesTxParams{
		CreatePostsParams: CreatePostsParams{
			Title:       gofakeit.Sentence(3),
			Content:     gofakeit.Paragraph(3, 5, 10, " "),
			Description: gofakeit.Sentence(10),
			UserID:      user.ID,
			Username:    user.Username,
			Url:         fmt.Sprintf("https://example.com/posts/%s", gofakeit.UUID()),
		},
		AuthorIDs: []int64{user.ID},
		ImageIDs:  []int64{image1.ID},
	}

	_, err := testStore.CreatePostWithImagesTx(context.Background(), arg)
	require.NoError(t, err)

	userPosts, err := testQueries.GetPostsByUserWithImages(context.Background(), GetPostsByUserWithImagesParams{
		UserID: user.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(userPosts), 1)

	for _, post := range userPosts {
		require.Equal(t, user.ID, post.UserID)
		require.NotNil(t, post.Images)
	}
}

func TestGetImagesByUser(t *testing.T) {

	user1 := createTestUser(t)
	user2 := createTestUser(t)

	timestamp := time.Now().Format("20060102150405")
	image1 := createImageWithName(t, user1.ID, fmt.Sprintf("user1_image1_%s.jpg", timestamp), "User 1 first image")
	image2 := createImageWithName(t, user1.ID, fmt.Sprintf("user1_image2_%s.jpg", timestamp), "User 1 second image")
	image3 := createImageWithName(t, user1.ID, fmt.Sprintf("user1_image3_%s.jpg", timestamp), "User 1 third image")

	createImageWithName(t, user2.ID, fmt.Sprintf("user2_image1_%s.jpg", timestamp), "User 2 first image")
	createImageWithName(t, user2.ID, fmt.Sprintf("user2_image2_%s.jpg", timestamp), "User 2 second image")

	user1Images, err := testQueries.GetImagesByUser(context.Background(), GetImagesByUserParams{
		UserID: user1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(user1Images), 3)

	user1ImageIDs := make([]int64, 0)
	for _, img := range user1Images {
		require.Equal(t, user1.ID, img.UserID)
		if img.ID == image1.ID || img.ID == image2.ID || img.ID == image3.ID {
			user1ImageIDs = append(user1ImageIDs, img.ID)
		}
	}
	require.ElementsMatch(t, []int64{image1.ID, image2.ID, image3.ID}, user1ImageIDs)

	user2Images, err := testQueries.GetImagesByUser(context.Background(), GetImagesByUserParams{
		UserID: user2.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(user2Images), 2)

	for _, img := range user2Images {
		require.Equal(t, user2.ID, img.UserID)
	}

	user1ImagesPage1, err := testQueries.GetImagesByUser(context.Background(), GetImagesByUserParams{
		UserID: user1.ID,
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Len(t, user1ImagesPage1, 2)

	user1ImagesPage2, err := testQueries.GetImagesByUser(context.Background(), GetImagesByUserParams{
		UserID: user1.ID,
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(user1ImagesPage2), 1)

	page1IDs := make([]int64, len(user1ImagesPage1))
	for i, img := range user1ImagesPage1 {
		page1IDs[i] = img.ID
	}

	for _, img := range user1ImagesPage2 {
		require.NotContains(t, page1IDs, img.ID, "Pages should not have overlapping images")
	}

	userNoImages := createTestUser(t)
	noImages, err := testQueries.GetImagesByUser(context.Background(), GetImagesByUserParams{
		UserID: userNoImages.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Len(t, noImages, 0)
}

func TestGetUserImageCount(t *testing.T) {
	user := createTestUser(t)
	timestamp := time.Now().Format("20060102150405")

	count, err := testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	createImageWithName(t, user.ID, fmt.Sprintf("count_test1_%s.jpg", timestamp), "First image")
	count, err = testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	createImageWithName(t, user.ID, fmt.Sprintf("count_test2_%s.jpg", timestamp), "Second image")
	count, err = testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	image3 := createImageWithName(t, user.ID, fmt.Sprintf("count_test3_%s.jpg", timestamp), "Third image")
	count, err = testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	err = testQueries.DeleteImage(context.Background(), image3.ID)
	require.NoError(t, err)
	count, err = testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	user2 := createTestUser(t)
	count2, err := testQueries.GetUserImageCount(context.Background(), user2.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count2)

	createImageWithName(t, user2.ID, fmt.Sprintf("user2_count_%s.jpg", timestamp), "User2 image")
	count2, err = testQueries.GetUserImageCount(context.Background(), user2.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count2)

	count, err = testQueries.GetUserImageCount(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestSearchImagesByName(t *testing.T) {
	user := createTestUser(t)
	timestamp := time.Now().Format("20060102150405")

	js := createImageWithName(t, user.ID, fmt.Sprintf("javascript_%s.jpg", timestamp), "JavaScript tutorial screenshot")
	java := createImageWithName(t, user.ID, fmt.Sprintf("java_%s.png", timestamp), "Java programming guide")
	react := createImageWithName(t, user.ID, fmt.Sprintf("react_%s.svg", timestamp), "React component diagram")
	vue := createImageWithName(t, user.ID, fmt.Sprintf("vue_%s.jpg", timestamp), "Vue.js application screenshot")
	python := createImageWithName(t, user.ID, fmt.Sprintf("python_%s.png", timestamp), "Python script example")

	t.Logf("Created images:")
	t.Logf("  JavaScript: ID=%d, Name=%s, Description=%s", js.ID, js.Name, js.Description)
	t.Logf("  Java: ID=%d, Name=%s, Description=%s", java.ID, java.Name, java.Description)
	t.Logf("  React: ID=%d, Name=%s, Description=%s", react.ID, react.Name, react.Description)
	t.Logf("  Vue: ID=%d, Name=%s, Description=%s", vue.ID, vue.Name, vue.Description)
	t.Logf("  Python: ID=%d, Name=%s, Description=%s", python.ID, python.Name, python.Description)

	testCases := []struct {
		name           string
		searchTerm     string
		expectedCount  int
		mustContain    []int64
		mustNotContain []int64
	}{
		{
			name:           "Search for 'java' should find JavaScript and Java",
			searchTerm:     "java",
			expectedCount:  2,
			mustContain:    []int64{js.ID, java.ID},
			mustNotContain: []int64{react.ID, vue.ID, python.ID},
		},
		{
			name:           "Search for 'script' should find JavaScript and Python",
			searchTerm:     "script",
			expectedCount:  2,
			mustContain:    []int64{js.ID, python.ID},
			mustNotContain: []int64{java.ID, react.ID, vue.ID},
		},
		{
			name:           "Search for 'react' should find React",
			searchTerm:     "react",
			expectedCount:  1,
			mustContain:    []int64{react.ID},
			mustNotContain: []int64{js.ID, java.ID, vue.ID, python.ID},
		},
		{
			name:           "Search for 'component' should find React (in description)",
			searchTerm:     "component",
			expectedCount:  1,
			mustContain:    []int64{react.ID},
			mustNotContain: []int64{js.ID, java.ID, vue.ID, python.ID},
		},
		{
			name:           "Search for 'programming' should find Java (in description)",
			searchTerm:     "programming",
			expectedCount:  1,
			mustContain:    []int64{java.ID},
			mustNotContain: []int64{js.ID, react.ID, vue.ID, python.ID},
		},
		{
			name:           "Search for timestamp should find all our images",
			searchTerm:     timestamp,
			expectedCount:  5,
			mustContain:    []int64{js.ID, java.ID, react.ID, vue.ID, python.ID},
			mustNotContain: []int64{},
		},
		{
			name:           "Search for nonexistent should find nothing",
			searchTerm:     "definitely_does_not_exist_12345",
			expectedCount:  0,
			mustContain:    []int64{},
			mustNotContain: []int64{js.ID, java.ID, react.ID, vue.ID, python.ID},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
				Column1: sql.NullString{String: tc.searchTerm, Valid: true},
				Limit:   50,
				Offset:  0,
			})
			require.NoError(t, err)

			t.Logf("Search term: '%s' returned %d total results:", tc.searchTerm, len(results))
			for i, img := range results {
				if img.UserID == user.ID {
					t.Logf("  Result %d: ID=%d, Name=%s, Description=%s", i+1, img.ID, img.Name, img.Description)
				}
			}

			testResults := []Image{}
			testResultIDs := []int64{}
			for _, img := range results {
				if img.UserID == user.ID &&
					(img.ID == js.ID || img.ID == java.ID || img.ID == react.ID || img.ID == vue.ID || img.ID == python.ID) {
					testResults = append(testResults, img)
					testResultIDs = append(testResultIDs, img.ID)
				}
			}

			t.Logf("Filtered to our test images: %d results with IDs %v", len(testResults), testResultIDs)

			require.Len(t, testResults, tc.expectedCount, "Search term: '%s'. Expected %d results, got %d", tc.searchTerm, tc.expectedCount, len(testResults))

			resultIDs := make([]int64, len(testResults))
			for i, img := range testResults {
				resultIDs[i] = img.ID
			}

			for _, mustHaveID := range tc.mustContain {
				require.Contains(t, resultIDs, mustHaveID, "Search term '%s' should contain image ID %d", tc.searchTerm, mustHaveID)
			}

			for _, mustNotHaveID := range tc.mustNotContain {
				require.NotContains(t, resultIDs, mustNotHaveID, "Search term '%s' should NOT contain image ID %d", tc.searchTerm, mustNotHaveID)
			}

			t.Logf("Search term: '%s' found %d results", tc.searchTerm, len(testResults))
		})
	}
}

func TestSearchImagesByNameCaseInsensitive(t *testing.T) {
	user := createTestUser(t)
	timestamp := time.Now().Format("20060102150405")

	image := createImageWithName(t, user.ID, fmt.Sprintf("MixedCase_%s.jpg", timestamp), "MixedCase Description Test")

	testCases := []struct {
		name       string
		searchTerm string
	}{
		{"Lowercase search", "mixedcase"},
		{"Uppercase search", "MIXEDCASE"},
		{"Mixed case search", "MiXeDcAsE"},
		{"Description lowercase", "description"},
		{"Description uppercase", "DESCRIPTION"},
		{"Description mixed", "DeScRiPtIoN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
				Column1: sql.NullString{String: tc.searchTerm, Valid: true},
				Limit:   10,
				Offset:  0,
			})
			require.NoError(t, err)

			found := false
			for _, img := range results {
				if img.ID == image.ID {
					found = true
					break
				}
			}
			require.True(t, found, "Search term '%s' should find the MixedCase image", tc.searchTerm)
		})
	}
}

func TestSearchImagesByNamePagination(t *testing.T) {
	user := createTestUser(t)
	timestamp := time.Now().Format("20060102150405")

	searchTerm := fmt.Sprintf("pagination_test_%s", timestamp)
	var createdImages []Image
	for i := 0; i < 5; i++ {
		img := createImageWithName(t, user.ID, fmt.Sprintf("%s_image_%d.jpg", searchTerm, i), fmt.Sprintf("Pagination test image %d", i))
		createdImages = append(createdImages, img)
	}

	page1, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
		Column1: sql.NullString{String: searchTerm, Valid: true},
		Limit:   2,
		Offset:  0,
	})
	require.NoError(t, err)

	testPage1 := []Image{}
	for _, img := range page1 {
		if img.UserID == user.ID {
			for _, created := range createdImages {
				if img.ID == created.ID {
					testPage1 = append(testPage1, img)
				}
			}
		}
	}
	require.Len(t, testPage1, 2)

	page2, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
		Column1: sql.NullString{String: searchTerm, Valid: true},
		Limit:   2,
		Offset:  2,
	})
	require.NoError(t, err)

	testPage2 := []Image{}
	for _, img := range page2 {
		if img.UserID == user.ID {
			for _, created := range createdImages {
				if img.ID == created.ID {
					testPage2 = append(testPage2, img)
				}
			}
		}
	}
	require.Len(t, testPage2, 2)

	page3, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
		Column1: sql.NullString{String: searchTerm, Valid: true},
		Limit:   2,
		Offset:  4,
	})
	require.NoError(t, err)

	testPage3 := []Image{}
	for _, img := range page3 {
		if img.UserID == user.ID {
			for _, created := range createdImages {
				if img.ID == created.ID {
					testPage3 = append(testPage3, img)
				}
			}
		}
	}
	require.Len(t, testPage3, 1)

	allPageIDs := make([]int64, 0)
	for _, img := range testPage1 {
		allPageIDs = append(allPageIDs, img.ID)
	}
	for _, img := range testPage2 {
		require.NotContains(t, allPageIDs, img.ID, "Page 2 should not overlap with page 1")
		allPageIDs = append(allPageIDs, img.ID)
	}
	for _, img := range testPage3 {
		require.NotContains(t, allPageIDs, img.ID, "Page 3 should not overlap with pages 1 and 2")
		allPageIDs = append(allPageIDs, img.ID)
	}

	createdIDs := make([]int64, len(createdImages))
	for i, img := range createdImages {
		createdIDs[i] = img.ID
	}
	require.ElementsMatch(t, createdIDs, allPageIDs, "All created images should be found across pages")
}

func TestSearchImagesByNameEmptyTerm(t *testing.T) {

	results, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
		Column1: sql.NullString{String: "", Valid: true},
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(results), 0)

	results2, err := testQueries.SearchImagesByName(context.Background(), SearchImagesByNameParams{
		Column1: sql.NullString{Valid: false},
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err)

	require.Len(t, results2, 0)
}
