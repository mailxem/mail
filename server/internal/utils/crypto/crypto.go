package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	base64_ "kori/internal/utils/base64"
	"kori/internal/utils/logger"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/ssh"
)

var log = logger.New("crypto")

var PrivateKey *rsa.PrivateKey
var PublicKey *rsa.PublicKey

func InitializeKeys(privateKeyEnv string) error {

	log.Info("Initializing keys")

	if privateKeyEnv == "" {
		return errors.New("private key not found")
	}

	privateKeyEnv, err := base64_.DecodeFromBase64(privateKeyEnv)

	if err != nil {
		return fmt.Errorf("failed to decode private key: %w", err)
	}

	key, err := ssh.ParseRawPrivateKey([]byte(privateKeyEnv))

	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	PrivateKey = key.(*rsa.PrivateKey)
	PublicKey = &PrivateKey.PublicKey
	return nil
}

func SignJWT(data string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"data": data,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	signedString, err := token.SignedString(PrivateKey)

	if err != nil {
		return "", err
	}

	return signedString, nil
}

// Encrypt Not used in this snippet but can be used to Encrypt data
func Encrypt(plaintext string) (string, error) {
	if PublicKey == nil {
		return "", errors.New("public key not initialized")
	}

	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		PublicKey,
		[]byte(plaintext),
		nil,
	)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext string) (string, error) {
	if PrivateKey == nil {
		return "", errors.New("private key not initialized")
	}

	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		PrivateKey,
		decodedCiphertext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func ComputeWebhookSignature(requestBody []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(requestBody)
	return hex.EncodeToString(h.Sum(nil))
}
