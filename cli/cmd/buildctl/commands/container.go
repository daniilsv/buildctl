package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	containerToken   string
	containerBackend string
	containerProject string
	containerBranch  string
	containerCommit  string
	containerImage   string
	containerDigest  string
	containerFile    string
)

var containerCmd = &cobra.Command{
	Use:   "container register",
	Short: "Register container image",
	Long:  "Register a container image artifact in the build system",
	RunE:  runContainerRegister,
}

func init() {
	containerCmd.Flags().StringVar(&containerToken, "token", "", "Authentication token (required)")
	containerCmd.Flags().StringVar(&containerBackend, "backend", "", "Backend URL (required)")
	containerCmd.Flags().StringVar(&containerProject, "project", "", "Project name (required)")
	containerCmd.Flags().StringVar(&containerBranch, "branch", "", "Branch name (required)")
	containerCmd.Flags().StringVar(&containerCommit, "commit", "", "Commit hash (required)")
	containerCmd.Flags().StringVar(&containerImage, "image", "", "Container image with tag (e.g., registry.example.com/app:v1.0) (required)")
	containerCmd.Flags().StringVar(&containerDigest, "digest", "", "Image digest (e.g., sha256:abc123...) (required)")
	containerCmd.Flags().StringVar(&containerFile, "file", "", "Optional path to image tarball to upload")

	containerCmd.MarkFlagRequired("token")
	containerCmd.MarkFlagRequired("backend")
	containerCmd.MarkFlagRequired("project")
	containerCmd.MarkFlagRequired("branch")
	containerCmd.MarkFlagRequired("commit")
	containerCmd.MarkFlagRequired("image")
	containerCmd.MarkFlagRequired("digest")
}

func runContainerRegister(cmd *cobra.Command, args []string) error {
	var s3Key *string
	var filename *string
	var sizeBytes *int64

	// Если указан файл с образом, загружаем его в S3
	if containerFile != "" {
		file, err := os.Open(containerFile)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		fname := fileInfo.Name()
		size := fileInfo.Size()

		// Получаем presigned URL
		presignReq := map[string]string{
			"project_name": containerProject,
			"branch_name":  containerBranch,
			"commit_hash":  containerCommit,
			"filename":     fname,
		}

		jsonPayload, err := json.Marshal(presignReq)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		url := containerBackend + "/api/v1/artifacts/presign"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+containerToken)

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

		// Загружаем файл в S3
		file.Seek(0, 0)
		uploadReq, err := http.NewRequest("PUT", presignResp.UploadURL, file)
		if err != nil {
			return fmt.Errorf("failed to create upload request: %w", err)
		}

		uploadReq.ContentLength = fileInfo.Size()
		uploadReq.Header.Set("Content-Type", "application/octet-stream")

		uploadClient := &http.Client{Timeout: 10 * time.Minute}
		uploadResp, err := uploadClient.Do(uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
		defer uploadResp.Body.Close()

		if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(uploadResp.Body)
			return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(body))
		}

		s3Key = &presignResp.S3Key
		filename = &fname
		sizeBytes = &size
	}

	// Парсим имя образа и тег
	// Ищем последний ':' который не является частью порта (registry:port)
	imageName := containerImage
	imageTag := "latest"
	
	// Ищем последний ':' после последнего '/'
	lastSlash := -1
	for i := len(containerImage) - 1; i >= 0; i-- {
		if containerImage[i] == '/' {
			lastSlash = i
			break
		}
	}
	
	// Ищем ':' после последнего '/'
	for i := len(containerImage) - 1; i > lastSlash; i-- {
		if containerImage[i] == ':' {
			imageName = containerImage[:i]
			imageTag = containerImage[i+1:]
			break
		}
	}

	// Регистрируем образ в системе
	registerReq := map[string]interface{}{
		"project_name": containerProject,
		"branch_name":  containerBranch,
		"commit_hash":  containerCommit,
		"image_name":   imageName,
		"image_tag":    imageTag,
		"image_digest": containerDigest,
	}

	if filename != nil {
		registerReq["filename"] = *filename
	}
	if s3Key != nil {
		registerReq["s3_key"] = *s3Key
	}
	if sizeBytes != nil {
		registerReq["size_bytes"] = *sizeBytes
	}

	registerPayload, err := json.Marshal(registerReq)
	if err != nil {
		return fmt.Errorf("failed to marshal register request: %w", err)
	}

	registerURL := containerBackend + "/api/v1/artifacts/container"
	registerHTTPReq, err := http.NewRequest("POST", registerURL, bytes.NewBuffer(registerPayload))
	if err != nil {
		return fmt.Errorf("failed to create register request: %w", err)
	}

	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerHTTPReq.Header.Set("Authorization", "Bearer "+containerToken)

	client := &http.Client{Timeout: 30 * time.Second}
	registerResp, err := client.Do(registerHTTPReq)
	if err != nil {
		return fmt.Errorf("failed to register container: %w", err)
	}
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(registerResp.Body)
		return fmt.Errorf("register request failed with status %d: %s", registerResp.StatusCode, string(body))
	}

	var registerResult struct {
		PublicURL string `json:"public_url"`
	}
	if err := json.NewDecoder(registerResp.Body).Decode(&registerResult); err == nil && registerResult.PublicURL != "" {
		fmt.Printf("Container image registered successfully!\nImage: %s\nDigest: %s\nPublic URL: %s\n", containerImage, containerDigest, registerResult.PublicURL)
	} else {
		fmt.Printf("Container image registered successfully!\nImage: %s\nDigest: %s\n", containerImage, containerDigest)
	}

	return nil
}
