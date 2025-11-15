package common

import (
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func GetExecName() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execName := filepath.Base(execPath)
	return execName, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
