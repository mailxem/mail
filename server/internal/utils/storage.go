package utils

import (
	"io"
	"net/http"
)

type StorageHandler struct{}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}

func (h *StorageHandler) DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
