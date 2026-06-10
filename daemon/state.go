package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Commit struct {
	Hash    string `json:"hash"`
	Message string `json:"message,omitempty"`
}

type StateResponse struct {
	RawContext        string  `json:"raw_context,omitempty"`
	SummarizedContext string  `json:"summarized_context,omitempty"`
	FileTree          string  `json:"file_tree"`
	LastCommit        Commit  `json:"last_commit"`
}

type summaryCache struct {
	mdModTime time.Time
	text      string
}

type StateManager struct {
	mu              sync.RWMutex
	root            string
	mdPath          string
	ollamaEndpoint  string
	ollamaModel     string
	threshold       int
	mdContent       string
	fileTree        string
	lastCommit      Commit
	mdModTime       time.Time
	cache           *summaryCache
}

func NewState(root, ollamaEndpoint, ollamaModel string, threshold int) *StateManager {
	return &StateManager{
		root:           root,
		mdPath:         filepath.Join(root, ".echo-state.md"),
		ollamaEndpoint: ollamaEndpoint,
		ollamaModel:    ollamaModel,
		threshold:      threshold,
	}
}

func (s *StateManager) Refresh() error {
	var mdContent, summary string
	var mdModTime time.Time
	var lastCommit Commit

	// 1. Read .echo-state.md
	if info, err := os.Stat(s.mdPath); err == nil {
		mdModTime = info.ModTime()
		content, err := os.ReadFile(s.mdPath)
		if err != nil {
			return err
		}
		mdContent = string(content)
	}

	// 2. Summarize if needed
	if len(mdContent) > s.threshold {
		if s.cache != nil && s.cache.mdModTime.Equal(mdModTime) {
			summary = s.cache.text
		} else {
			var err error
			summary, err = summarize(mdContent, s.ollamaEndpoint, s.ollamaModel)
			if err != nil {
				// Fallback to raw if Ollama fails
				summary = ""
			}
			if summary != "" {
				s.cache = &summaryCache{mdModTime: mdModTime, text: summary}
			}
		}
	}

	// 3. Build file tree respecting .gitignore
	tree, err := BuildFileTree(s.root)
	if err != nil {
		// Gracefully continue without tree on error
		tree = ""
	}

	// 4. Get last git commit
	commit, err := GetLastCommit(s.root)
	if err == nil {
		lastCommit = commit
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.mdContent = mdContent
	s.mdModTime = mdModTime
	s.fileTree = tree
	s.lastCommit = lastCommit
	if summary != "" {
		s.cache = &summaryCache{mdModTime: mdModTime, text: summary}
	}

	return nil
}

func (s *StateManager) Get() StateResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := StateResponse{
		FileTree:   s.fileTree,
		LastCommit: s.lastCommit,
	}

	if s.cache != nil && !s.mdModTime.IsZero() && s.cache.mdModTime.Equal(s.mdModTime) && len(s.mdContent) > s.threshold {
		resp.SummarizedContext = s.cache.text
	} else {
		resp.RawContext = s.mdContent
	}

	return resp
}
