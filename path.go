package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrNoHome = errors.New("no home found")
)

func Expand(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := Home()
	if err != nil {
		return ""
	}

	return home + path[1:]
}

func Home() (string, error) {
	home := ""

	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}

	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", ErrNoHome
	}
	return home, nil
}

func GetCodeRoot() (string, error) {
	path := os.Getenv("MARKDIR_CODE_ROOT")
	if path != "" {
		return path, nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return pwd, nil
}

func GetContentRoot(contentRoot string) (string, error) {
	path := Expand(contentRoot)
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return path, nil
}
