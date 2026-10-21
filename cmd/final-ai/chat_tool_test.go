package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/androticus/final/finalgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

func TestRunChatTurn_toolThenStop(t *testing.T) {
	var callN int
	client := openai.NewClient(
		option.WithBaseURL("http://example.invalid"),
		option.WithAPIKey("test"),
		option.WithMiddleware(func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			callN++
			var body string
			switch callN {
			case 1:
				body = `{
					"id":"1","object":"chat.completion","created":0,"model":"grok-3",
					"choices":[{
						"index":0,
						"message":{
							"role":"assistant",
							"content":null,
							"tool_calls":[{
								"id":"call_abc",
								"type":"function",
								"function":{"name":"final","arguments":"{\"script\":\"3 4 + $\"}"}
							}]
						},
						"finish_reason":"tool_calls"
					}]
				}`
			case 2:
				body = `{
					"id":"2","object":"chat.completion","created":0,"model":"grok-3",
					"choices":[{
						"index":0,
						"message":{"role":"assistant","content":"The answer is 7."},
						"finish_reason":"stop"
					}]
				}`
			default:
				t.Fatalf("unexpected completion call %d", callN)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	)

	ft := finalgo.NewFinalTool()
	snd := &mockFinalSender{response: "7\n"}
	tools := []openai.ChatCompletionToolParam{{
		Type: "function",
		Function: shared.FunctionDefinitionParam{Name: "final"},
	}}

	reply, hist, err := runChatTurn(
		context.Background(),
		client,
		"grok-3",
		tools,
		ft,
		snd,
		nil,
		"What is 3 4 +?",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if reply != "The answer is 7." {
		t.Fatalf("reply = %q", reply)
	}
	if snd.lastScript != "3 4 + $" {
		t.Fatalf("script = %q", snd.lastScript)
	}
	if len(hist) < 4 {
		t.Fatalf("history len = %d, want user+assistant+tool+assistant", len(hist))
	}
}

type mockFinalSender struct {
	lastScript string
	response   string
}

func (m *mockFinalSender) SetSecurityRestrictions(restrictions finalgo.SecurityRestrictions) {
	_ = restrictions
}

func (m *mockFinalSender) Send(script string, timeoutS int) (string, error) {
	m.lastScript = script
	_ = timeoutS
	return m.response, nil
}
