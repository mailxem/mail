package base64

import (
	"encoding/base64"
)

// EncodeToBase64 encodes the input string to a base64 string
func EncodeToBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// DecodeFromBase64 decodes the input base64 string to a normal string
func DecodeFromBase64(input string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
