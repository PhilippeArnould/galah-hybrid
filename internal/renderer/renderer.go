// Package renderer loads trusted scenario templates once and escapes dynamic data.
package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/0x4d31/galah/internal/config"
)

type route struct {
	config.RouteConfig
	pattern  *regexp.Regexp
	template *template.Template
}

type Renderer struct{ routes []route }

func New(cfg config.ScenarioConfig) (*Renderer, error) {
	r := &Renderer{}
	if !cfg.Enabled {
		return r, nil
	}
	if cfg.TemplatesDir == "" {
		return nil, fmt.Errorf("scenario templates_dir is required")
	}
	root, err := filepath.EvalSymlinks(cfg.TemplatesDir)
	if err != nil {
		return nil, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	for _, rc := range cfg.Routes {
		if rc.Pattern == "" {
			return nil, fmt.Errorf("scenario route pattern is required")
		}
		pattern, err := regexp.Compile(rc.Pattern)
		if err != nil {
			return nil, fmt.Errorf("scenario pattern %q: %w", rc.Pattern, err)
		}
		if rc.Template == "" || filepath.IsAbs(rc.Template) {
			return nil, fmt.Errorf("invalid template name %q", rc.Template)
		}
		path, err := filepath.EvalSymlinks(filepath.Join(root, rc.Template))
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("template outside templates_dir: %q", rc.Template)
		}
		seen := map[string]bool{}
		for _, field := range rc.Fields {
			if strings.TrimSpace(field) == "" || seen[field] {
				return nil, fmt.Errorf("empty or duplicate scenario field %q", field)
			}
			seen[field] = true
		}
		if len(rc.Fields) == 0 {
			return nil, fmt.Errorf("scenario route %q requires fields", rc.Pattern)
		}
		tmpl, err := template.New(filepath.Base(path)).Option("missingkey=error").ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", rc.Template, err)
		}
		r.routes = append(r.routes, route{rc, pattern, tmpl})
	}
	return r, nil
}

// Match uses only the URL path, excluding query parameters. First match wins.
func (r *Renderer) Match(path string) *config.RouteConfig {
	for i := range r.routes {
		if r.routes[i].pattern.MatchString(path) {
			return &r.routes[i].RouteConfig
		}
	}
	return nil
}

func (r *Renderer) Render(name string, data map[string]string) (string, error) {
	for _, route := range r.routes {
		if route.Template == name {
			var buf bytes.Buffer
			if err := route.template.Execute(&buf, data); err != nil {
				return "", err
			}
			if buf.Len() == 0 {
				return "", fmt.Errorf("empty rendered template %q", name)
			}
			return buf.String(), nil
		}
	}
	return "", fmt.Errorf("unknown template %q", name)
}
