package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "github.com/build-assistant/back/db/gen"
)

type TelegramNotifier struct {
	botToken string
	client   *http.Client
}

func NewTelegramNotifier(botToken string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *TelegramNotifier) SendBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, summary string, artifactURLs []string) error {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var chatIDs []interface{}
	if branchChatIDs, ok := branchSettings["telegram_chat_ids"].([]interface{}); ok && len(branchChatIDs) > 0 {
		chatIDs = branchChatIDs
	} else if projectChatIDs, ok := projectSettings["telegram_chat_ids"].([]interface{}); ok && len(projectChatIDs) > 0 {
		chatIDs = projectChatIDs
	}

	if len(chatIDs) == 0 {
		return nil
	}

	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	message := fmt.Sprintf("🎉 Сборка завершена: [Проект: %s]\n\nВетка: %s\nКоммит: %s\n\n📝 Изменения:\n%s\n", projectTitle, branch.Name, commitHash[:8], summary)

	if len(artifactURLs) > 0 {
		message += "\n📦 Артефакты:\n"
		for _, url := range artifactURLs {
			message += fmt.Sprintf("• %s\n", url)
		}
	}

	for _, chatID := range chatIDs {
		chatIDStr := fmt.Sprintf("%v", chatID)
		if err := n.sendMessage(ctx, chatIDStr, message); err != nil {
			return fmt.Errorf("failed to send to chat %s: %w", chatIDStr, err)
		}
	}

	return nil
}

func (n *TelegramNotifier) SendFailedBuildNotification(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, errorMessage string) error {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var chatIDs []interface{}
	if branchChatIDs, ok := branchSettings["telegram_chat_ids"].([]interface{}); ok && len(branchChatIDs) > 0 {
		chatIDs = branchChatIDs
	} else if projectChatIDs, ok := projectSettings["telegram_chat_ids"].([]interface{}); ok && len(projectChatIDs) > 0 {
		chatIDs = projectChatIDs
	}

	if len(chatIDs) == 0 {
		return nil
	}

	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	message := fmt.Sprintf("❌ Сборка провалилась: [Проект: %s]\n\nВетка: %s\nКоммит: %s\n\nОшибка:\n%s\n", projectTitle, branch.Name, commitHash[:8], errorMessage)

	for _, chatID := range chatIDs {
		chatIDStr := fmt.Sprintf("%v", chatID)
		if err := n.sendMessage(ctx, chatIDStr, message); err != nil {
			return fmt.Errorf("failed to send to chat %s: %w", chatIDStr, err)
		}
	}

	return nil
}

func (n *TelegramNotifier) SendWebhooks(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) error {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var webhookURLs []interface{}
	if branchWebhooks, ok := branchSettings["webhook_urls"].([]interface{}); ok && len(branchWebhooks) > 0 {
		webhookURLs = branchWebhooks
	}
	if projectWebhooks, ok := projectSettings["webhook_urls"].([]interface{}); ok && len(projectWebhooks) > 0 {
		webhookURLs = append(webhookURLs, projectWebhooks...)
	}

	if len(webhookURLs) == 0 {
		return nil
	}

	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	payload := map[string]interface{}{
		"project":     projectTitle,
		"branch":      branch.Name,
		"commit_hash": commitHash,
		"timestamp":   time.Now().Unix(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	for _, url := range webhookURLs {
		urlStr := fmt.Sprintf("%v", url)
		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewBuffer(jsonPayload))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}

	return nil
}

func (n *TelegramNotifier) sendMessage(ctx context.Context, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s - %s", resp.Status, string(body))
	}

	return nil
}
