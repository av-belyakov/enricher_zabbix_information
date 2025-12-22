package appversion

import "os"

// GetVersion версия приложения
func GetVersion() (string, error) {
	b, err := os.ReadFile("version")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
