package openai

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

// ToolParam builds one OpenAI Chat Completions tools[] entry.
// Call once per tool; the client assembles []ChatCompletionToolParam for the request.
func ToolParam(name, description string, parametersJSON []byte) (openai.ChatCompletionToolParam, error) {
	if name == "" {
		return openai.ChatCompletionToolParam{}, fmt.Errorf("openai: tool name is required")
	}
	var params shared.FunctionParameters
	if len(parametersJSON) > 0 {
		if err := json.Unmarshal(parametersJSON, &params); err != nil {
			return openai.ChatCompletionToolParam{}, fmt.Errorf("openai: tool parameters: %w", err)
		}
	}
	return openai.ChatCompletionToolParam{
		Type: "function",
		Function: shared.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(description),
			Parameters:  params,
		},
	}, nil
}
