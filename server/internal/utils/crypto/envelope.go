package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type envelope struct {
	Version int    `json:"v"`
	Key     []byte `json:"k"`
	Nonce   []byte `json:"n"`
	Data    []byte `json:"d"`
}

// SealSecret uses a per-secret AES key wrapped by the existing installation RSA
// key. AAD prevents moving credentials between tenants or connections.
func SealSecret(plaintext []byte, aad string) (string, error) {
	if PublicKey == nil {
		return "", errors.New("encryption key unavailable")
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	e := envelope{Version: 1, Nonce: make([]byte, gcm.NonceSize())}
	if _, err := rand.Read(e.Nonce); err != nil {
		return "", err
	}
	e.Key, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, PublicKey, key, []byte(aad))
	if err != nil {
		return "", err
	}
	e.Data = gcm.Seal(nil, e.Nonce, plaintext, []byte(aad))
	raw, err := json.Marshal(e)
	return base64.StdEncoding.EncodeToString(raw), err
}

func OpenSecret(encoded, aad string) ([]byte, error) {
	if PrivateKey == nil {
		return nil, errors.New("encryption key unavailable")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var e envelope
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, err
	}
	if e.Version != 1 {
		return nil, errors.New("unsupported secret version")
	}
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, PrivateKey, e.Key, []byte(aad))
	if err != nil {
		return nil, errors.New("unable to open secret")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(e.Nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid secret nonce")
	}
	return gcm.Open(nil, e.Nonce, e.Data, []byte(aad))
}
