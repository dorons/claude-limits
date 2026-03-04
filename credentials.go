package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type credentialsFile struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

func readCredentials() (string, error) {
	if runtime.GOOS == "darwin" {
		token, err := readFromKeychain()
		if err == nil && token != "" {
			return token, nil
		}
	}
	return readFromFile()
}

func keychainServiceName() string {
	base := "Claude Code-credentials"
	configDir := os.Getenv("CLAUDE_CONFIG_DIR")
	if configDir == "" {
		return base
	}
	hash := sha256.Sum256([]byte(configDir))
	suffix := fmt.Sprintf("%x", hash[:4])
	return base + "-" + suffix
}

func readFromKeychain() (string, error) {
	service := keychainServiceName()
	cmd := exec.Command("/usr/bin/security", "find-generic-password", "-s", service, "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain lookup failed: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", fmt.Errorf("keychain returned empty value")
	}
	return extractToken([]byte(raw))
}

func readFromFile() (string, error) {
	dir := os.Getenv("CLAUDE_CONFIG_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".claude")
	}
	path := filepath.Join(dir, ".credentials.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read credentials file %s: %w", path, err)
	}
	return extractToken(data)
}

func extractToken(data []byte) (string, error) {
	var creds credentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("cannot parse credentials JSON: %w", err)
	}
	token := creds.ClaudeAiOauth.AccessToken
	if token == "" {
		return "", fmt.Errorf("no OAuth access token found in credentials")
	}
	return token, nil
}
