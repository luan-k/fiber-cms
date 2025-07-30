package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdb "github.com/luan-k/fiber-cms/db/mock"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

func randomTaxonomy() db.Taxonomy {
	gofakeit.Seed(0)
	return db.Taxonomy{
		ID:          gofakeit.Int64(),
		Name:        gofakeit.BuzzWord(),
		Description: gofakeit.Sentence(10),
	}
}

func TestCreateTaxonomyAPI(t *testing.T) {
	taxonomy := randomTaxonomy()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":        taxonomy.Name,
				"description": taxonomy.Description,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Eq(taxonomy.Name)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)

				arg := db.CreateTaxonomyParams{
					Name:        taxonomy.Name,
					Description: taxonomy.Description,
				}
				store.EXPECT().
					CreateTaxonomy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(taxonomy, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchTaxonomy(t, recorder.Body.String(), taxonomy)
			},
		},
		{
			name: "DuplicateName",
			body: gin.H{
				"name":        taxonomy.Name,
				"description": taxonomy.Description,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Eq(taxonomy.Name)).
					Times(1).
					Return(taxonomy, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
		{
			name: "InvalidName",
			body: gin.H{
				"name":        "A",
				"description": taxonomy.Description,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidDescription",
			body: gin.H{
				"name":        taxonomy.Name,
				"description": "Bad",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Any()).
					Times(0)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/taxonomies"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetTaxonomyAPI(t *testing.T) {
	taxonomy := randomTaxonomy()

	testCases := []struct {
		name          string
		taxonomyID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			taxonomyID: taxonomy.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTaxonomy(t, recorder.Body.String(), taxonomy)
			},
		},
		{
			name:       "NotFound",
			taxonomyID: taxonomy.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			taxonomyID: taxonomy.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			taxonomyID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Any()).
					Times(0)
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
				url = "/api/v1/taxonomies/invalid_id"
			} else {
				url = fmt.Sprintf("/api/v1/taxonomies/%d", tc.taxonomyID)
			}

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListTaxonomiesAPI(t *testing.T) {
	n := 5
	taxonomies := make([]db.Taxonomy, n)
	for i := 0; i < n; i++ {
		taxonomies[i] = randomTaxonomy()
		taxonomies[i].ID = int64(i + 1)
	}

	taxonomiesWithCount := make([]db.ListTaxonomiesWithPostCountRow, n)
	for i := 0; i < n; i++ {
		taxonomiesWithCount[i] = db.ListTaxonomiesWithPostCountRow{
			ID:          taxonomies[i].ID,
			Name:        taxonomies[i].Name,
			Description: taxonomies[i].Description,
			PostCount:   int64(i * 2),
		}
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?limit=5&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTaxonomies(gomock.Any(), db.ListTaxonomiesParams{
						Limit:  5,
						Offset: 0,
					}).
					Times(1).
					Return(taxonomies, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTaxonomies(t, recorder.Body.String(), taxonomies)
			},
		},
		{
			name:  "WithCounts",
			query: "?limit=5&offset=0&with_counts=true",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTaxonomiesWithPostCount(gomock.Any(), db.ListTaxonomiesWithPostCountParams{
						Limit:  5,
						Offset: 0,
					}).
					Times(1).
					Return(taxonomiesWithCount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTaxonomiesWithCount(t, recorder.Body.String(), taxonomiesWithCount)
			},
		},
		{
			name:  "InternalError",
			query: "?limit=5&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTaxonomies(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Taxonomy{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidLimit",
			query: "?limit=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTaxonomies(gomock.Any(), gomock.Any()).
					Times(0)
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

			url := "/api/v1/taxonomies" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetPopularTaxonomiesAPI(t *testing.T) {
	n := 3
	popularTaxonomies := make([]db.GetPopularTaxonomiesRow, n)
	for i := 0; i < n; i++ {
		taxonomy := randomTaxonomy()
		taxonomy.ID = int64(i + 1)
		popularTaxonomies[i] = db.GetPopularTaxonomiesRow{
			ID:          taxonomy.ID,
			Name:        taxonomy.Name,
			Description: taxonomy.Description,
			PostCount:   int64((n - i) * 10),
		}
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?limit=10",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPopularTaxonomies(gomock.Any(), int32(10)).
					Times(1).
					Return(popularTaxonomies, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPopularTaxonomies(t, recorder.Body.String(), popularTaxonomies)
			},
		},
		{
			name:  "InternalError",
			query: "?limit=10",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPopularTaxonomies(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.GetPopularTaxonomiesRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/api/v1/taxonomies/popular" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestSearchTaxonomiesAPI(t *testing.T) {
	taxonomy := randomTaxonomy()
	searchResults := []db.Taxonomy{taxonomy}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "?q=tech&limit=10&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchTaxonomiesByName(gomock.Any(), db.SearchTaxonomiesByNameParams{
						Column1: sql.NullString{String: "tech", Valid: true},
						Limit:   10,
						Offset:  0,
					}).
					Times(1).
					Return(searchResults, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTaxonomies(t, recorder.Body.String(), searchResults)
			},
		},
		{
			name:  "EmptyQuery",
			query: "?q=&limit=10",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchTaxonomiesByName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "MissingQuery",
			query: "?limit=10",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchTaxonomiesByName(gomock.Any(), gomock.Any()).
					Times(0)
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

			url := "/api/v1/taxonomies/search" + tc.query
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateTaxonomyAPI(t *testing.T) {
	taxonomy := randomTaxonomy()
	newName := gofakeit.BuzzWord()
	newDescription := gofakeit.Sentence(10)

	testCases := []struct {
		name          string
		taxonomyID    int64
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			taxonomyID: taxonomy.ID,
			body: gin.H{
				"name":        newName,
				"description": newDescription,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Eq(newName)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)

				updatedTaxonomy := taxonomy
				updatedTaxonomy.Name = newName
				updatedTaxonomy.Description = newDescription

				store.EXPECT().
					UpdateTaxonomy(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updatedTaxonomy, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "NotFound",
			taxonomyID: taxonomy.ID,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "DuplicateName",
			taxonomyID: taxonomy.ID,
			body: gin.H{
				"name": newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyByName(gomock.Any(), gomock.Eq(newName)).
					Times(1).
					Return(db.Taxonomy{ID: 999, Name: newName}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/v1/taxonomies/%d", tc.taxonomyID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteTaxonomyAPI(t *testing.T) {
	taxonomy := randomTaxonomy()

	testCases := []struct {
		name          string
		taxonomyID    int64
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK_NoPosts",
			taxonomyID: taxonomy.ID,
			query:      "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyPostCount(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(int64(0), nil)

				store.EXPECT().
					DeleteTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "ConflictWithPosts",
			taxonomyID: taxonomy.ID,
			query:      "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyPostCount(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(int64(5), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
		{
			name:       "ForceDelete",
			taxonomyID: taxonomy.ID,
			query:      "?force=true",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyPostCount(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(int64(5), nil)

				store.EXPECT().
					DeleteTaxonomyPosts(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(nil)

				store.EXPECT().
					DeleteTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "NotFound",
			taxonomyID: taxonomy.ID,
			query:      "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

			url := fmt.Sprintf("/api/v1/taxonomies/%d%s", tc.taxonomyID, tc.query)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetTaxonomyPostsAPI(t *testing.T) {
	taxonomy := randomTaxonomy()
	user := randomUserForPosts()
	post := randomPost(user)
	posts := []db.Post{post}

	testCases := []struct {
		name          string
		taxonomyID    int64
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			taxonomyID: taxonomy.ID,
			query:      "?limit=10&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(taxonomy, nil)

				store.EXPECT().
					GetTaxonomyPosts(gomock.Any(), db.GetTaxonomyPostsParams{
						TaxonomyID: taxonomy.ID,
						Limit:      10,
						Offset:     0,
					}).
					Times(1).
					Return(posts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "TaxonomyNotFound",
			taxonomyID: taxonomy.ID,
			query:      "?limit=10&offset=0",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTaxonomy(gomock.Any(), gomock.Eq(taxonomy.ID)).
					Times(1).
					Return(db.Taxonomy{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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

			url := fmt.Sprintf("/api/v1/taxonomies/%d/posts%s", tc.taxonomyID, tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetPostTaxonomiesAPI(t *testing.T) {
	user := randomUserForPosts()
	post := randomPost(user)
	taxonomy := randomTaxonomy()
	taxonomies := []db.Taxonomy{taxonomy}

	testCases := []struct {
		name          string
		postID        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			postID: post.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPost(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)

				store.EXPECT().
					GetPostTaxonomies(gomock.Any(), gomock.Eq(post.ID)).
					Times(1).
					Return(taxonomies, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "PostNotFound",
			postID: post.ID,
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

			url := fmt.Sprintf("/api/v1/posts/%d/taxonomies", tc.postID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchTaxonomy(t *testing.T, body string, taxonomy db.Taxonomy) {
	var response struct {
		Taxonomy TaxonomyResponse `json:"taxonomy"`
	}
	err := json.Unmarshal([]byte(body), &response)
	require.NoError(t, err)

	require.Equal(t, taxonomy.ID, response.Taxonomy.ID)
	require.Equal(t, taxonomy.Name, response.Taxonomy.Name)
	require.Equal(t, taxonomy.Description, response.Taxonomy.Description)
}

func requireBodyMatchTaxonomies(t *testing.T, body string, taxonomies []db.Taxonomy) {
	var response struct {
		Taxonomies []TaxonomyResponse `json:"taxonomies"`
		Meta       struct {
			Limit      int  `json:"limit"`
			Offset     int  `json:"offset"`
			Count      int  `json:"count"`
			WithCounts bool `json:"with_counts"`
		} `json:"meta"`
	}
	err := json.Unmarshal([]byte(body), &response)
	require.NoError(t, err)

	require.Equal(t, len(taxonomies), len(response.Taxonomies))
	for i, taxonomy := range taxonomies {
		require.Equal(t, taxonomy.ID, response.Taxonomies[i].ID)
		require.Equal(t, taxonomy.Name, response.Taxonomies[i].Name)
		require.Equal(t, taxonomy.Description, response.Taxonomies[i].Description)
	}
}

func requireBodyMatchTaxonomiesWithCount(t *testing.T, body string, taxonomies []db.ListTaxonomiesWithPostCountRow) {
	var response struct {
		Taxonomies []TaxonomyResponse `json:"taxonomies"`
		Meta       struct {
			Limit      int  `json:"limit"`
			Offset     int  `json:"offset"`
			Count      int  `json:"count"`
			WithCounts bool `json:"with_counts"`
		} `json:"meta"`
	}
	err := json.Unmarshal([]byte(body), &response)
	require.NoError(t, err)

	require.Equal(t, len(taxonomies), len(response.Taxonomies))
	require.True(t, response.Meta.WithCounts)
	for i, taxonomy := range taxonomies {
		require.Equal(t, taxonomy.ID, response.Taxonomies[i].ID)
		require.Equal(t, taxonomy.Name, response.Taxonomies[i].Name)
		require.Equal(t, taxonomy.Description, response.Taxonomies[i].Description)
		require.Equal(t, taxonomy.PostCount, *response.Taxonomies[i].PostCount)
	}
}

func requireBodyMatchPopularTaxonomies(t *testing.T, body string, taxonomies []db.GetPopularTaxonomiesRow) {
	var response struct {
		Taxonomies []PopularTaxonomyResponse `json:"taxonomies"`
		Meta       struct {
			Limit int `json:"limit"`
			Count int `json:"count"`
		} `json:"meta"`
	}
	err := json.Unmarshal([]byte(body), &response)
	require.NoError(t, err)

	require.Equal(t, len(taxonomies), len(response.Taxonomies))
	for i, taxonomy := range taxonomies {
		require.Equal(t, taxonomy.ID, response.Taxonomies[i].ID)
		require.Equal(t, taxonomy.Name, response.Taxonomies[i].Name)
		require.Equal(t, taxonomy.Description, response.Taxonomies[i].Description)
		require.Equal(t, taxonomy.PostCount, response.Taxonomies[i].PostCount)
	}
}
