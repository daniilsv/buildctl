package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	eventToken    string
	eventBackend string
	eventProject string
	eventCommit  string
	eventBranch  string
	eventStatus  string
	eventLog     string
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Send build event",
	Long:  "Send a build event to the backend",
	RunE:  runEvent,
}

func init() {
	eventCmd.Flags().StringVar(&eventToken, "token", "", "Authentication token (required)")
	eventCmd.Flags().StringVar(&eventBackend, "backend", "", "Backend URL (required)")
	eventCmd.Flags().StringVar(&eventProject, "project", "", "Project name (required)")
	eventCmd.Flags().StringVar(&eventCommit, "commit", "", "Commit hash (required)")
	eventCmd.Flags().StringVar(&eventBranch, "branch", "", "Branch name (required)")
	eventCmd.Flags().StringVar(&eventStatus, "status", "", "Build status: queued|started|in_progress|success|failed|cancelled (required)")
	eventCmd.Flags().StringVar(&eventLog, "log", "", "Log message")

	eventCmd.MarkFlagRequired("token")
	eventCmd.MarkFlagRequired("backend")
	eventCmd.MarkFlagRequired("project")
	eventCmd.MarkFlagRequired("commit")
	eventCmd.MarkFlagRequired("branch")
	eventCmd.MarkFlagRequired("status")
}

func runEvent(cmd *cobra.Command, args []string) error {
	validStatuses := map[string]bool{
		"queued":      true,
		"started":     true,
		"in_progress": true,
		"success":     true,
		"failed":      true,
		"cancelled":   true,
	}

	if !validStatuses[eventStatus] {
		return fmt.Errorf("invalid status: %s. Must be one of: queued, started, in_progress, success, failed, cancelled", eventStatus)
	}

	payload := map[string]interface{}{
		"project_name": eventProject,
		"commit_hash":  eventCommit,
		"branch":       eventBranch,
		"status":       eventStatus,
	}

	if eventLog != "" {
		payload["log"] = eventLog
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := eventBackend + "/api/v1/events"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+eventToken)

	client := &http.Client{Timeout: 30 * time.Second}

	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusAccepted {
			break
		}
		if i < 2 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Event sent successfully")
	return nil
}

