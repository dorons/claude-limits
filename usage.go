package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	usageURL  = "https://api.anthropic.com/api/oauth/usage"
	betaHeader = "oauth-2025-04-20"
)

type UsageBucket struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

type UsageResponse struct {
	FiveHour *UsageBucket `json:"five_hour"`
	SevenDay *UsageBucket `json:"seven_day"`
}

func fetchUsage(token string) (UsageResponse, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(http.MethodGet, usageURL, nil)
	if err != nil {
		return UsageResponse{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Anthropic-Beta", betaHeader)

	resp, err := client.Do(req)
	if err != nil {
		return UsageResponse{}, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UsageResponse{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return UsageResponse{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var usage UsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return UsageResponse{}, fmt.Errorf("parsing response: %w", err)
	}
	return usage, nil
}
