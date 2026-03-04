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

	var notifications []map[string]interface{}
	if branchNotifications, ok := branchSettings["telegram_notifications"].([]interface{}); ok && len(branchNotifications) > 0 {
		for _, notif := range branchNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	} else if projectNotifications, ok := projectSettings["telegram_notifications"].([]interface{}); ok && len(projectNotifications) > 0 {
		for _, notif := range projectNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	}

	if len(notifications) == 0 {
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

	for _, notif := range notifications {
		chatID := notif["chat_id"]
		if chatID == nil {
			continue
		}
		chatIDStr := fmt.Sprintf("%v", chatID)

		var threadID *string
		if threadIDVal, ok := notif["thread_id"]; ok && threadIDVal != nil {
			threadIDStr := fmt.Sprintf("%v", threadIDVal)
			threadID = &threadIDStr
		}

		if err := n.sendMessage(ctx, chatIDStr, message, threadID); err != nil {
			return fmt.Errorf("failed to send to chat %s: %w", chatIDStr, err)
		}
	}

	return nil
}

func (n *TelegramNotifier) SendBuildNotificationWithArtifacts(ctx context.Context, project *db.Project, branch *db.Branch, commitHash, authorName, summary string, artifactURLs []string, containerImages []string) error {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var notifications []map[string]interface{}
	if branchNotifications, ok := branchSettings["telegram_notifications"].([]interface{}); ok && len(branchNotifications) > 0 {
		for _, notif := range branchNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	} else if projectNotifications, ok := projectSettings["telegram_notifications"].([]interface{}); ok && len(projectNotifications) > 0 {
		for _, notif := range projectNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	}

	if len(notifications) == 0 {
		return nil
	}

	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	message := fmt.Sprintf(
		`🎉 Сборка завершена: [Проект: %s]

Ветка: %s
Коммит: %s
Автор: %s

📝 Изменения:
%s
`,
		projectTitle,
		branch.Name,
		commitHash[:8],
		authorName,
		summary,
	)

	if len(artifactURLs) > 0 {
		message += "\n📦 Файловые артефакты:\n"
		for _, url := range artifactURLs {
			message += fmt.Sprintf("• %s\n", url)
		}
	}

	if len(containerImages) > 0 {
		message += "\n🐳 Контейнерные образы:\n"
		for _, image := range containerImages {
			message += fmt.Sprintf("• %s\n", image)
		}
	}

	for _, notif := range notifications {
		chatID := notif["chat_id"]
		if chatID == nil {
			continue
		}
		chatIDStr := fmt.Sprintf("%v", chatID)

		var threadID *string
		if threadIDVal, ok := notif["thread_id"]; ok && threadIDVal != nil {
			threadIDStr := fmt.Sprintf("%v", threadIDVal)
			threadID = &threadIDStr
		}

		if err := n.sendMessage(ctx, chatIDStr, message, threadID); err != nil {
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

	var notifications []map[string]interface{}
	if branchNotifications, ok := branchSettings["telegram_notifications"].([]interface{}); ok && len(branchNotifications) > 0 {
		for _, notif := range branchNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	} else if projectNotifications, ok := projectSettings["telegram_notifications"].([]interface{}); ok && len(projectNotifications) > 0 {
		for _, notif := range projectNotifications {
			if notifMap, ok := notif.(map[string]interface{}); ok {
				notifications = append(notifications, notifMap)
			}
		}
	}

	if len(notifications) == 0 {
		return nil
	}

	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	message := fmt.Sprintf("❌ Сборка провалилась: [Проект: %s]\n\nВетка: %s\nКоммит: %s\n\nОшибка:\n%s\n", projectTitle, branch.Name, commitHash[:8], errorMessage)

	for _, notif := range notifications {
		chatID := notif["chat_id"]
		if chatID == nil {
			continue
		}
		chatIDStr := fmt.Sprintf("%v", chatID)

		var threadID *string
		if threadIDVal, ok := notif["thread_id"]; ok && threadIDVal != nil {
			threadIDStr := fmt.Sprintf("%v", threadIDVal)
			threadID = &threadIDStr
		}

		if err := n.sendMessage(ctx, chatIDStr, message, threadID); err != nil {
			return fmt.Errorf("failed to send to chat %s: %w", chatIDStr, err)
		}
	}

	return nil
}

func (n *TelegramNotifier) SendWebhooks(ctx context.Context, project *db.Project, branch *db.Branch, commitHash string) ([]WebhookResult, error) {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return nil, fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return nil, fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var webhookURLs []interface{}
	if branchWebhooks, ok := branchSettings["webhook_urls"].([]interface{}); ok && len(branchWebhooks) > 0 {
		webhookURLs = branchWebhooks
	}
	if projectWebhooks, ok := projectSettings["webhook_urls"].([]interface{}); ok && len(projectWebhooks) > 0 {
		webhookURLs = append(webhookURLs, projectWebhooks...)
	}

	if len(webhookURLs) == 0 {
		return nil, nil
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
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var results []WebhookResult
	for _, url := range webhookURLs {
		urlStr := fmt.Sprintf("%v", url)
		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewBuffer(jsonPayload))
		if err != nil {
			results = append(results, WebhookResult{URL: urlStr, StatusCode: 0, Error: err})
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(req)
		if err != nil {
			results = append(results, WebhookResult{URL: urlStr, StatusCode: 0, Error: err})
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			results = append(results, WebhookResult{URL: urlStr, StatusCode: resp.StatusCode})
		} else {
			results = append(results, WebhookResult{
				URL:        urlStr,
				StatusCode: resp.StatusCode,
				Error:      fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)),
			})
		}
	}

	return results, nil
}

func (n *TelegramNotifier) SendTestTelegramNotification(ctx context.Context, project *db.Project, branch *db.Branch, chatID string, threadID *string) error {
	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	message := fmt.Sprintf("🧪 Тест уведомления: [Проект: %s]\n\nВетка: %s\n\nЭто тестовое сообщение для проверки настроек уведомлений.", projectTitle, branch.Name)

	if err := n.sendMessage(ctx, chatID, message, threadID); err != nil {
		return fmt.Errorf("failed to send test notification to chat %s: %w", chatID, err)
	}

	return nil
}

func (n *TelegramNotifier) SendTestWebhook(ctx context.Context, project *db.Project, branch *db.Branch, webhookURL string) error {
	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}

	payload := map[string]interface{}{
		"project":   projectTitle,
		"branch":    branch.Name,
		"test":      true,
		"timestamp": time.Now().Unix(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (n *TelegramNotifier) sendMessage(ctx context.Context, chatID, text string, threadID *string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	if threadID != nil && *threadID != "" {
		payload["message_thread_id"] = *threadID
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
