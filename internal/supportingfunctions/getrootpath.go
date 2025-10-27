package supportingfunctions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetRootPath основной путь
func GetRootPath(rootDir string) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	beforeStr, _, ok := strings.Cut(currentDir, rootDir)
	if !ok {
		return "", fmt.Errorf("it is impossible to get a prefix from a string '%s'", currentDir)
	}

	return filepath.Join(beforeStr, rootDir), nil
}
