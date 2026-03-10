package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheFilePath(t *testing.T) {
	t.Run("default uses home .claude dir", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "")
		path := cacheFilePath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".claude", cacheFileName)
		if path != want {
			t.Errorf("cacheFilePath() = %q, want %q", path, want)
		}
	})

	t.Run("respects CLAUDE_CONFIG_DIR", func(t *testing.T) {
		customDir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", customDir)
		path := cacheFilePath()
		want := filepath.Join(customDir, cacheFileName)
		if path != want {
			t.Errorf("cacheFilePath() = %q, want %q", path, want)
		}
	})
}

func TestCacheTTL(t *testing.T) {
	t.Run("default TTL", func(t *testing.T) {
		t.Setenv("CLAUDE_LIMITS_CACHE_TTL", "")
		got := cacheTTL()
		want := 300 * time.Second
		if got != want {
			t.Errorf("cacheTTL() = %v, want %v", got, want)
		}
	})

	t.Run("custom TTL", func(t *testing.T) {
		t.Setenv("CLAUDE_LIMITS_CACHE_TTL", "60")
		got := cacheTTL()
		want := 60 * time.Second
		if got != want {
			t.Errorf("cacheTTL() = %v, want %v", got, want)
		}
	})

	t.Run("invalid TTL falls back to default", func(t *testing.T) {
		t.Setenv("CLAUDE_LIMITS_CACHE_TTL", "not-a-number")
		got := cacheTTL()
		want := 300 * time.Second
		if got != want {
			t.Errorf("cacheTTL() = %v, want %v", got, want)
		}
	})

	t.Run("zero TTL is valid", func(t *testing.T) {
		t.Setenv("CLAUDE_LIMITS_CACHE_TTL", "0")
		got := cacheTTL()
		if got != 0 {
			t.Errorf("cacheTTL() = %v, want 0", got)
		}
	})
}

func testUsageResponse() UsageResponse {
	return UsageResponse{
		FiveHour: &UsageBucket{Utilization: 0.42, ResetsAt: "2026-03-08T12:00:00Z"},
		SevenDay: &UsageBucket{Utilization: 0.15, ResetsAt: "2026-03-14T00:00:00Z"},
	}
}

func TestReadCache(t *testing.T) {
	t.Run("cache miss when file does not exist", func(t *testing.T) {
		_, ok := readCache("/nonexistent/path", time.Hour)
		if ok {
			t.Error("expected cache miss for nonexistent file")
		}
	})

	t.Run("cache hit when file is fresh", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, cacheFileName)

		cached := cachedUsage{
			FetchedAt: time.Now(),
			Usage:     testUsageResponse(),
		}
		data, _ := json.Marshal(cached)
		os.WriteFile(path, data, 0600)

		usage, ok := readCache(path, time.Hour)
		if !ok {
			t.Fatal("expected cache hit")
		}
		if usage.FiveHour.Utilization != 0.42 {
			t.Errorf("utilization = %f, want 0.42", usage.FiveHour.Utilization)
		}
	})

	t.Run("cache miss when file is expired", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, cacheFileName)

		cached := cachedUsage{
			FetchedAt: time.Now().Add(-2 * time.Hour),
			Usage:     testUsageResponse(),
		}
		data, _ := json.Marshal(cached)
		os.WriteFile(path, data, 0600)

		_, ok := readCache(path, time.Hour)
		if ok {
			t.Error("expected cache miss for expired file")
		}
	})

	t.Run("cache miss on corrupt JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, cacheFileName)
		os.WriteFile(path, []byte("{corrupt json!!!"), 0600)

		_, ok := readCache(path, time.Hour)
		if ok {
			t.Error("expected cache miss for corrupt file")
		}
	})
}

func TestFetchUsageCached(t *testing.T) {
	t.Run("force bypasses valid cache", func(t *testing.T) {
		var apiCalls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			resp := testUsageResponse()
			data, _ := json.Marshal(resp)
			w.Write(data)
		}))
		defer srv.Close()

		origURL := usageURL
		usageURL = srv.URL
		defer func() { usageURL = origURL }()

		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)

		// Write a fresh, valid cache entry
		path := filepath.Join(dir, cacheFileName)
		cached := cachedUsage{
			FetchedAt: time.Now(),
			Usage:     testUsageResponse(),
		}
		data, _ := json.Marshal(cached)
		os.WriteFile(path, data, 0600)

		// force=true must bypass the cache and call the API
		_, err := fetchUsageCached("test-token", true)
		if err != nil {
			t.Fatalf("fetchUsageCached() error: %v", err)
		}
		if n := apiCalls.Load(); n != 1 {
			t.Errorf("API called %d times, want 1 (cache should have been bypassed)", n)
		}
	})

	t.Run("no force uses valid cache", func(t *testing.T) {
		var apiCalls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiCalls.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		origURL := usageURL
		usageURL = srv.URL
		defer func() { usageURL = origURL }()

		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)

		path := filepath.Join(dir, cacheFileName)
		cached := cachedUsage{
			FetchedAt: time.Now(),
			Usage:     testUsageResponse(),
		}
		data, _ := json.Marshal(cached)
		os.WriteFile(path, data, 0600)

		_, err := fetchUsageCached("test-token", false)
		if err != nil {
			t.Fatalf("fetchUsageCached() error: %v", err)
		}
		if n := apiCalls.Load(); n != 0 {
			t.Errorf("API called %d times, want 0 (should have used cache)", n)
		}
	})
}

func TestWriteCache(t *testing.T) {
	t.Run("writes and reads back", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, cacheFileName)

		usage := testUsageResponse()
		if err := writeCache(path, usage); err != nil {
			t.Fatalf("writeCache() error: %v", err)
		}

		got, ok := readCache(path, time.Hour)
		if !ok {
			t.Fatal("expected cache hit after write")
		}
		if got.FiveHour.Utilization != usage.FiveHour.Utilization {
			t.Errorf("utilization = %f, want %f", got.FiveHour.Utilization, usage.FiveHour.Utilization)
		}
	})

	t.Run("atomic write does not leave partial files", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, cacheFileName)

		usage := testUsageResponse()
		_ = writeCache(path, usage)

		// Verify no temp files remain
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.Name() != cacheFileName {
				t.Errorf("unexpected file left behind: %s", e.Name())
			}
		}
	})
}
