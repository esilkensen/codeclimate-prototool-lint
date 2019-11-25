package main

import (
	"fmt"
	"github.com/codeclimate/cc-engine-go/engine"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	rootPath := "/code/"

	config, err := engine.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	analysisFiles, err := protoFileWalk(rootPath, engine.IncludePaths(rootPath, config))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}

	for _, path := range analysisFiles {
		analyzeFile(path)
	}
}

func analyzeFile(path string) {
	cmd := exec.Command("prototool", "lint", path)
	output, err := cmd.CombinedOutput()
	if err != nil && output != nil {
		lines := strings.Split(string(output[:]), "\n")
		for _, line := range lines {
			parseIssue(line)
		}
	}
}

// path:line:column:description
var issuePattern = regexp.MustCompile(`^(.*):([0-9]+):([0-9]+):(.*)$`)

func parseIssue(output string) {
	match := issuePattern.FindStringSubmatch(output)
	if match != nil && len(match) == 5 {
		path := match[1]
		line, _ := strconv.Atoi(match[2])
		column, _ := strconv.Atoi(match[3])
		description := match[4]

		lineColumn := &engine.LineColumn{
			Line:   line,
			Column: column,
		}

		location := &engine.Location{
			Path: path,
			Positions: &engine.LineColumnPosition{
				Begin: lineColumn,
				End:   lineColumn,
			},
		}

		issue := &engine.Issue{
			Type:        "issue",
			Check:       "prototool lint",
			Description: description,
			Categories:  []string{"Style"},
			Location:    location,
		}

		engine.PrintIssue(issue)
	}
}

func protoFileWalk(rootPath string, includePaths []string) (fileList []string, err error) {
	walkFunc := func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".proto") && prefixInArr(path, includePaths) {
			fileList = append(fileList, path)
			return nil
		}
		return err
	}

	err = filepath.Walk(rootPath, walkFunc)

	return fileList, err
}

func prefixInArr(str string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}