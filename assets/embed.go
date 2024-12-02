package assets

import (
	"embed"
	"fmt"
	"os"
	"strings"
)

//go:embed public/*
var Public embed.FS

func OutCSSFilename() string {
	filename, err := findFilename("public/css", "out-", ".css")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	return filename
}

func StylesCSSFilename() string {
	filename, err := findFilename("public/css", "styles", ".css")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	return filename
}

func findFilename(dirPath, prefix, suffix string) (string, error) {
	files, err := Public.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read '%s' directory: %w", dirPath, err)
	}

	var matchingFile string

	for _, file := range files {
		filename := file.Name()
		if strings.HasPrefix(filename, prefix) && strings.HasSuffix(filename, suffix) {
			matchingFile = filename
			break
		}
	}

	if matchingFile == "" {
		return "", fmt.Errorf("failed to find file in dir '%s' with prefix %s and suffix %s", dirPath, prefix, suffix)
	}

	return matchingFile, nil
}
