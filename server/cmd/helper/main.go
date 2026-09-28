package main

import (
	"bufio"
	"fmt"
	"kori/internal/config"
	"kori/internal/utils/crypto"
	"kori/internal/utils/logger"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	var log = logger.New("helper")
	log.Info("🔑 Starting encryption/decryption helper CLI")

	err := godotenv.Load()
	if err != nil {
		log.Error("❌ Failed to load environment variables", err)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Error("❌ Failed to load configuration", err)
		return
	}
	err = crypto.InitializeKeys(cfg.Crypto.PrivateKey)
	if err != nil {
		log.Error("❌ Failed to initialize keys", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter 'e' to encrypt, 'd' to decrypt, or 'q' to quit: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "q" {
			log.Info("👋 Exiting helper CLI")
			break
		}

		fmt.Print("Enter the string to process: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if choice == "e" {
			encrypted, err := crypto.Encrypt(input)
			if err != nil {
				log.Error("❌ Encryption failed", err)
			} else {
				log.Success("✅ Encrypted string: %s", encrypted)
			}
		} else if choice == "d" {
			decrypted, err := crypto.Decrypt(input)
			if err != nil {
				log.Error("❌ Decryption failed", err)
			} else {
				log.Success("✅ Decrypted string: %s", decrypted)
			}
		} else {
			log.Warn("⚠️ Invalid choice. Please enter 'e', 'd', or 'q'.")
		}
	}
}
