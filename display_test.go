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

func assertJSONBucket(t *testing.T, label string, got, want *JSONBucket) {
	t.Helper()
	if (got == nil) != (want == nil) {
		t.Fatalf("%s nil mismatch: got %v, want %v", label, got, want)
	}
	if got != nil && *got != *want {
		t.Errorf("%s = %+v, want %+v", label, *got, *want)
	}
}

func TestBuildJSONOutput(t *testing.T) {
	now := time.Date(2026, 3, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		usage UsageResponse
		want  JSONOutput
	}{
		{
			name: "pro plan: both buckets",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 0.40, ResetsAt: "2026-03-04T14:13:00Z"},
				SevenDay: &UsageBucket{Utilization: 72, ResetsAt: "2026-03-07T17:00:00Z"},
			},
			want: JSONOutput{
				Session: &JSONBucket{Percent: 40, ResetsAt: "2026-03-04T14:13:00Z", ResetsInSeconds: 7980},
				Weekly:  &JSONBucket{Percent: 72, ResetsAt: "2026-03-07T17:00:00Z", ResetsInSeconds: 277200},
			},
		},
		{
			name: "session only",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 0.50, ResetsAt: "2026-03-04T14:00:00Z"},
			},
			want: JSONOutput{
				Session: &JSONBucket{Percent: 50, ResetsAt: "2026-03-04T14:00:00Z", ResetsInSeconds: 7200},
			},
		},
		{
			name: "weekly only",
			usage: UsageResponse{
				SevenDay: &UsageBucket{Utilization: 10, ResetsAt: "2026-03-05T12:00:00Z"},
			},
			want: JSONOutput{
				Weekly: &JSONBucket{Percent: 10, ResetsAt: "2026-03-05T12:00:00Z", ResetsInSeconds: 86400},
			},
		},
		{
			name:  "neither bucket",
			usage: UsageResponse{},
			want:  JSONOutput{},
		},
		{
			name: "past reset time gives zero seconds",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 0.80, ResetsAt: "2026-03-04T11:00:00Z"},
			},
			want: JSONOutput{
				Session: &JSONBucket{Percent: 80, ResetsAt: "2026-03-04T11:00:00Z", ResetsInSeconds: 0},
			},
		},
		{
			// now=2026-03-04T12:00:00Z; 2026-03-11T08:00:00Z is 6d20h = 590400s away
			// 2026-03-11T09:00:00Z is 6d21h = 594000s away
			name: "max plan: per-model weekly buckets and extra usage",
			usage: UsageResponse{
				FiveHour:       &UsageBucket{Utilization: 13.0, ResetsAt: "2026-03-04T13:00:00Z"},
				SevenDay:       &UsageBucket{Utilization: 2.0, ResetsAt: "2026-03-11T08:00:00Z"},
				SevenDaySonnet: &UsageBucket{Utilization: 5.0, ResetsAt: "2026-03-11T09:00:00Z"},
				ExtraUsage: &ExtraUsageBucket{
					IsEnabled:    true,
					MonthlyLimit: 2000,
					UsedCredits:  374.0,
					Utilization:  18.7,
				},
			},
			want: JSONOutput{
				Session:      &JSONBucket{Percent: 13, ResetsAt: "2026-03-04T13:00:00Z", ResetsInSeconds: 3600},
				Weekly:       &JSONBucket{Percent: 2, ResetsAt: "2026-03-11T08:00:00Z", ResetsInSeconds: 590400},
				WeeklySonnet: &JSONBucket{Percent: 5, ResetsAt: "2026-03-11T09:00:00Z", ResetsInSeconds: 594000},
				ExtraUsage: &JSONExtraUsage{
					IsEnabled:    true,
					MonthlyLimit: 2000,
					UsedCredits:  374.0,
					Percent:      18.7,
				},
			},
		},
		{
			// now=2026-03-04T12:00:00Z; 2026-03-11T08:00:00Z is 590400s away
			name: "max plan: opus weekly bucket",
			usage: UsageResponse{
				SevenDay:     &UsageBucket{Utilization: 5.0, ResetsAt: "2026-03-11T08:00:00Z"},
				SevenDayOpus: &UsageBucket{Utilization: 3.0, ResetsAt: "2026-03-11T08:00:00Z"},
			},
			want: JSONOutput{
				Weekly:     &JSONBucket{Percent: 5, ResetsAt: "2026-03-11T08:00:00Z", ResetsInSeconds: 590400},
				WeeklyOpus: &JSONBucket{Percent: 3, ResetsAt: "2026-03-11T08:00:00Z", ResetsInSeconds: 590400},
			},
		},
		{
			name: "extra usage not enabled is omitted",
			usage: UsageResponse{
				ExtraUsage: &ExtraUsageBucket{IsEnabled: false, MonthlyLimit: 2000, UsedCredits: 0, Utilization: 0},
			},
			want: JSONOutput{
				ExtraUsage: &JSONExtraUsage{IsEnabled: false, MonthlyLimit: 2000, UsedCredits: 0, Percent: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildJSONOutput(tt.usage, now)
			assertJSONBucket(t, "Session", got.Session, tt.want.Session)
			assertJSONBucket(t, "Weekly", got.Weekly, tt.want.Weekly)
			assertJSONBucket(t, "WeeklyOpus", got.WeeklyOpus, tt.want.WeeklyOpus)
			assertJSONBucket(t, "WeeklySonnet", got.WeeklySonnet, tt.want.WeeklySonnet)
			assertJSONBucket(t, "WeeklyOAuth", got.WeeklyOAuth, tt.want.WeeklyOAuth)
			assertJSONBucket(t, "WeeklyCowork", got.WeeklyCowork, tt.want.WeeklyCowork)

			if (got.ExtraUsage == nil) != (tt.want.ExtraUsage == nil) {
				t.Fatalf("ExtraUsage nil mismatch: got %v, want %v", got.ExtraUsage, tt.want.ExtraUsage)
			}
			if got.ExtraUsage != nil && *got.ExtraUsage != *tt.want.ExtraUsage {
				t.Errorf("ExtraUsage = %+v, want %+v", *got.ExtraUsage, *tt.want.ExtraUsage)
			}
		})
	}
}

func TestStatuslineColor(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"0% cyan", 0, colorCyan},
		{"49% cyan", 49, colorCyan},
		{"50% yellow", 50, colorYellow},
		{"79% yellow", 79, colorYellow},
		{"80% magenta", 80, colorMagenta},
		{"100% magenta", 100, colorMagenta},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statuslineColor(tt.pct)
			if got != tt.want {
				t.Errorf("statuslineColor(%v) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestBuildStatusline(t *testing.T) {
	tests := []struct {
		name  string
		usage UsageResponse
		want  string
	}{
		{
			name: "pro plan: only FiveHour and SevenDay",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 0.40},
				SevenDay: &UsageBucket{Utilization: 72},
			},
			want: colorCyan + "5h:40%" + colorReset + " " + colorYellow + "7d:72%" + colorReset,
		},
		{
			name: "max plan with extra usage",
			usage: UsageResponse{
				FiveHour:       &UsageBucket{Utilization: 13},
				SevenDay:       &UsageBucket{Utilization: 2},
				SevenDaySonnet: &UsageBucket{Utilization: 5},
				ExtraUsage: &ExtraUsageBucket{
					IsEnabled:    true,
					UsedCredits:  374,
					MonthlyLimit: 2000,
				},
			},
			want: colorCyan + "5h:13%" + colorReset + " " +
				colorCyan + "7d:2%" + colorReset + " " +
				colorCyan + "Sn:5%" + colorReset + " " +
				colorGreen + "Ex:$3.74/$20" + colorReset,
		},
		{
			name: "extra usage not shown when disabled",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 40},
				ExtraUsage: &ExtraUsageBucket{
					IsEnabled:    false,
					UsedCredits:  100,
					MonthlyLimit: 2000,
				},
			},
			want: colorCyan + "5h:40%" + colorReset,
		},
		{
			name:  "empty response: no crash, empty output",
			usage: UsageResponse{},
			want:  "",
		},
		{
			name: "all buckets shown",
			usage: UsageResponse{
				FiveHour:       &UsageBucket{Utilization: 0.40},
				SevenDay:       &UsageBucket{Utilization: 72},
				SevenDayOpus:   &UsageBucket{Utilization: 3},
				SevenDaySonnet: &UsageBucket{Utilization: 5},
				SevenDayOAuth:  &UsageBucket{Utilization: 2},
				SevenDayCowork: &UsageBucket{Utilization: 10},
			},
			want: colorCyan + "5h:40%" + colorReset + " " +
				colorYellow + "7d:72%" + colorReset + " " +
				colorCyan + "Op:3%" + colorReset + " " +
				colorCyan + "Sn:5%" + colorReset + " " +
				colorCyan + "OA:2%" + colorReset + " " +
				colorCyan + "CW:10%" + colorReset,
		},
		{
			name: "magenta at 80%",
			usage: UsageResponse{
				FiveHour: &UsageBucket{Utilization: 80},
			},
			want: colorMagenta + "5h:80%" + colorReset,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildStatusline(tt.usage)
			if got != tt.want {
				t.Errorf("buildStatusline() = %q, want %q", got, tt.want)
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
