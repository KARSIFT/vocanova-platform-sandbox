package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductionAIGenerationRequiresConfiguredRealProvider(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  ProductionConfig
		want bool
	}{
		{"disabled", ProductionConfig{APIProvider: "opencode", APIKey: "test-key"}, false},
		{"missing key", ProductionConfig{AIEnabled: true, APIProvider: "opencode"}, false},
		{"mock is not learner feedback", ProductionConfig{AIEnabled: true, APIProvider: "mock", APIKey: "test-key"}, false},
		{"unknown provider", ProductionConfig{AIEnabled: true, APIProvider: "unknown", APIKey: "test-key"}, false},
		{"opencode", ProductionConfig{AIEnabled: true, APIProvider: "opencode", APIKey: "test-key"}, true},
		{"gemini", ProductionConfig{AIEnabled: true, APIProvider: "gemini", APIKey: "test-key"}, true},
		{"cloudflare missing account", ProductionConfig{AIEnabled: true, APIProvider: "cloudflare", APIKey: "test-key"}, false},
		{"cloudflare", ProductionConfig{AIEnabled: true, APIProvider: "cloudflare", APIKey: "test-key", APIAccountID: "test-account"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, productionAIGenerationEnabled(tc.cfg))
		})
	}
}
