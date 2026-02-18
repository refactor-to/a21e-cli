package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func getAPIKey() string {
	if k := os.Getenv("A21E_API_KEY"); k != "" {
		return k
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	f, err := os.Open(filepath.Join(home, ".a21e", "credentials"))
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "A21E_API_KEY=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "A21E_API_KEY="))
		}
	}
	return ""
}

func getAPIBaseURL() string {
	u := os.Getenv("A21E_API_URL")
	if u != "" {
		return u
	}
	return "https://api.a21e.com"
}
