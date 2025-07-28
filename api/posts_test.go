package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdb "github.com/luan-k/fiber-cms/db/mock"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

func randomUser() *db.User {
	gofakeit.Seed(0)
	return &db.User{
		ID:                gofakeit.Int64(),
		Username:          gofakeit.Username(),
		FullName:          gofakeit.Name(),
		Email:             gofakeit.Email(),
		HashedPassword:    gofakeit.Password(true, true, true, true, false, 12),
		PasswordChangedAt: gofakeit.Date(),
		CreatedAt:         gofakeit.Date(),
		Role:              gofakeit.Word(),
	}
}

func randomPost(user *db.User) db.Post {
	gofakeit.Seed(0)
	slug := gofakeit.Sentence(3)
	return db.Post{
		ID:          gofakeit.Int64(),
		Title:       gofakeit.Sentence(3),
		Content:     gofakeit.Paragraph(3, 5, 10, " "),
		Description: gofakeit.Sentence(10),
		UserID:      user.ID,
		Username:    user.Username,
		Url:         fmt.Sprintf("https://example.com/posts/%s", slug),
	}
}

func TestGetPostAPI(t *testing.T) {
	gofakeit.Seed(0)
	user := randomUser()
	post := randomPost(user)

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: user.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPost(t, recorder.Body.String(), post)
			},
		},
		{
			name:      "NotFound",
			accountID: user.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(db.Post{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: user.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(db.Post{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: user.ID,
			buildStubs: func(store *mockdb.MockStore) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			var url string
			if tc.name == "InvalidID" {

				url = "/api/v1/posts/invalid_id"
			} else {
				url = fmt.Sprintf("/api/v1/posts/%d", post.ID)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})

	}

}

func TestListPostsAPI(t *testing.T) {
	user := &db.User{
		ID:       1,
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}

	posts := []db.Post{
		{
			ID:          1,
			Title:       "First Post",
			Content:     "Content of first post",
			Description: "Description of first post",
			UserID:      user.ID,
			Username:    user.Username,
			Url:         "https://example.com/posts/first-post",
		},
		{
			ID:          2,
			Title:       "Second Post",
			Content:     "Content of second post",
			Description: "Description of second post",
			UserID:      user.ID,
			Username:    user.Username,
			Url:         "https://example.com/posts/second-post",
		},
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?limit=10&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPosts(gomock.Any(), db.ListPostsParams{
						Limit:  10,
						Offset: 0,
					}).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPosts(t, recorder.Body.String(), posts)
			},
		},
		{
			name:  "OKWithPagination",
			query: "?limit=5&offset=5",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPosts(gomock.Any(), db.ListPostsParams{
						Limit:  5,
						Offset: 5,
					}).
					Times(1).
					Return([]db.Post{posts[1]}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPosts(t, recorder.Body.String(), []db.Post{posts[1]})
			},
		},
		{
			name:  "InternalError",
			query: "?limit=10&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPosts(gomock.Any(), db.ListPostsParams{
						Limit:  10,
						Offset: 0,
					}).
					Times(1).
					Return([]db.Post{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidLimit",
			query: "?limit=0&offset=0",
			buildStubs: func(store *mockdb.MockStore) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "LimitTooHigh",
			query: "?limit=200&offset=0",
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListPosts(gomock.Any(), db.ListPostsParams{
						Limit:  100,
						Offset: 0,
					}).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPosts(t, recorder.Body.String(), posts)
			},
		},
		{
			name:  "InvalidOffset",
			query: "?limit=10&offset=-1",
			buildStubs: func(store *mockdb.MockStore) {

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "EmptyResult",
			query: "?limit=10&offset=100",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPosts(gomock.Any(), db.ListPostsParams{
						Limit:  10,
						Offset: 100,
					}).
					Times(1).
					Return([]db.Post{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPosts(t, recorder.Body.String(), []db.Post{})
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/api/v1/posts" + tc.query
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

type GetPostResponse struct {
	Post db.Post `json:"post"`
}

type ListPostsResponse struct {
	Posts []db.Post `json:"posts"`
}

func requireBodyMatchPost(t *testing.T, body string, post db.Post) {
	data, err := io.ReadAll(io.NopCloser(strings.NewReader(body)))
	require.NoError(t, err)

	var response GetPostResponse
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	gotPost := response.Post

	require.Equal(t, post.ID, gotPost.ID)
	require.Equal(t, post.Title, gotPost.Title)
	require.Equal(t, post.Content, gotPost.Content)
	require.Equal(t, post.Description, gotPost.Description)
	require.Equal(t, post.UserID, gotPost.UserID)
	require.Equal(t, post.Username, gotPost.Username)
	require.Equal(t, post.Url, gotPost.Url)
}

func requireBodyMatchPosts(t *testing.T, body string, posts []db.Post) {
	data, err := io.ReadAll(io.NopCloser(strings.NewReader(body)))
	require.NoError(t, err)

	var response ListPostsResponse
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	require.Equal(t, len(posts), len(response.Posts))

	for i, post := range posts {
		gotPost := response.Posts[i]
		require.Equal(t, post.ID, gotPost.ID)
		require.Equal(t, post.Title, gotPost.Title)
		require.Equal(t, post.Content, gotPost.Content)
		require.Equal(t, post.Description, gotPost.Description)
		require.Equal(t, post.UserID, gotPost.UserID)
		require.Equal(t, post.Username, gotPost.Username)
		require.Equal(t, post.Url, gotPost.Url)
	}
}
