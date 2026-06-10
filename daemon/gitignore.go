package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Gitignore struct {
	patterns []gitignorePattern
}

type gitignorePattern struct {
	pattern string
	raw     string
	negate  bool
	dirOnly bool
}

func LoadGitignore(path string) (*Gitignore, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Gitignore{}, nil
		}
		return nil, err
	}
	defer file.Close()

	gi := &Gitignore{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := gitignorePattern{raw: line}

		// Handle negation
		if strings.HasPrefix(line, "!") {
			p.negate = true
			line = line[1:]
		}

		// Handle directory-only patterns (trailing /)
		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		p.pattern = line
		gi.patterns = append(gi.patterns, p)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return gi, nil
}

func (g *Gitignore) Match(relPath string, isDir bool) bool {
	if g == nil {
		return false
	}

	matched := false
	for _, p := range g.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		if matchGitignorePattern(p.pattern, relPath, isDir) {
			matched = !p.negate
		}
	}

	return matched
}

func matchGitignorePattern(pattern, relPath string, isDir bool) bool {
	parts := strings.Split(relPath, string(filepath.Separator))

	// Try matching each suffix of the path (gitignore matches any suffix by default)
	for i := range parts {
		subPath := strings.Join(parts[i:], string(filepath.Separator))

		// Match the full subpath
		if matched, _ := filepath.Match(pattern, subPath); matched {
			return true
		}

		// Match just this component
		if matched, _ := filepath.Match(pattern, parts[i]); matched {
			return true
		}

		// Match pattern as a prefix of a deeper path (e.g., "node_modules" should match "node_modules/package/index.js")
		prefixPattern := pattern + string(filepath.Separator)
		if strings.HasPrefix(subPath, prefixPattern) {
			return true
		}
	}

	return false
}
