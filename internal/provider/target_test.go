package provider

import "testing"

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input      string
		wantScheme string
		wantID     string
		wantErr    bool
	}{
		{"github:facebook/react", "github", "facebook/react", false},
		{"npm:react", "npm", "react", false},
		{"github:owner/repo", "github", "owner/repo", false},
		{"", "", "", true},
		{"nocolon", "", "", true},
		{":missingscheme", "", "", true},
		{"github:", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tgt, err := ParseTarget(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tgt.Scheme != tt.wantScheme {
				t.Errorf("scheme: got %q, want %q", tgt.Scheme, tt.wantScheme)
			}
			if tgt.Identifier != tt.wantID {
				t.Errorf("identifier: got %q, want %q", tgt.Identifier, tt.wantID)
			}
		})
	}
}
