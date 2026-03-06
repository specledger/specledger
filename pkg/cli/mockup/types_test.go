package mockup

import "testing"

func TestFrameworkType_String(t *testing.T) {
	tests := []struct {
		framework FrameworkType
		want      string
	}{
		{FrameworkReact, "React"},
		{FrameworkNextJS, "Next.js"},
		{FrameworkVue, "Vue"},
		{FrameworkNuxt, "Nuxt"},
		{FrameworkSvelte, "Svelte"},
		{FrameworkSvelteKit, "SvelteKit"},
		{FrameworkAngular, "Angular"},
		{FrameworkUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.framework), func(t *testing.T) {
			if got := tt.framework.String(); got != tt.want {
				t.Errorf("FrameworkType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMockupFormat_IsValid(t *testing.T) {
	tests := []struct {
		format MockupFormat
		valid  bool
	}{
		{MockupFormatHTML, true},
		{MockupFormatJSX, true},
		{MockupFormat("svg"), false},
		{MockupFormat(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.IsValid(); got != tt.valid {
				t.Errorf("MockupFormat.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
