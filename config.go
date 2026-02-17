package main

import "os"

func getAPIKey() string {
	return os.Getenv("A21E_API_KEY")
}

func getAPIBaseURL() string {
	u := os.Getenv("A21E_API_URL")
	if u != "" {
		return u
	}
	return "https://api.a21e.com"
}
