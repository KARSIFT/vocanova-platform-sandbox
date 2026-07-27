package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
)

// VOC-030-T06 — API-layer cross-cutting unauthorized/cross-user
// safety tests (VOC-030-TEST-32, VOC-030-AC-08, VOC-030-R03). The
// per-task auth tests in apps/api/app/api/learning_test.go,
// reviews_test.go, aifeedback_test.go, and missions_test.go already
// exercise the unauthenticated path on each endpoint in isolation.
// This file is the T06 cross-cutting counterpart that drives the
// new P4 reads through a single shared API harness and verifies
// the cross-cutting guarantees that go beyond a single endpoint:
// (a) both new reads reject the same invalid IANA timezone
// (VOC-030-R02) — the gamification.ResolveSettings validator is the
// single source of truth, and a regression in either endpoint's
// validation would let a malformed client value corrupt
// daily-date-boundary math; (b) neither new read accepts a path
// parameter that could be used to enumerate another learner's
// mission/streak/point data; and (c) the contract-level
// cross-user-isolation guarantee is structural, not a UI
// convention, so the test guards the OpenAPI rather than runtime
// behavior.

// newT06MissionsAPI builds a huma.API with the P4 read endpoints
// (daily-mission, progress) wired against a single sqlmock-backed
// database. The strict-mode sqlmock catches any SQL the
// auth-failed or invalid-timezone request would have triggered.
func newT06MissionsAPI(t *testing.T) (huma.API, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)

	authSvc := authStubService()
	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	api := humachi.New(chi.NewMux(), config)
	api.UseMiddleware(withHumaContext)
	api.UseMiddleware(AuthMiddleware(authSvc))
	RegisterMissions(api, missionsSvc)
	return api, mock
}

// TestCrossCuttingNewReadsRejectInvalidClientTimezone is the
// VOC-030-R02 cross-cutting counterpart. Both new reads must
// validate the optional client-supplied IANA timezone and
// reject (400) any non-IANA value. A regression in either
// endpoint's timezone validation would let a malformed client
// value corrupt daily-date-boundary math (a bogus zone could
// shift the local-date window the read returns). The
// cross-cutting aspect is that both reads share the same
// gamification.ResolveSettings validator (the source of truth
// for R02) and must therefore reject identically.
func TestCrossCuttingNewReadsRejectInvalidClientTimezone(t *testing.T) {
	api, mock := newT06MissionsAPI(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	cases := []struct {
		name string
		path string
	}{
		{"P4 daily-mission read", "/api/v1/daily-mission"},
		{"P4 progress read", "/api/v1/progress"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// For a "no stored row" path, the client timezone is
			// the only timezone; the resolver rejects the unknown
			// value before any mission read.
			expectGetUserSettings(mock, userID, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, c.path+"?timezone=Not/A_Real_Zone", nil)
			req = req.WithContext(WithRequester(req.Context(), &auth.User{ID: userID}))
			api.Adapter().ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code,
				"invalid IANA timezone must be rejected (got %d)", w.Code)
		})
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestCrossCuttingNoPathParameterOnNewReads is the structural
// counterpart to VOC-030-TEST-32's cross-user surface: a caller
// cannot enumerate another learner's mission/streak/point data
// by passing a different id parameter, because neither endpoint
// accepts one. The Huma OpenAPI contract is the source of truth
// — if either endpoint were ever to declare a path parameter
// (e.g. /api/v1/daily-mission/{userId}), this test would fail
// immediately, blocking the cross-user exposure at the contract
// level rather than at runtime.
func TestCrossCuttingNoPathParameterOnNewReads(t *testing.T) {
	// The Huma registration must not declare any path parameter on
	// /api/v1/daily-mission or /api/v1/progress. A path parameter
	// would show up in the OpenAPI as a "/{...}" suffix on the
	// path. We assert the bare path exists and the parameterized
	// form does not.
	document, err := encodeContractOpenAPI(NewContractAPI().OpenAPI())
	require.NoError(t, err)
	contract := string(document)

	dailyPath := "/api/v1/daily-mission"
	progressPath := "/api/v1/progress"
	dailyPathParam := dailyPath + "/{"
	progressPathParam := progressPath + "/{"
	assert.NotContains(t, contract, dailyPathParam,
		"%s must not accept a path parameter (cross-user enumeration risk)", dailyPath)
	assert.NotContains(t, contract, progressPathParam,
		"%s must not accept a path parameter (cross-user enumeration risk)", progressPath)
}

// encodeContractOpenAPI marshals an arbitrary value (the huma
// OpenAPI document) to JSON via the standard library. Kept as a
// small helper to keep the test self-contained.
func encodeContractOpenAPI(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// TestCrossCuttingUnauthenticatedRequestsNeverReachRewardPath is
// the VOC-030-TEST-32 cross-cutting counterpart at the
// read-side: an unauthenticated request to either new read
// returns 401, and the sqlmock fixture proves no SQL was issued
// (a regression where the auth gate were skipped would surface
// as an "unexpected SQL" error). The test verifies both reads in
// sequence.
func TestCrossCuttingUnauthenticatedRequestsNeverReachRewardPath(t *testing.T) {
	api, mock := newT06MissionsAPI(t)
	authSvc := authStubService()

	for _, path := range []string{"/api/v1/daily-mission", "/api/v1/progress"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, bytes.NewReader(nil))
			addCSRF(req, authSvc)
			api.Adapter().ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"unauthenticated %s must return 401, not reach a reward path (got %d body=%s)",
				path, w.Code, w.Body.String())
		})
	}
	// Strict-mode sqlmock: any un-staged SQL would have failed
	// the test via ExpectationsWereMet. The unauthenticated 401
	// path must not issue a single SQL statement.
	require.NoError(t, mock.ExpectationsWereMet(),
		"unauthenticated requests must not issue any SQL")
}
