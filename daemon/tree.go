package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type treeNode struct {
	name     string
	isDir    bool
	children []*treeNode
}

func BuildFileTree(root string) (string, error) {
	gitignore, err := LoadGitignore(filepath.Join(root, ".gitignore"))
	if err != nil {
		return "", err
	}

	node, err := buildTreeNode(root, "", gitignore)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(".")
	for i, child := range node.children {
		formatNode(child, "", i == len(node.children)-1, &sb)
	}
	return sb.String(), nil
}

func buildTreeNode(dir string, rel string, gitignore *Gitignore) (*treeNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	node := &treeNode{name: filepath.Base(dir), isDir: true}

	for _, entry := range entries {
		name := entry.Name()
		childRel := filepath.Join(rel, name)

		// Skip .git directory
		if name == ".git" && entry.IsDir() {
			continue
		}

		isDir := entry.IsDir()

		// Check .gitignore if available
		if gitignore != nil && gitignore.Match(childRel, isDir) {
			continue
		}

		child := &treeNode{name: name, isDir: isDir}
		if isDir {
			childPath := filepath.Join(dir, name)
			childNode, err := buildTreeNode(childPath, childRel, gitignore)
			if err == nil {
				child.children = childNode.children
			}
		}
		node.children = append(node.children, child)
	}

	// Sort: directories first, then alphabetically
	sort.SliceStable(node.children, func(i, j int) bool {
		if node.children[i].isDir != node.children[j].isDir {
			return node.children[i].isDir
		}
		return node.children[i].name < node.children[j].name
	})

	return node, nil
}

func formatNode(node *treeNode, prefix string, isLast bool, sb *strings.Builder) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	sb.WriteString(fmt.Sprintf("\n%s%s%s", prefix, connector, node.name))

	if !node.isDir {
		return
	}

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range node.children {
		formatNode(child, childPrefix, i == len(node.children)-1, sb)
	}
}
