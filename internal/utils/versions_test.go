package utils

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		filename string
		want     int
		wantErr  bool
	}{
		{"SEED-Shadows-Awaken-v1.md", 1, false},
		{"WORLD-Overview-v3.md", 3, false},
		{"CHAR-Jax-Tarkin-v12.md", 12, false},
		{"no-version.md", 0, true},
		{"SEED-v.md", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := ParseVersion(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseVersion(%q) = %d, want %d", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		filename string
		want     string
		wantErr  bool
	}{
		{"SEED-Shadows-Awaken-v1.md", "SEED-Shadows-Awaken-v2.md", false},
		{"WORLD-Overview-v3.md", "WORLD-Overview-v4.md", false},
		{"CHAR-Jax-Tarkin-v12.md", "CHAR-Jax-Tarkin-v13.md", false},
		{"no-version.md", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := IncrementVersion(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("IncrementVersion(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestHighestVersion(t *testing.T) {
	files := []string{
		"WORLD-Overview-v1.md",
		"WORLD-Overview-v3.md",
		"WORLD-Overview-v2.md",
	}
	got, err := HighestVersion(files)
	if err != nil {
		t.Fatalf("HighestVersion() error = %v", err)
	}
	if got != "WORLD-Overview-v3.md" {
		t.Errorf("HighestVersion() = %q, want %q", got, "WORLD-Overview-v3.md")
	}

	_, err = HighestVersion(nil)
	if err == nil {
		t.Error("HighestVersion(nil) should return error")
	}

	_, err = HighestVersion([]string{"no-version.md"})
	if err == nil {
		t.Error("HighestVersion with no versioned files should return error")
	}
}
