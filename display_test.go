package main

import (
	"testing"
	"time"
)

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"zero", 0, "░░░░░░░░░░"},
		{"fifty", 50, "█████░░░░░"},
		{"hundred", 100, "██████████"},
		{"thirty", 30, "███░░░░░░░"},
		{"over hundred", 120, "██████████"},
		{"negative", -5, "░░░░░░░░░░"},
		{"ten", 10, "█░░░░░░░░░"},
		{"ninety nine", 99, "█████████░"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderBar(tt.pct)
			if got != tt.want {
				t.Errorf("renderBar(%v) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0m"},
		{"minutes only", 45 * time.Minute, "45m"},
		{"hours and minutes", 2*time.Hour + 13*time.Minute, "2h 13m"},
		{"days hours minutes", 3*24*time.Hour + 5*time.Hour + 30*time.Minute, "3d 5h 30m"},
		{"exact hours", 3 * time.Hour, "3h 0m"},
		{"exact days", 2 * 24 * time.Hour, "2d 0h 0m"},
		{"negative", -10 * time.Minute, "now"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatResetTime(t *testing.T) {
	now := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		resetsAt string
		want     string
	}{
		{"future 2h13m", "2026-03-04T14:13:00Z", "2h 13m"},
		{"future 3d5h", "2026-03-07T17:00:00Z", "3d 5h 0m"},
		{"past", "2026-03-04T11:00:00Z", "now"},
		{"invalid", "not-a-date", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatResetTime(tt.resetsAt, now)
			if got != tt.want {
				t.Errorf("formatResetTime(%q) = %q, want %q", tt.resetsAt, got, tt.want)
			}
		})
	}
}

func TestToPercent(t *testing.T) {
	tests := []struct {
		name string
		v    float64
		want float64
	}{
		{"fraction", 0.32, 32},
		{"whole number", 47, 47},
		{"zero", 0, 0},
		{"one", 1.0, 100},
		{"over 100", 150, 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPercent(tt.v)
			if got != tt.want {
				t.Errorf("toPercent(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			"valid",
			`{"claudeAiOauth":{"accessToken":"test-token-123"}}`,
			"test-token-123",
			false,
		},
		{
			"empty token",
			`{"claudeAiOauth":{"accessToken":""}}`,
			"",
			true,
		},
		{
			"invalid json",
			`not json`,
			"",
			true,
		},
		{
			"missing field",
			`{"other":"value"}`,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractToken([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("extractToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractToken() = %q, want %q", got, tt.want)
			}
		})
	}
}
