package api

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humachi"
	"github.com/go-chi/chi/v5"
)

func NewContractAPI() huma.API {
	config := huma.DefaultConfig("Vocanova API", "0.1.0")
	config.Info.Description = "Explicit Vocanova HTTP DTO contract. Internal persistence models are not exposed."
	contractAPI := humachi.New(chi.NewMux(), config)
	RegisterContract(contractAPI)
	return contractAPI
}
