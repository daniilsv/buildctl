package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	artifactToken    string
	artifactBackend  string
	artifactProject  string
	artifactBranch   string
	artifactCommit   string
	artifactFile     string
)

var artifactCmd = &cobra.Command{
	Use:   "artifact upload",
	Short: "Upload build artifact",
	Long:  "Upload a build artifact to S3 via presigned URL",
	RunE:  runArtifactUpload,
}

func init() {
	artifactCmd.Flags().StringVar(&artifactToken, "token", "", "Authentication token (required)")
	artifactCmd.Flags().StringVar(&artifactBackend, "backend", "", "Backend URL (required)")
	artifactCmd.Flags().StringVar(&artifactProject, "project", "", "Project name (required)")
	artifactCmd.Flags().StringVar(&artifactBranch, "branch", "", "Branch name (required)")
	artifactCmd.Flags().StringVar(&artifactCommit, "commit", "", "Commit hash (required)")
	artifactCmd.Flags().StringVar(&artifactFile, "file", "", "File path to upload (required)")

	artifactCmd.MarkFlagRequired("token")
	artifactCmd.MarkFlagRequired("backend")
	artifactCmd.MarkFlagRequired("project")
	artifactCmd.MarkFlagRequired("branch")
	artifactCmd.MarkFlagRequired("commit")
	artifactCmd.MarkFlagRequired("file")
}

func runArtifactUpload(cmd *cobra.Command, args []string) error {
	file, err := os.Open(artifactFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	filename := fileInfo.Name()

	presignReq := map[string]string{
		"project_name": artifactProject,
		"branch_name":  artifactBranch,
		"commit_hash":  artifactCommit,
		"filename":     filename,
	}

	jsonPayload, err := json.Marshal(presignReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := artifactBackend + "/api/v1/artifacts/presign"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+artifactToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("presign request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var presignResp struct {
		UploadURL string `json:"upload_url"`
		S3Key     string `json:"s3_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&presignResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	file.Seek(0, 0)
	uploadReq, err := http.NewRequest("PUT", presignResp.UploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	uploadReq.ContentLength = fileInfo.Size()
	uploadReq.Header.Set("Content-Type", "application/octet-stream")

	uploadClient := &http.Client{Timeout: 5 * time.Minute}
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(body))
	}

	// Подтверждение загрузки на backend
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	confirmReq := map[string]interface{}{
		"project_name": artifactProject,
		"branch_name":  artifactBranch,
		"commit_hash":  artifactCommit,
		"s3_key":       presignResp.S3Key,
		"filename":     filename,
		"size_bytes":   fileInfo.Size(),
		"content_type": contentType,
	}

	confirmPayload, err := json.Marshal(confirmReq)
	if err != nil {
		return fmt.Errorf("failed to marshal confirm request: %w", err)
	}

	confirmURL := artifactBackend + "/api/v1/artifacts/confirm"
	confirmHTTPReq, err := http.NewRequest("POST", confirmURL, bytes.NewBuffer(confirmPayload))
	if err != nil {
		return fmt.Errorf("failed to create confirm request: %w", err)
	}

	confirmHTTPReq.Header.Set("Content-Type", "application/json")
	confirmHTTPReq.Header.Set("Authorization", "Bearer "+artifactToken)

	confirmResp, err := client.Do(confirmHTTPReq)
	if err != nil {
		return fmt.Errorf("failed to confirm upload: %w", err)
	}
	defer confirmResp.Body.Close()

	if confirmResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(confirmResp.Body)
		return fmt.Errorf("confirm request failed with status %d: %s", confirmResp.StatusCode, string(body))
	}

	var confirmResult struct {
		PublicURL string `json:"public_url"`
	}
	if err := json.NewDecoder(confirmResp.Body).Decode(&confirmResult); err == nil && confirmResult.PublicURL != "" {
		fmt.Printf("Artifact uploaded successfully!\nS3 Key: %s\nPublic URL: %s\n", presignResp.S3Key, confirmResult.PublicURL)
	} else {
		fmt.Printf("Artifact uploaded successfully: %s\n", presignResp.S3Key)
	}

	return nil
}

