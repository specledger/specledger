package playbooks

import (
	"testing"
)

func TestMergeSentinelSection(t *testing.T) {
	managed := "specledger/*/issues.jsonl linguist-generated=true\nspecledger/*/tasks.md linguist-generated=true"

	expectedBlock := SentinelBegin + "\n" +
		SentinelComment + "\n" +
		managed + "\n" +
		SentinelEnd

	tests := []struct {
		name     string
		existing string
		managed  string
		want     string
	}{
		{
			name:     "empty existing content",
			existing: "",
			managed:  managed,
			want:     expectedBlock + "\n",
		},
		{
			name:     "whitespace-only existing content",
			existing: "   \n\n  ",
			managed:  managed,
			want:     expectedBlock + "\n",
		},
		{
			name:     "append to existing without sentinels",
			existing: "*.pbxproj binary\n",
			managed:  managed,
			want:     "*.pbxproj binary\n\n" + expectedBlock + "\n",
		},
		{
			name:     "append to existing with trailing newlines",
			existing: "*.pbxproj binary\n\n\n",
			managed:  managed,
			want:     "*.pbxproj binary\n\n" + expectedBlock + "\n",
		},
		{
			name: "replace existing sentinel section",
			existing: "*.pbxproj binary\n\n" +
				SentinelBegin + "\n" +
				SentinelComment + "\n" +
				"old content\n" +
				SentinelEnd + "\n",
			managed: managed,
			want:    "*.pbxproj binary\n\n" + expectedBlock + "\n",
		},
		{
			name: "preserve content after sentinel section",
			existing: "*.pbxproj binary\n\n" +
				SentinelBegin + "\n" +
				SentinelComment + "\n" +
				"old content\n" +
				SentinelEnd + "\n" +
				"\n# Custom rules\n*.log text\n",
			managed: managed,
			want: "*.pbxproj binary\n\n" + expectedBlock + "\n\n" +
				"# Custom rules\n*.log text\n",
		},
		{
			name: "malformed sentinel - begin without end",
			existing: "*.pbxproj binary\n\n" +
				SentinelBegin + "\n" +
				"orphaned content\n" +
				"more orphaned\n",
			managed: managed,
			want:    "*.pbxproj binary\n\n" + expectedBlock + "\n",
		},
		{
			name:     "idempotency - merge same content twice",
			existing: "*.pbxproj binary\n\n" + expectedBlock + "\n",
			managed:  managed,
			want:     "*.pbxproj binary\n\n" + expectedBlock + "\n",
		},
		{
			name:     "empty managed content",
			existing: "*.pbxproj binary\n",
			managed:  "",
			want: "*.pbxproj binary\n\n" +
				SentinelBegin + "\n" +
				SentinelComment + "\n" +
				"\n" +
				SentinelEnd + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeSentinelSection(tt.existing, tt.managed)
			if got != tt.want {
				t.Errorf("MergeSentinelSection() mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestMergeSentinelSection_Idempotency(t *testing.T) {
	managed := "specledger/*/issues.jsonl linguist-generated=true\nspecledger/*/tasks.md linguist-generated=true"
	existing := "*.pbxproj binary\n# my custom rules\n*.log text\n"

	// First merge
	first := MergeSentinelSection(existing, managed)
	// Second merge with output of first
	second := MergeSentinelSection(first, managed)
	// Third merge
	third := MergeSentinelSection(second, managed)

	if first != second {
		t.Errorf("Not idempotent: first != second\nfirst:\n%q\nsecond:\n%q", first, second)
	}
	if second != third {
		t.Errorf("Not idempotent: second != third\nsecond:\n%q\nthird:\n%q", second, third)
	}
}
