package models

import (
	"context"
	"sync"
	"time"
)

// FileURLGenerator interface for generating signed URLs
type FileURLGenerator interface {
	GetSignedURL(ctx context.Context, path string, duration time.Duration) (string, error)
}

var (
	urlGenerator FileURLGenerator
	registryMu   sync.RWMutex
)

// RegisterFileURLGenerator sets the URL generator for files
func RegisterFileURLGenerator(generator FileURLGenerator) {
	registryMu.Lock()
	defer registryMu.Unlock()
	urlGenerator = generator
}
