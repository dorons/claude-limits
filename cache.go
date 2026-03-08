package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const cacheFileName = ".usage-cache.json"

var defaultCacheTTL = 180 * time.Second

type cachedUsage struct {
	FetchedAt time.Time     `json:"fetched_at"`
	Usage     UsageResponse `json:"usage"`
}

func cacheFilePath() string {
	dir := os.Getenv("CLAUDE_CONFIG_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".claude")
	}
	return filepath.Join(dir, cacheFileName)
}

func cacheTTL() time.Duration {
	if v := os.Getenv("CLAUDE_LIMITS_CACHE_TTL"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultCacheTTL
}

func readCache(path string, ttl time.Duration) (UsageResponse, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return UsageResponse{}, false
	}

	var cached cachedUsage
	if err := json.Unmarshal(data, &cached); err != nil {
		return UsageResponse{}, false
	}

	if time.Since(cached.FetchedAt) > ttl {
		return UsageResponse{}, false
	}

	return cached.Usage, true
}

func writeCache(path string, usage UsageResponse) error {
	cached := cachedUsage{
		FetchedAt: time.Now(),
		Usage:     usage,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".usage-cache-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

func fetchUsageCached(token string) (UsageResponse, error) {
	path := cacheFilePath()
	ttl := cacheTTL()

	if path != "" {
		if usage, ok := readCache(path, ttl); ok {
			return usage, nil
		}
	}

	usage, err := fetchUsage(token)
	if err != nil {
		return UsageResponse{}, err
	}

	if path != "" {
		// Best-effort cache write; don't fail the request on cache errors
		_ = writeCache(path, usage)
	}

	return usage, nil
}
