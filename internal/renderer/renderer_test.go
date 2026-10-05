package renderer

import (
	"github.com/0x4d31/galah/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderer(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "page.html"), []byte("<h1>{{.Title}}</h1>"), 0600)
	cfg := config.ScenarioConfig{Enabled: true, TemplatesDir: dir, Routes: []config.RouteConfig{{Pattern: "^/ticket", Template: "page.html", Fields: []string{"Title"}}, {Pattern: ".*", Template: "page.html", Fields: []string{"Title"}}}}
	r, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Match("/ticket/1").Pattern != "^/ticket" {
		t.Fatal("route ordering")
	}
	body, err := r.Render("page.html", map[string]string{"Title": "<script>alert(1)</script>"})
	if err != nil || strings.Contains(body, "<script>") || !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("escaping: %s %v", body, err)
	}
	if _, err := r.Render("page.html", map[string]string{}); err == nil {
		t.Fatal("expected missing field error")
	}
	for _, name := range []string{"../outside.html", "/etc/passwd", "missing.html"} {
		cfg.Routes[0].Template = name
		if _, err := New(cfg); err == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	cfg.Routes[0].Template = "page.html"
	cfg.Routes[0].Pattern = "["
	if _, err := New(cfg); err == nil {
		t.Fatal("invalid regexp")
	}
	cfg.Enabled = false
	r, err = New(cfg)
	if err != nil || r.Match("/") != nil {
		t.Fatal("disabled scenario")
	}
}
func TestExampleTemplates(t *testing.T) {
	cfg, err := config.LoadConfig("../../scenarios/technova/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Scenario.TemplatesDir = "../../scenarios/technova/templates"
	r, err := New(cfg.Scenario)
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range cfg.Scenario.Routes {
		data := map[string]string{}
		for _, field := range route.Fields {
			data[field] = "example"
		}
		if _, err := r.Render(route.Template, data); err != nil {
			t.Fatal(err)
		}
	}
	if r.Match("/unmapped") != nil {
		t.Fatal("example must preserve unmapped fallback")
	}
}
