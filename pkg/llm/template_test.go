package llm_test

import (
	"context"
	"errors"
	"github.com/0x4d31/galah/pkg/llm"
	"github.com/tmc/langchaingo/llms"
	"net/http/httptest"
	"testing"
)

func TestTemplateData(t *testing.T) {
	for _, raw := range []string{`{"Title":"hello"}`, "```json\n{\"Title\":\"hello\"}\n```", `null`, `[]`, `{"Other":"hello"}`, `{"Title":null}`, `{"Title":3}`, `{"Title":"hello","Extra":"x"}`, `{"Title":`} {
		t.Run(raw, func(t *testing.T) {
			model := &MockModel{GenerateContentFunc: func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: raw}}}, nil
			}}
			data, err := llm.GenerateTemplateData(context.Background(), model, 0.6, nil, []string{"Title"})
			valid := raw == `{"Title":"hello"}` || raw == "```json\n{\"Title\":\"hello\"}\n```"
			if valid && (err != nil || data["Title"] != "hello") {
				t.Fatalf("%v %v", data, err)
			}
			if !valid && err == nil {
				t.Fatal("accepted invalid fields")
			}
		})
	}
	for _, response := range []*llms.ContentResponse{nil, {}, {Choices: []*llms.ContentChoice{nil}}} {
		model := &MockModel{GenerateContentFunc: func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
			return response, nil
		}}
		if _, err := llm.GenerateTemplateData(context.Background(), model, 0, nil, []string{"Title"}); err == nil {
			t.Fatal("empty response")
		}
	}
	model := &MockModel{GenerateContentFunc: func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
		return nil, errors.New("offline")
	}}
	if _, err := llm.GenerateTemplateData(context.Background(), model, 0, nil, []string{"Title"}); err == nil {
		t.Fatal("model error")
	}
}
func TestTemplateMessages(t *testing.T) {
	for _, provider := range []string{"ollama", "googleai"} {
		r := httptest.NewRequest("POST", "http://example.com/ticket", nil)
		messages, err := llm.CreateTemplateMessageContent(r, "TechNova", []string{"Title"}, provider)
		if err != nil {
			t.Fatal(err)
		}
		want := 1
		if provider == "ollama" {
			want = 2
		}
		if len(messages) != want {
			t.Fatalf("provider %s: %d messages", provider, len(messages))
		}
	}
}
