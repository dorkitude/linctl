package mcpcache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dorkitude/linctl/pkg/api"
)

const (
	cacheDirName  = ".linctl"
	cacheFileName = "mcp-tools-cache.json"
)

type Cache struct {
	FetchedAt time.Time     `json:"fetched_at"`
	Tools     []api.MCPTool `json:"tools"`
}

func GetPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, cacheDirName, cacheFileName), nil
}

func Load() (*Cache, error) {
	path, err := GetPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("mcp cache not found")
		}
		return nil, err
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	if cache.Tools == nil {
		cache.Tools = []api.MCPTool{}
	}
	return &cache, nil
}

func Save(cache Cache) error {
	path, err := GetPath()
	if err != nil {
		return err
	}
	if cache.FetchedAt.IsZero() {
		cache.FetchedAt = time.Now().UTC()
	}
	if cache.Tools == nil {
		cache.Tools = []api.MCPTool{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), strings.TrimSpace(cacheFileName)+".tmp.*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Chmod(0600); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func IsStale(cache *Cache, maxAge time.Duration) bool {
	if cache == nil {
		return true
	}
	if cache.FetchedAt.IsZero() {
		return true
	}
	return time.Since(cache.FetchedAt) > maxAge
}
