package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// EventGridEvent represents the structure of Event Grid events
type EventGridEvent struct {
	ID              string          `json:"id"`
	EventType       string          `json:"eventType"`
	Subject         string          `json:"subject"`
	EventTime       string          `json:"eventTime"`
	Data            json.RawMessage `json:"data"`
	DataVersion     string          `json:"dataVersion"`
	MetadataVersion string          `json:"metadataVersion"`
}

// UserAuditLog represents the structure of user-related audit log data
type UserAuditLog struct {
	Category        string           `json:"category"`
	InitiatedBy     Initiator        `json:"initiatedBy"`
	OperationType   string           `json:"operationType"`
	Result          string           `json:"result"`
	TargetResources []TargetResource `json:"targetResources"`
}

type Initiator struct {
	User User `json:"user"`
}

type User struct {
	UserPrincipalName string `json:"userPrincipalName"`
}

type TargetResource struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	UserPrincipalName string `json:"userPrincipalName"`
}

// GitHubAPIRequest represents the structure of the GitHub API request
type GitHubAPIRequest struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var events []EventGridEvent
	if err := json.Unmarshal(body, &events); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	// 	var events []EventGridEvent
	// if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
	//     http.Error(w, "Error parsing request body", http.StatusBadRequest)
	//     return
	// }

	for _, event := range events {
		var auditLog UserAuditLog
		if err := json.Unmarshal(event.Data, &auditLog); err != nil {
			log.Printf("Error parsing event data: %s", err)
			continue
		}

		// Determine the action based on the operation type
		var action string
		switch auditLog.OperationType {
		case "Add user":
			action = "allocate"
		case "Delete user":
			action = "release"
		default:
			log.Printf("Unhandled operation type: %s", auditLog.OperationType)
			continue
		}

		// Process the user event
		// if len(auditLog.TargetResources) > 0 {
		// 	processUserEvent(auditLog.TargetResources[0].ID, action)
		// }

		for _, target := range auditLog.TargetResources {
			if target.Type == "User" {
				if err := manageGitHubCopilotLicenseOther(target.UserPrincipalName, action); err != nil {
					http.Error(w, fmt.Sprintf("Error managing GitHub Copilot license: %s", err), http.StatusInternalServerError)
					return
				}
			}
		}
	}
}

// func processUserEvent(userID, action string) {
// 	if err := manageGitHubCopilotLicense(userID, action); err != nil {
// 		log.Printf("Error managing GitHub Copilot license: %s", err)
// 	}
// }

func manageGitHubCopilotLicense(userID, action string) error {
	// Replace with the actual GitHub API endpoint and token
	gitHubAPIEndpoint := "https://api.github.com/user/copilot-license"
	gitHubToken := os.Getenv("GITHUB_TOKEN")

	requestBody, err := json.Marshal(GitHubAPIRequest{
		UserID: userID,
		Action: action,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", gitHubAPIEndpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gitHubToken))
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
