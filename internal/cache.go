package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	CacheDirPerms  = 0755
	CacheFilePerms = 0644
)

type cacheMetadata struct {
	URL        string    `json:"url"`
	CachedAt   time.Time `json:"cached_at"`
	Expiration time.Time `json:"expiration"`
}

type Cache struct {
	dir string
	ttl time.Duration
}

func NewCache(dir string, ttl time.Duration) *Cache {
	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

func (c *Cache) LoadFromURL(url string) ([]byte, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(c.dir, CacheDirPerms); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache key from URL
	cacheKey := c.generateKey(url)
	cacheFile := filepath.Join(c.dir, cacheKey+".json")
	metaFile := filepath.Join(c.dir, cacheKey+".meta.json")

	// Check if cached version exists and is valid
	if cachedData, err := c.loadFromCache(cacheFile, metaFile); err == nil {
		log.Printf("Using cached OpenAPI spec from %s\n", cacheFile)
		return cachedData, nil
	}

	// Download from URL
	log.Printf("Downloading OpenAPI spec from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download spec: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Save to cache
	if err := c.saveToCache(cacheFile, metaFile, url, data); err != nil {
		// Log error but continue - cache is optional
		log.Printf("Warning: failed to save to cache: %v", err)
	}

	log.Printf("Downloaded and cached OpenAPI spec from %s", url)

	return data, nil
}

func (c *Cache) generateKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Cache) loadFromCache(cacheFile, metaFile string) ([]byte, error) {
	// Read metadata
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		return nil, err
	}

	var meta cacheMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, err
	}

	// Check if cache is still valid
	if time.Now().After(meta.Expiration) {
		return nil, fmt.Errorf("cache expired")
	}

	// Read cached data
	return os.ReadFile(cacheFile)
}

func (c *Cache) saveToCache(cacheFile, metaFile, url string, data []byte) error {
	// Save data
	if err := os.WriteFile(cacheFile, data, CacheFilePerms); err != nil {
		return err
	}

	// Save metadata
	meta := cacheMetadata{
		URL:        url,
		CachedAt:   time.Now(),
		Expiration: time.Now().Add(c.ttl),
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaFile, metaData, CacheFilePerms)
}
