package finalgo

import (
	"encoding/json"
	"fmt"
)

const finalToolName = "final"

const finalToolDescription = `Final is a simple RPN (reverse Polish notation) calculator for basic math.
Pass a "script" string of postfix expressions. Use "$" to print the numeric result to stdout.
Example: script "3 4 + $" prints 7.`

var finalToolParametersJSON = []byte(`{
  "type": "object",
  "properties": {
    "script": {
      "type": "string",
      "description": "Final source to interpret (postfix arithmetic; use $ to emit results)"
    }
  },
  "required": ["script"]
}`)

// FinalTool is the OpenAI-compat tool for running Final in a Session.
// v0: static description and parameters; Dispatch runs script via Sender.
type FinalTool struct{}

// NewFinalTool returns the Final interpret tool (static v0 metadata).
func NewFinalTool() *FinalTool {
	return &FinalTool{}
}

// Name is the tool function name for the provider tools[] entry.
func (ft *FinalTool) Name() string { return finalToolName }

// Description is the tool function description for the provider tools[] entry.
func (ft *FinalTool) Description() string { return finalToolDescription }

// ParametersJSON is a JSON Schema object for the tool function parameters.
func (ft *FinalTool) ParametersJSON() []byte { return finalToolParametersJSON }

type interpretArgs struct {
	Script string `json:"script"`
}

// Dispatch runs one tool call: parse arguments JSON, Send script, return captured stdout.
func (ft *FinalTool) Dispatch(sender Sender, toolCallID, functionName, argumentsJSON string) (content string, err error) {
	_ = toolCallID
	if functionName != ft.Name() {
		return "", fmt.Errorf("finalgo: unexpected tool %q", functionName)
	}
	var args interpretArgs
	if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
		return "", fmt.Errorf("finalgo: tool arguments: %w", err)
	}
	if args.Script == "" {
		return "", fmt.Errorf("finalgo: empty script")
	}
	out, runErr := sender.Send(args.Script, 0)
	if runErr != nil {
		if out != "" {
			return out + "\n[error] " + runErr.Error(), nil
		}
		return "[error] " + runErr.Error(), nil
	}
	return out, nil
}
