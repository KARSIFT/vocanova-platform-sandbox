package api

import (
	"context"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/wordknowledge"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
	"time"
)

type MeaningKnowledgeDTO struct {
	MeaningID         string     `json:"meaningId" format:"uuid"`
	SelfReportedKnown bool       `json:"selfReportedKnown" doc:"Explicit self-assessment; never review mastery or credit"`
	Note              string     `json:"note" doc:"Private plain-text note belonging only to the requester"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty" format:"date-time"`
}
type GetMeaningKnowledgeInput struct {
	MeaningID string `path:"meaningId" format:"uuid"`
}
type PutMeaningKnowledgeInput struct {
	MeaningID      string `path:"meaningId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           struct {
		SelfReportedKnown bool   `json:"selfReportedKnown" required:"true"`
		Note              string `json:"note" required:"true" maxLength:"2000"`
	}
}
type PatchMeaningKnowledgeInput struct {
	MeaningID      string `path:"meaningId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body           struct {
		SelfReportedKnown bool `json:"selfReportedKnown" required:"true"`
	}
}
type DeleteMeaningKnowledgeInput struct {
	MeaningID      string `path:"meaningId" format:"uuid"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
}
type MeaningKnowledgeOutput struct{ Body MeaningKnowledgeDTO }

func knowledgeDTO(st *wordknowledge.State) *MeaningKnowledgeOutput {
	return &MeaningKnowledgeOutput{Body: MeaningKnowledgeDTO{MeaningID: st.MeaningID.String(), SelfReportedKnown: st.SelfReportedKnown, Note: st.Note, UpdatedAt: st.UpdatedAt}}
}
func mapWordKnowledgeError(err error) error {
	switch {
	case errors.Is(err, wordknowledge.ErrNotFound):
		return huma.Error404NotFound("Meaning not found")
	case errors.Is(err, wordknowledge.ErrConflict):
		return huma.Error409Conflict("Idempotency key already used for a different request")
	case errors.Is(err, wordknowledge.ErrInvalid):
		return huma.Error400BadRequest("Invalid word knowledge request")
	default:
		return huma.Error500InternalServerError("Could not access your word knowledge")
	}
}
func RegisterWordKnowledge(api huma.API, svc *wordknowledge.Service, authSvc *auth.Service) {
	path := "/api/v1/meaning-knowledge/{meaningId}"
	huma.Register(api, huma.Operation{OperationID: "GetMeaningKnowledge", Method: http.MethodGet, Path: path, Summary: "Get your private assessment and note for one meaning", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}}, func(ctx context.Context, in *GetMeaningKnowledgeInput) (*MeaningKnowledgeOutput, error) {
		st, err := svc.Get(ctx, RequesterUserID(ctx), parseUUID(in.MeaningID))
		if err != nil {
			return nil, mapWordKnowledgeError(err)
		}
		return knowledgeDTO(st), nil
	})
	huma.Register(api, huma.Operation{OperationID: "PutMeaningKnowledge", Method: http.MethodPut, Path: path, Summary: "Replace your private assessment and note without changing saved words or review progress", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *PutMeaningKnowledgeInput) (*MeaningKnowledgeOutput, error) {
		st, err := svc.Write(ctx, wordknowledge.WriteRequest{UserID: RequesterUserID(ctx), MeaningID: parseUUID(in.MeaningID), SelfReportedKnown: in.Body.SelfReportedKnown, Note: in.Body.Note, IdempotencyKey: in.IdempotencyKey})
		if err != nil {
			return nil, mapWordKnowledgeError(err)
		}
		return knowledgeDTO(st), nil
	})
	huma.Register(api, huma.Operation{OperationID: "PatchMeaningKnowledge", Method: http.MethodPatch, Path: path, Summary: "Set your self-assessment while preserving your private note", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *PatchMeaningKnowledgeInput) (*MeaningKnowledgeOutput, error) {
		st, err := svc.Write(ctx, wordknowledge.WriteRequest{UserID: RequesterUserID(ctx), MeaningID: parseUUID(in.MeaningID), SelfReportedKnown: in.Body.SelfReportedKnown, IdempotencyKey: in.IdempotencyKey, AssessmentOnly: true})
		if err != nil {
			return nil, mapWordKnowledgeError(err)
		}
		return knowledgeDTO(st), nil
	})
	huma.Register(api, huma.Operation{OperationID: "DeleteMeaningKnowledge", Method: http.MethodDelete, Path: path, DefaultStatus: 204, Summary: "Clear only your private assessment and note", Tags: []string{"Learning"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)}}, func(ctx context.Context, in *DeleteMeaningKnowledgeInput) (*struct{}, error) {
		_, err := svc.Write(ctx, wordknowledge.WriteRequest{UserID: RequesterUserID(ctx), MeaningID: parseUUID(in.MeaningID), IdempotencyKey: in.IdempotencyKey, Delete: true})
		if err != nil {
			return nil, mapWordKnowledgeError(err)
		}
		return &struct{}{}, nil
	})
}
