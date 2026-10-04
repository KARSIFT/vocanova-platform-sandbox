package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordlists"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type WordListInput struct {
	ListID string `path:"listId" format:"uuid"`
}
type WordListWriteBody struct {
	Name             string `json:"name" minLength:"1" maxLength:"80"`
	ExpectedRevision int    `json:"expectedRevision" required:"true" minimum:"0"`
}
type WordListRevisionBody struct {
	ExpectedRevision int `json:"expectedRevision" required:"true" minimum:"1"`
}
type PutWordListInput struct {
	ListID         string `path:"listId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           WordListWriteBody
}
type DeleteWordListInput struct {
	ListID           string `path:"listId" format:"uuid"`
	IdempotencyKey   string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	ExpectedRevision int    `query:"expectedRevision" required:"true" minimum:"1"`
}
type PutWordListMemberInput struct {
	ListID         string `path:"listId" format:"uuid"`
	MeaningID      string `path:"meaningId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           WordListRevisionBody
}
type DeleteWordListMemberInput struct {
	ListID           string `path:"listId" format:"uuid"`
	MeaningID        string `path:"meaningId" format:"uuid"`
	IdempotencyKey   string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	ExpectedRevision int    `query:"expectedRevision" required:"true" minimum:"1"`
}
type WordListsOutput struct{ Body wordlists.WordListsResponse }
type WordListOutput struct{ Body wordlists.WordListDetail }

func mapWordListError(err error) error {
	switch {
	case errors.Is(err, wordlists.ErrNotFound):
		return huma.Error404NotFound("List or meaning not found")
	case errors.Is(err, wordlists.ErrInvalid):
		return huma.Error400BadRequest("Invalid list request")
	case errors.Is(err, wordlists.ErrConflict):
		return huma.Error409Conflict("This list changed. Reload it before trying again.")
	case errors.Is(err, wordlists.ErrLimit):
		return huma.Error409Conflict("You have reached the limit of 50 lists or 500 meanings per list.")
	default:
		return huma.Error500InternalServerError("Could not update your list. Try again.")
	}
}
func RegisterWordLists(api huma.API, svc *wordlists.Service, authSvc *auth.Service) {
	read := []func(huma.Context, func(huma.Context)){RequireAuth()}
	write := []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}
	huma.Register(api, huma.Operation{OperationID: "ListWordLists", Method: http.MethodGet, Path: "/api/v1/word-lists", Summary: "List your private named vocabulary collections", Tags: []string{"Word lists"}, Middlewares: read}, func(ctx context.Context, _ *struct{}) (*WordListsOutput, error) {
		st, e := svc.List(ctx, RequesterUserID(ctx))
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &WordListsOutput{*st}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "GetWordList", Method: http.MethodGet, Path: "/api/v1/word-lists/{listId}", Summary: "Get your list and meaning memberships", Tags: []string{"Word lists"}, Middlewares: read}, func(ctx context.Context, in *WordListInput) (*WordListOutput, error) {
		st, e := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.ListID))
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &WordListOutput{*st}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "PutWordList", Method: http.MethodPut, Path: "/api/v1/word-lists/{listId}", Summary: "Create or rename your private list with revision protection", Tags: []string{"Word lists"}, Middlewares: write}, func(ctx context.Context, in *PutWordListInput) (*WordListOutput, error) {
		st, e := svc.Write(ctx, wordlists.WriteRequest{UserID: RequesterUserID(ctx), ListID: parseUUID(in.ListID), Operation: "put", Name: in.Body.Name, ExpectedRevision: in.Body.ExpectedRevision, IdempotencyKey: in.IdempotencyKey})
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &WordListOutput{*st}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "DeleteWordList", Method: http.MethodDelete, Path: "/api/v1/word-lists/{listId}", DefaultStatus: 204, Summary: "Delete your list without changing saved words or practice history", Tags: []string{"Word lists"}, Middlewares: write}, func(ctx context.Context, in *DeleteWordListInput) (*struct{}, error) {
		_, e := svc.Write(ctx, wordlists.WriteRequest{UserID: RequesterUserID(ctx), ListID: parseUUID(in.ListID), Operation: "delete", ExpectedRevision: in.ExpectedRevision, IdempotencyKey: in.IdempotencyKey})
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &struct{}{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "PutWordListMember", Method: http.MethodPut, Path: "/api/v1/word-lists/{listId}/members/{meaningId}", Summary: "Add a canonical meaning to your list", Tags: []string{"Word lists"}, Middlewares: write}, func(ctx context.Context, in *PutWordListMemberInput) (*WordListOutput, error) {
		st, e := svc.Write(ctx, wordlists.WriteRequest{UserID: RequesterUserID(ctx), ListID: parseUUID(in.ListID), MeaningID: parseUUID(in.MeaningID), Operation: "add", ExpectedRevision: in.Body.ExpectedRevision, IdempotencyKey: in.IdempotencyKey})
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &WordListOutput{*st}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "DeleteWordListMember", Method: http.MethodDelete, Path: "/api/v1/word-lists/{listId}/members/{meaningId}", Summary: "Remove a meaning without changing SRS or knowledge", Tags: []string{"Word lists"}, Middlewares: write}, func(ctx context.Context, in *DeleteWordListMemberInput) (*WordListOutput, error) {
		st, e := svc.Write(ctx, wordlists.WriteRequest{UserID: RequesterUserID(ctx), ListID: parseUUID(in.ListID), MeaningID: parseUUID(in.MeaningID), Operation: "remove", ExpectedRevision: in.ExpectedRevision, IdempotencyKey: in.IdempotencyKey})
		if e != nil {
			return nil, mapWordListError(e)
		}
		return &WordListOutput{*st}, nil
	})
}
