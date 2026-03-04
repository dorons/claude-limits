package main

import (
	"fmt"
	"strings"
	"time"
)

const barWidth = 10

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

func printUsage(usage UsageResponse) {
	now := time.Now()
	fmt.Println("Claude Usage")
	fmt.Println("─────────────────────────────")

	if usage.FiveHour != nil {
		pct := toPercent(usage.FiveHour.Utilization)
		reset := formatResetTime(usage.FiveHour.ResetsAt, now)
		fmt.Printf("Session (5h)  %s  %3.0f%%  resets in %s\n", renderBar(pct), pct, reset)
	}

	if usage.SevenDay != nil {
		pct := toPercent(usage.SevenDay.Utilization)
		reset := formatResetTime(usage.SevenDay.ResetsAt, now)
		fmt.Printf("Weekly  (7d)  %s  %3.0f%%  resets in %s\n", renderBar(pct), pct, reset)
	}
}
