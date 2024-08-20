package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GitHubAPIRequestOther struct {
	SelectedUsernames []string `json:"selected_usernames"`
}

func manageGitHubCopilotLicenseOther(username, action string) error {
	gitHubAPIEndpoint := "https://api.github.com/orgs/YOUR_ORG/copilot/billing/selected_users"
	gitHubToken := os.Getenv("GITHUB_API_TOKEN")

	var method string
	if action == "allocate" {
		method = "POST"
	} else if action == "release" {
		method = "DELETE"
	} else {
		return fmt.Errorf("Invalid action: %s", action)
	}

	requestBody, err := json.Marshal(GitHubAPIRequestOther{
		SelectedUsernames: []string{username},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, gitHubAPIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gitHubToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API request failed with status: %s", resp.Status)
	}

	return nil
}
