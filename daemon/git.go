package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func GetLastCommit(root string) (Commit, error) {
	logPath := filepath.Join(root, ".git", "logs", "HEAD")
	file, err := os.Open(logPath)
	if err != nil {
		return Commit{}, err
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return Commit{}, err
	}

	// Parse git reflog line format:
	// oldSha newSha user <email> timestamp tz\tmessage
	parts := strings.SplitN(lastLine, "\t", 2)
	if len(parts) < 2 {
		return Commit{}, nil
	}

	metaParts := strings.Fields(parts[0])
	if len(metaParts) < 2 {
		return Commit{}, nil
	}

	hash := metaParts[1] // new commit hash
	message := parts[1]

	return Commit{Hash: hash, Message: message}, nil
}
