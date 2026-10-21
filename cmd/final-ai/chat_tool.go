package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/androticus/final/finalgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

const maxToolRounds = 8

// runChatTurn runs one user-visible chat turn. The client owns the completion loop.
// tools may include Final and other entries from aiopenai.ToolParam; pass nil for no tools.
func runChatTurn(
	ctx context.Context,
	client openai.Client,
	model shared.ChatModel,
	tools []openai.ChatCompletionToolParam,
	ft *finalgo.FinalTool,
	sender finalgo.Sender,
	prior []openai.ChatCompletionMessageParamUnion,
	userText string,
	spy bool,
) (assistantText string, newHistory []openai.ChatCompletionMessageParamUnion, err error) {
	msgs := append(append([]openai.ChatCompletionMessageParamUnion(nil), prior...), openai.UserMessage(userText))

	for round := 0; round < maxToolRounds; round++ {
		params := openai.ChatCompletionNewParams{
			Model:    model,
			Messages: msgs,
		}
		if len(tools) > 0 {
			params.Tools = tools
		}

		resp, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			return "", nil, err
		}
		if len(resp.Choices) == 0 {
			return "", nil, errors.New("chat completion returned zero choices")
		}

		choice := resp.Choices[0]
		msg := choice.Message

		switch choice.FinishReason {
		case "stop", "":
			text := msg.Content
			if text == "" {
				text = msg.Refusal
			}
			if text == "" {
				return "", nil, errors.New("chat completion returned empty content")
			}
			msgs = append(msgs, msg.ToParam())
			return text, msgs, nil

		case "tool_calls":
			if ft == nil || sender == nil {
				return "", nil, errors.New("tool_calls but no Final tool configured")
			}
			if len(msg.ToolCalls) == 0 {
				return "", nil, errors.New("finish_reason tool_calls but no tool_calls")
			}
			msgs = append(msgs, msg.ToParam())
			for _, tc := range msg.ToolCalls {
				content, err := ft.Dispatch(sender, tc.ID, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					return "", nil, err
				}
				if spy {
					spyToolIO(tc.Function.Arguments, content)
				}
				msgs = append(msgs, openai.ToolMessage(content, tc.ID))
			}

		default:
			return "", nil, fmt.Errorf("unexpected finish_reason %q", choice.FinishReason)
		}
	}
	return "", nil, errors.New("tool loop exceeded max rounds")
}

// completeChat is a single completion without tools (non-dev / dumb chat).
func (g *grokClient) completeChat(ctx context.Context, prior []openai.ChatCompletionMessageParamUnion, user string) (reply string, newHistory []openai.ChatCompletionMessageParamUnion, err error) {
	msgs := append(append([]openai.ChatCompletionMessageParamUnion(nil), prior...), openai.UserMessage(user))

	resp, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    g.model,
		Messages: msgs,
	})
	if err != nil {
		return "", nil, err
	}
	if len(resp.Choices) == 0 {
		return "", nil, errors.New("chat completion returned zero choices")
	}
	msg := resp.Choices[0].Message
	reply = msg.Content
	if reply == "" {
		reply = msg.Refusal
	}
	if reply == "" {
		return "", nil, errors.New("chat completion returned empty content")
	}
	newHistory = append(msgs, msg.ToParam())
	return reply, newHistory, nil
}

// spyToolIO prints one Final tool round to stdout (temporary --spy debug).
func spyToolIO(argumentsJSON, sendResponse string) {
	req := argumentsJSON
	var args struct {
		Script string `json:"script"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &args) == nil && args.Script != "" {
		req = args.Script
	}
	fmt.Fprintf(os.Stdout, "req: %s\nres: %s\n\n", req, sendResponse)
}
