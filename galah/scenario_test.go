package galah

import (
	"context"
	"encoding/json"
	"github.com/0x4d31/galah/internal/config"
	"github.com/0x4d31/galah/pkg/llm"
	"github.com/tmc/langchaingo/llms"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestScenarioPipeline(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ticket.html"), []byte("<h1>{{.Title}}</h1>"), 0600)
	cfg := &config.Config{SystemPrompt: "sys", UserPrompt: "request: %q", Scenario: config.ScenarioConfig{Enabled: true, Name: "TechNova", TemplatesDir: dir, Routes: []config.RouteConfig{{Pattern: "^/ticket", Template: "ticket.html", Fields: []string{"Title"}}}}}
	svc, err := NewServiceFromConfig(context.Background(), cfg, nil, Options{LLMProvider: "openai", LLMAPIKey: "dummy", LLMModel: "test", CacheDBFile: ":memory:", CacheDuration: 1, EventLogFile: filepath.Join(dir, "events.json")})
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()
	calls := 0
	raw := `{"Title":"<script>x</script>"}`
	svc.Model = &MockModel{GenerateContentFunc: func(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
		calls++
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: raw}}}, nil
	}}
	req := httptest.NewRequest("GET", "http://example.com/ticket/12?q=1", nil)
	bytes, err := svc.GenerateHTTPResponse(req, "8888")
	if err != nil {
		t.Fatal(err)
	}
	var response llm.JSONResponse
	if err := json.Unmarshal(bytes, &response); err != nil {
		t.Fatal(err)
	}
	if response.Body != "<h1>&lt;script&gt;x&lt;/script&gt;</h1>" || response.Headers["Content-Type"] != "text/html; charset=UTF-8" {
		t.Fatalf("response: %+v", response)
	}
	cached, err := svc.CheckCache(req, "8888")
	if err != nil || string(cached) != string(bytes) || calls != 1 {
		t.Fatalf("cache: %s %v calls=%d", cached, err, calls)
	}
	raw = `{"headers":{"Content-Type":"text/plain"},"body":"original"}`
	req = httptest.NewRequest("GET", "http://example.com/unmapped", nil)
	bytes, err = svc.GenerateHTTPResponse(req, "8888")
	if err != nil || string(bytes) != raw {
		t.Fatalf("fallback: %s %v", bytes, err)
	}
	raw = `{"Unexpected":"bad"}`
	req = httptest.NewRequest("GET", "http://example.com/ticket/bad", nil)
	if _, err := svc.GenerateHTTPResponse(req, "8888"); err == nil {
		t.Fatal("invalid data accepted")
	}
	if _, err := svc.CheckCache(req, "8888"); err == nil {
		t.Fatal("invalid response cached")
	}
}
