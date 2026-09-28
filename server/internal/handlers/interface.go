package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// StorageHandler interface for file operations
type StorageHandler interface {
	UploadFile(ctx context.Context, file []byte, filename string, acl types.ObjectCannedACL, contentType string) (string, error)
	GetSignedURL(ctx context.Context, path string, duration time.Duration) (string, error)
}

var (
	storageHandler StorageHandler
	handlerMu      sync.RWMutex
)

// RegisterStorageHandler sets the storage handler
func RegisterStorageHandler(h StorageHandler) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	storageHandler = h
}

// GetStorageHandler returns the registered storage handler
func GetStorageHandler() StorageHandler {
	handlerMu.RLock()
	defer handlerMu.RUnlock()
	return storageHandler
}
