package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ai "github.com/beeper/ai-bridge/pkg/ai"
	"github.com/beeper/ai-bridge/pkg/aiid"
)

const openCodeGoID = "opencode-go"
const openCodeGoBaseURL = "https://opencode.ai/inference/go/openai/v1"
const openCodeGoModelsURL = "https://opencode.ai/inference/go/v1/models"

func openCodeGoModels(ctx context.Context, provider aiid.ProviderConfig, catalogURL string) ([]ai.Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "beeper-ai-bridge/1.0")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenCode Go catalog unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenCode Go catalog returned HTTP %d", resp.StatusCode)
	}
	var body struct {
		Data []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Type       string `json:"type"`
			Modalities struct {
				Input  []string `json:"input"`
				Output []string `json:"output"`
			} `json:"modalities"`
			Limit struct {
				Context int `json:"context"`
				Output  int `json:"output"`
			} `json:"limit"`
			Capabilities struct {
				Reasoning struct {
					Supported bool     `json:"supported"`
					Effort    []string `json:"effort"`
				} `json:"reasoning"`
				Tools struct {
					Supported bool `json:"supported"`
				} `json:"tools"`
			} `json:"capabilities"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("invalid OpenCode Go catalog")
	}
	models := []ai.Model{}
	seen := map[string]bool{}
	for _, item := range body.Data {
		if item.ID == "" || seen[item.ID] || item.Type != "chat" {
			continue
		}
		seen[item.ID] = true
		input := []string{"text"}
		for _, modality := range item.Modalities.Input {
			if strings.HasPrefix(modality, "image/") {
				input = append(input, "image")
				break
			}
		}
		levels := map[ai.ModelThinkingLevel]*string{}
		for _, level := range []ai.ModelThinkingLevel{"off", "minimal", "low", "medium", "high", "xhigh"} {
			levels[level] = nil
		}
		var defaultLevel ai.ModelThinkingLevel
		for _, effort := range item.Capabilities.Reasoning.Effort {
			level := ai.ModelThinkingLevel(effort)
			if _, ok := levels[level]; ok {
				value := effort
				levels[level] = &value
				if defaultLevel == "" {
					defaultLevel = level
				}
			}
		}
		models = append(models, normalizeProviderModel(ai.Model{
			ID: item.ID, Name: item.Name, Input: input, Output: item.Modalities.Output,
			Reasoning: item.Capabilities.Reasoning.Supported, ThinkingLevelMap: levels, DefaultThinkingLevel: defaultLevel,
			ContextWindow: item.Limit.Context, MaxTokens: item.Limit.Output,
			Compat: map[string]any{"tools_supported": item.Capabilities.Tools.Supported, "supportsStore": false, "maxTokensField": "max_tokens", "supportsDeveloperRole": false, "supportsReasoningEffort": defaultLevel != ""},
		}, provider))
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("OpenCode Go catalog has no chat models")
	}
	return models, nil
}
