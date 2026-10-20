package api

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMeaningImagesAPIRequiresAuthenticationAndBoundedReadOnlyLookup(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	api := humachi.New(chi.NewMux(), huma.DefaultConfig("images", "test"))
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authStubService()))
	RegisterMeaningImages(api, meaningmedia.NewService(db))
	id := uuid.NewString()
	send := func(method, query string, authenticated bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/meaning-images"+query, nil)
		if authenticated {
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uuid.New()}))
		}
		w := httptest.NewRecorder()
		api.Adapter().ServeHTTP(w, req)
		return w
	}
	require.Equal(t, 401, send("GET", "?meaningIds="+id, false).Code)
	require.Equal(t, 400, send("GET", "?meaningIds=bad", true).Code)
	require.Equal(t, 422, send("GET", "?meaningIds="+strings.Repeat(id+",", 101), true).Code)
	require.Equal(t, 405, send("POST", "?meaningIds="+id, true).Code)
	mock.ExpectQuery("SELECT a.meaning_id").WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "sha256", "alt", "width", "height", "word", "pos", "definition"}))
	w := send("GET", "?meaningIds="+id, true)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"items":[]`)
	require.Equal(t, "private, max-age=60", w.Header().Get("Cache-Control"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCanonicalWordEmbedsOnlyMatchingMeaningImageAndRetainsTextOnMediaFailure(t *testing.T) {
	for _, tc := range []string{"approved", "different revision", "media unavailable"} {
		t.Run(tc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo, reader := contentSampleData()
			api := humachi.New(chi.NewMux(), huma.DefaultConfig("image detail", "test"))
			api.UseMiddleware(withHumaContext)
			api.UseMiddleware(AuthMiddleware(authStubService()))
			RegisterContentWithImages(api, content.NewService(repo, reader), nil, meaningmedia.NewService(db))
			id := "00000000-0000-0000-0000-000000000003"
			definition := "A document that lets you get on your flight."
			revision := meaningidentity.Revision(id, "boarding pass", "noun", definition)
			if tc == "media unavailable" {
				mock.ExpectQuery("SELECT a.meaning_id").WillReturnError(errors.New("storage metadata offline"))
			} else {
				word := "boarding pass"
				if tc == "different revision" {
					word = "changed spelling"
					revision = meaningidentity.Revision(id, word, "noun", definition)
				}
				mock.ExpectQuery("SELECT a.meaning_id").WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "sha256", "alt", "width", "height", "word", "pos", "definition"}).AddRow(id, revision, strings.Repeat("a", 64), "A passenger showing a boarding pass.", 960, 640, word, "noun", definition))
			}
			req := httptest.NewRequest("GET", "/api/v1/canonical-words/boarding-pass", nil)
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: uuid.MustParse("00000000-0000-0000-0000-000000000009")}))
			w := httptest.NewRecorder()
			api.Adapter().ServeHTTP(w, req)
			require.Equal(t, 200, w.Code, w.Body.String())
			require.Contains(t, w.Body.String(), definition)
			if tc == "approved" {
				require.Contains(t, w.Body.String(), `"imageUrl":"https://images.vocanova.site/meanings/`+id+`/`)
			} else {
				require.NotContains(t, w.Body.String(), `"image":`)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
