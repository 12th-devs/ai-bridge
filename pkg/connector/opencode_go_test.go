package connector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ai "github.com/beeper/ai-bridge/pkg/ai"
	"github.com/beeper/ai-bridge/pkg/aiid"
)

func TestOpenCodeGoCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("public catalog must not receive API key")
		}
		if r.UserAgent() != "beeper-ai-bridge/1.0" {
			t.Error("missing client identity")
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"kimi-k3","name":"Kimi K3","type":"chat","modalities":{"input":["text","image/png"],"output":["text"]},"limit":{"context":200000,"output":16000},"capabilities":{"tools":{"supported":true}}},{"id":"image-model","type":"image"},{"id":"kimi-k3","type":"chat"}]}`))
	}))
	defer server.Close()
	p := aiid.ProviderConfig{ID: openCodeGoID, Provider: ai.Provider(openCodeGoID), API: ai.ApiOpenAICompletions, BaseURL: openCodeGoBaseURL, APIKey: "private"}
	models, err := openCodeGoModels(context.Background(), p, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ID != "kimi-k3" || models[0].ContextWindow != 200000 || len(models[0].Input) != 2 || models[0].BaseURL != openCodeGoBaseURL || !modelSupportsAgentTools(models[0]) {
		t.Fatalf("unexpected models %#v", models)
	}
	if provider, api := inferProviderRoute(openCodeGoID, ""); provider != p.Provider || api != p.API || !providerAPIAllowed(provider, api) {
		t.Fatal("Go route not supported")
	}
}

func TestOpenCodeGoCatalogErrors(t *testing.T) {
	for _, body := range []string{`{"data":[]}`, `invalid`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
		_, err := openCodeGoModels(context.Background(), aiid.ProviderConfig{}, server.URL)
		server.Close()
		if err == nil {
			t.Fatal("invalid or empty catalog accepted")
		}
	}
}
