package providers

import (
	ai "github.com/beeper/ai-bridge/pkg/ai"
	"testing"
)

func TestOpenCodeGoSessionHeaders(t *testing.T) {
	model := ai.Model{Provider: ai.Provider("opencode-go"), API: ai.ApiOpenAICompletions}
	for _, session := range []string{"conversation-1", "conversation-2"} {
		cfg, err := buildOpenAIClientConfig(model, ai.Context{}, ai.StreamOptions{APIKey: "test", SessionID: session, CacheRetention: ai.CacheRetentionNone})
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Headers["x-opencode-session"] != session || cfg.Headers["User-Agent"] != "beeper-ai-bridge/1.0" {
			t.Fatalf("missing session headers: %#v", cfg.Headers)
		}
	}
	if _, err := buildOpenAIClientConfig(model, ai.Context{}, ai.StreamOptions{}); err == nil {
		t.Fatal("sessionless request accepted")
	}
}
