package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func GetHTMLFromURL(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("template storage returned HTTP %d", resp.StatusCode)
	}
	const limit = 2 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return "", err
	}
	if len(body) > limit {
		return "", fmt.Errorf("template exceeds 2 MB")
	}
	return string(body), nil
}
