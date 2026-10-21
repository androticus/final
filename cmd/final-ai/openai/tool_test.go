package openai

import (
	"encoding/json"
	"testing"
)

func TestToolParam_finalCalculator(t *testing.T) {
	params := []byte(`{
		"type": "object",
		"properties": {
			"script": { "type": "string" }
		},
		"required": ["script"]
	}`)
	tp, err := ToolParam("final", "RPN calculator", params)
	if err != nil {
		t.Fatal(err)
	}
	if tp.Type != "function" {
		t.Fatalf("type = %q", tp.Type)
	}
	if tp.Function.Name != "final" {
		t.Fatalf("name = %q", tp.Function.Name)
	}
	if !tp.Function.Description.Valid() || tp.Function.Description.Value != "RPN calculator" {
		t.Fatalf("description = %v", tp.Function.Description)
	}
	raw, err := json.Marshal(tp)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	fn, ok := wire["function"].(map[string]any)
	if !ok {
		t.Fatalf("wire = %s", raw)
	}
	if fn["name"] != "final" {
		t.Fatalf("wire name = %v", fn["name"])
	}
}

func TestToolParam_emptyName(t *testing.T) {
	_, err := ToolParam("", "desc", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestToolParam_badJSON(t *testing.T) {
	_, err := ToolParam("x", "desc", []byte(`{`))
	if err == nil {
		t.Fatal("expected error")
	}
}
