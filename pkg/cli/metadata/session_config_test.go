package metadata

import (
	"testing"
)

func TestGetSessionTTLDays(t *testing.T) {
	tests := []struct {
		name string
		meta ProjectMetadata
		want int
	}{
		{
			name: "default when no session config",
			meta: ProjectMetadata{},
			want: 30,
		},
		{
			name: "default when TTL is zero",
			meta: ProjectMetadata{Session: &SessionConfig{TTLDays: 0}},
			want: 30,
		},
		{
			name: "custom TTL",
			meta: ProjectMetadata{Session: &SessionConfig{TTLDays: 90}},
			want: 90,
		},
		{
			name: "7 day TTL",
			meta: ProjectMetadata{Session: &SessionConfig{TTLDays: 7}},
			want: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.meta.GetSessionTTLDays()
			if got != tt.want {
				t.Errorf("GetSessionTTLDays() = %d, want %d", got, tt.want)
			}
		})
	}
}
