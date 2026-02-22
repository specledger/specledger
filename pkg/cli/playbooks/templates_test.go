package playbooks

import (
	"testing"
)

func TestLoadTemplates(t *testing.T) {
	templates, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}

	// Verify we have all 8 templates
	expectedCount := 8
	if len(templates) != expectedCount {
		t.Errorf("LoadTemplates() returned %d templates, want %d", len(templates), expectedCount)
	}

	// Verify expected template IDs exist
	expectedIDs := []string{
		"general-purpose",
		"full-stack",
		"batch-data",
		"realtime-workflow",
		"ml-image",
		"realtime-data",
		"ai-chatbot",
		"adk-chatbot",
	}

	templateMap := make(map[string]bool)
	for _, tmpl := range templates {
		templateMap[tmpl.ID] = true
	}

	for _, id := range expectedIDs {
		if !templateMap[id] {
			t.Errorf("LoadTemplates() missing template with ID %q", id)
		}
	}
}

func TestGetDefaultTemplate(t *testing.T) {
	tmpl, err := GetDefaultTemplate()
	if err != nil {
		t.Fatalf("GetDefaultTemplate() error = %v", err)
	}

	// Verify default template is general-purpose
	if tmpl.ID != "general-purpose" {
		t.Errorf("GetDefaultTemplate() ID = %q, want %q", tmpl.ID, "general-purpose")
	}

	if !tmpl.IsDefault {
		t.Error("GetDefaultTemplate() returned template with IsDefault = false")
	}
}

func TestGetTemplateByID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"general-purpose", false},
		{"full-stack", false},
		{"batch-data", false},
		{"realtime-workflow", false},
		{"ml-image", false},
		{"realtime-data", false},
		{"ai-chatbot", false},
		{"adk-chatbot", false},
		{"nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			tmpl, err := GetTemplateByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateByID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
				return
			}
			if !tt.wantErr && tmpl.ID != tt.id {
				t.Errorf("GetTemplateByID(%q) returned template with ID = %q", tt.id, tmpl.ID)
			}
		})
	}
}

func TestTemplateValidation(t *testing.T) {
	templates, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}

	for _, tmpl := range templates {
		if err := tmpl.Validate(); err != nil {
			t.Errorf("Template %q validation failed: %v", tmpl.ID, err)
		}
	}
}
