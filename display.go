package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const barWidth = 10

const (
	colorReset   = "\033[0m"
	colorCyan    = "\033[36m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorGreen   = "\033[32m"
)

// toPercent normalises the API's utilization value to a 0–100 percentage.
// The API may return either a fraction (0.32) or a whole number (32).
func toPercent(v float64) float64 {
	if v > 0 && v <= 1.0 {
		return v * 100
	}
	return v
}

func renderBar(pct float64) string {
	filled := int(pct / 100.0 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
}

func formatDurationCompact(d time.Duration) string {
	if d < 0 {
		return "now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if days > 0 || hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	parts = append(parts, fmt.Sprintf("%dm", minutes))
	return strings.Join(parts, "")
}

func formatDurationCompactHours(d time.Duration) string {
	if d < 0 {
		return "now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	parts = append(parts, fmt.Sprintf("%dh", hours))
	return strings.Join(parts, "")
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if days > 0 || hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	parts = append(parts, fmt.Sprintf("%dm", minutes))
	return strings.Join(parts, " ")
}

func formatDurationHours(d time.Duration) string {
	if d < 0 {
		return "now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	parts = append(parts, fmt.Sprintf("%dh", hours))
	return strings.Join(parts, " ")
}

func formatResetTime(resetsAt string, now time.Time) string {
	t, err := time.Parse(time.RFC3339, resetsAt)
	if err != nil {
		return "unknown"
	}
	d := t.Sub(now)
	if d < 0 {
		return "now"
	}
	return formatDuration(d)
}

func formatResetTimeHours(resetsAt string, now time.Time) string {
	t, err := time.Parse(time.RFC3339, resetsAt)
	if err != nil {
		return "unknown"
	}
	d := t.Sub(now)
	if d < 0 {
		return "now"
	}
	return formatDurationHours(d)
}

// JSONBucket represents a single usage bucket in JSON output.
type JSONBucket struct {
	Percent         float64 `json:"percent"`
	ResetsAt        string  `json:"resets_at"`
	ResetsInSeconds int     `json:"resets_in_seconds"`
}

// JSONExtraUsage represents the extra usage (overage) bucket in JSON output.
type JSONExtraUsage struct {
	IsEnabled    bool    `json:"is_enabled"`
	MonthlyLimit float64 `json:"monthly_limit"`
	UsedCredits  float64 `json:"used_credits"`
	Percent      float64 `json:"percent"`
}

// JSONOutput is the top-level JSON output structure.
type JSONOutput struct {
	Session        *JSONBucket     `json:"session,omitempty"`
	Weekly         *JSONBucket     `json:"weekly,omitempty"`
	WeeklyOpus     *JSONBucket     `json:"weekly_opus,omitempty"`
	WeeklySonnet   *JSONBucket     `json:"weekly_sonnet,omitempty"`
	WeeklyOAuth    *JSONBucket     `json:"weekly_oauth_apps,omitempty"`
	WeeklyCowork   *JSONBucket     `json:"weekly_cowork,omitempty"`
	ExtraUsage     *JSONExtraUsage `json:"extra_usage,omitempty"`
}

func buildJSONBucket(bucket *UsageBucket, now time.Time) *JSONBucket {
	if bucket == nil {
		return nil
	}
	pct := toPercent(bucket.Utilization)
	var seconds int
	t, err := time.Parse(time.RFC3339, bucket.ResetsAt)
	if err == nil {
		d := t.Sub(now)
		if d > 0 {
			seconds = int(d.Seconds())
		}
	}
	return &JSONBucket{
		Percent:         pct,
		ResetsAt:        bucket.ResetsAt,
		ResetsInSeconds: seconds,
	}
}

func buildJSONExtraUsage(e *ExtraUsageBucket) *JSONExtraUsage {
	if e == nil {
		return nil
	}
	return &JSONExtraUsage{
		IsEnabled:    e.IsEnabled,
		MonthlyLimit: e.MonthlyLimit,
		UsedCredits:  e.UsedCredits,
		Percent:      toPercent(e.Utilization),
	}
}

func buildJSONOutput(usage UsageResponse, now time.Time) JSONOutput {
	return JSONOutput{
		Session:      buildJSONBucket(usage.FiveHour, now),
		Weekly:       buildJSONBucket(usage.SevenDay, now),
		WeeklyOpus:   buildJSONBucket(usage.SevenDayOpus, now),
		WeeklySonnet: buildJSONBucket(usage.SevenDaySonnet, now),
		WeeklyOAuth:  buildJSONBucket(usage.SevenDayOAuth, now),
		WeeklyCowork: buildJSONBucket(usage.SevenDayCowork, now),
		ExtraUsage:   buildJSONExtraUsage(usage.ExtraUsage),
	}
}

func printUsageJSON(usage UsageResponse) {
	output := buildJSONOutput(usage, time.Now())
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printBucketRow(label string, bucket *UsageBucket, now time.Time, hourOnly bool) {
	pct := toPercent(bucket.Utilization)
	var reset string
	if hourOnly {
		reset = formatResetTimeHours(bucket.ResetsAt, now)
	} else {
		reset = formatResetTime(bucket.ResetsAt, now)
	}
	fmt.Printf("%-16s  %s  %3.0f%%  resets in %s\n", label, renderBar(pct), pct, reset)
}

func statuslineColor(pct float64) string {
	switch {
	case pct >= 80:
		return colorMagenta
	case pct >= 50:
		return colorYellow
	default:
		return colorGreen
	}
}

func formatStatuslineBucket(label string, bucket *UsageBucket, now time.Time, showReset bool, hourOnly bool) string {
	pct := toPercent(bucket.Utilization)
	color := statuslineColor(pct)
	if showReset {
		t, err := time.Parse(time.RFC3339, bucket.ResetsAt)
		var reset string
		if err != nil {
			reset = "unknown"
		} else if hourOnly {
			reset = formatDurationCompactHours(t.Sub(now))
		} else {
			reset = formatDurationCompact(t.Sub(now))
		}
		return fmt.Sprintf("%s%s:%.0f%% (%s)%s", color, label, pct, reset, colorReset)
	}
	return fmt.Sprintf("%s%s:%.0f%%%s", color, label, pct, colorReset)
}

func buildStatusline(usage UsageResponse, now time.Time) string {
	var parts []string

	buckets := []struct {
		label     string
		bucket    *UsageBucket
		showReset bool
		hourOnly  bool
	}{
		{"5h", usage.FiveHour, true, false},
		{"7d", usage.SevenDay, true, true},
		{"Op", usage.SevenDayOpus, false, false},
		{"Sn", usage.SevenDaySonnet, false, false},
		{"OA", usage.SevenDayOAuth, false, false},
		{"CW", usage.SevenDayCowork, false, false},
	}

	for _, b := range buckets {
		if b.bucket != nil {
			parts = append(parts, formatStatuslineBucket(b.label, b.bucket, now, b.showReset, b.hourOnly))
		}
	}

	if usage.ExtraUsage != nil && usage.ExtraUsage.IsEnabled {
		e := usage.ExtraUsage
		parts = append(parts, fmt.Sprintf("%sEx:$%.2f/$%.0f%s",
			colorGreen, e.UsedCredits/100, e.MonthlyLimit/100, colorReset))
	}

	return strings.Join(parts, " ")
}

func printStatusline(usage UsageResponse) {
	fmt.Println(buildStatusline(usage, time.Now()))
}

func printUsage(usage UsageResponse) {
	now := time.Now()
	fmt.Println("Claude Usage")
	fmt.Println("─────────────────────────────")

	if usage.FiveHour != nil {
		printBucketRow("Session (5h)", usage.FiveHour, now, false)
	}

	if usage.SevenDay != nil {
		printBucketRow("Weekly  (7d)", usage.SevenDay, now, true)
	}

	if usage.SevenDayOpus != nil {
		printBucketRow("  Opus only", usage.SevenDayOpus, now, true)
	}

	if usage.SevenDaySonnet != nil {
		printBucketRow("  Sonnet only", usage.SevenDaySonnet, now, true)
	}

	if usage.SevenDayOAuth != nil {
		printBucketRow("  OAuth apps", usage.SevenDayOAuth, now, true)
	}

	if usage.SevenDayCowork != nil {
		printBucketRow("  Cowork", usage.SevenDayCowork, now, true)
	}

	if usage.ExtraUsage != nil && usage.ExtraUsage.IsEnabled {
		pct := toPercent(usage.ExtraUsage.Utilization)
		fmt.Printf("%-16s  %s  %3.0f%%  $%.0f / $%.0f\n",
			"Extra usage",
			renderBar(pct),
			pct,
			usage.ExtraUsage.UsedCredits,
			usage.ExtraUsage.MonthlyLimit,
		)
	}
}
