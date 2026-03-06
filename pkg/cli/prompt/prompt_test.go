package prompt

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		data    any
		want    string
		wantErr bool
	}{
		{
			name: "simple template",
			tmpl: "Hello, {{.Name}}!",
			data: struct{ Name string }{Name: "World"},
			want: "Hello, World!",
		},
		{
			name: "template with multiple fields",
			tmpl: "{{.Title}} - {{.Format}}",
			data: struct {
				Title  string
				Format string
			}{Title: "Mockup", Format: "html"},
			want: "Mockup - html",
		},
		{
			name:    "invalid template syntax",
			tmpl:    "{{.Invalid",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderTemplate("test", tt.tmpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		min   int
		max   int
	}{
		{"empty string", "", 0, 0},
		{"short text", "hello", 1, 3},
		{"medium text", strings.Repeat("a", 350), 90, 110},
		{"long text", strings.Repeat("word ", 1000), 1300, 1600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got < tt.min || got > tt.max {
				t.Errorf("EstimateTokens() = %d, want between %d and %d", got, tt.min, tt.max)
			}
		})
	}
}
