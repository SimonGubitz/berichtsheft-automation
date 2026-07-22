package keychain

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

func SetPassword(username, password string) error {
	err := keyring.Set("berichtsheft", username, password)
	if err != nil {
		return fmt.Errorf("failed to save password in keyring. %w", err)
	}
	return nil
}

func GetPassword(username string) (string, error) {
	secret, err := keyring.Get("berichtsheft", username)
	if err != nil {
		return "", fmt.Errorf("failed to get password from keyring. %w", err)
	}
	return secret, nil
}
