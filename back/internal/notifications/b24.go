package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	db "github.com/build-assistant/back/db/gen"
)

func collectB24NotificationMaps(project *db.Project, branch *db.Branch) ([]map[string]interface{}, error) {
	var projectSettings map[string]interface{}
	if err := json.Unmarshal(project.Settings, &projectSettings); err != nil {
		return nil, fmt.Errorf("failed to parse project settings: %w", err)
	}

	var branchSettings map[string]interface{}
	if err := json.Unmarshal(branch.Settings, &branchSettings); err != nil {
		return nil, fmt.Errorf("failed to parse branch settings: %w", err)
	}

	var notifications []map[string]interface{}
	if branchRows, ok := branchSettings["b24_notifications"].([]interface{}); ok && len(branchRows) > 0 {
		for _, row := range branchRows {
			if m, ok := row.(map[string]interface{}); ok {
				notifications = append(notifications, m)
			}
		}
	} else if projectRows, ok := projectSettings["b24_notifications"].([]interface{}); ok && len(projectRows) > 0 {
		for _, row := range projectRows {
			if m, ok := row.(map[string]interface{}); ok {
				notifications = append(notifications, m)
			}
		}
	}

	return notifications, nil
}

func (n *TelegramNotifier) b24Configured() bool {
	return n.b24WebhookURL != "" && n.b24APIKey != ""
}

func (n *TelegramNotifier) sendB24(ctx context.Context, typeKey, message string) error {
	payload := map[string]string{
		"api_key":   n.b24APIKey,
		"type_key":  typeKey,
		"message":   message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.b24WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func b24ShortCommit(commitHash string) string {
	if len(commitHash) > 8 {
		return commitHash[:8]
	}
	return commitHash
}

func b24QuoteBlock(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return ">>: " + strings.ReplaceAll(text, "\n", "\n>>: ")
}

func b24LinkBB(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	label := u
	if len(label) > 64 {
		label = label[:61] + "..."
	}
	return fmt.Sprintf("[url=%s]%s[/url]", u, label)
}

func (n *TelegramNotifier) sendB24ForMessage(ctx context.Context, project *db.Project, branch *db.Branch, message string) {
	if !n.b24Configured() {
		return
	}

	rows, err := collectB24NotificationMaps(project, branch)
	if err != nil {
		slog.Error("Failed to parse settings for B24", "project", project.Name, "branch", branch.Name, "error", err)
		return
	}

	for _, row := range rows {
		tk := row["type_key"]
		if tk == nil {
			continue
		}
		typeKey := strings.TrimSpace(fmt.Sprintf("%v", tk))
		if typeKey == "" {
			continue
		}
		if err := n.sendB24(ctx, typeKey, message); err != nil {
			slog.Error("Failed to send B24 notification", "type_key", typeKey, "project", project.Name, "branch", branch.Name, "error", err)
		}
	}
}

func buildB24SuccessLegacy(projectTitle, branchName, commitHash, summary string, artifactURLs []string) string {
	var b strings.Builder
	b.WriteString("[COLOR=#008800][b]Сборка завершена[/b][/COLOR]\n\n")
	b.WriteString(fmt.Sprintf("[b]Проект:[/b] %s\n", projectTitle))
	b.WriteString(fmt.Sprintf("[b]Ветка:[/b] %s\n", branchName))
	b.WriteString(fmt.Sprintf("[b]Коммит:[/b] %s\n\n", b24ShortCommit(commitHash)))
	b.WriteString("[b]Изменения:[/b]\n")
	if q := b24QuoteBlock(summary); q != "" {
		b.WriteString(q)
		b.WriteString("\n")
	}
	if len(artifactURLs) > 0 {
		b.WriteString("\n[b]Артефакты:[/b]\n")
		for _, u := range artifactURLs {
			if link := b24LinkBB(u); link != "" {
				b.WriteString("• ")
				b.WriteString(link)
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildB24SuccessWithArtifacts(projectTitle, branchName, commitHash, authorName, summary string, artifactURLs, containerImages []string) string {
	var b strings.Builder
	b.WriteString("[COLOR=#008800][b]Сборка завершена[/b][/COLOR]\n\n")
	b.WriteString(fmt.Sprintf("[b]Проект:[/b] %s\n", projectTitle))
	b.WriteString(fmt.Sprintf("[b]Ветка:[/b] %s\n", branchName))
	b.WriteString(fmt.Sprintf("[b]Коммит:[/b] %s\n", b24ShortCommit(commitHash)))
	b.WriteString(fmt.Sprintf("[b]Автор:[/b] %s\n\n", authorName))
	b.WriteString("[b]Изменения:[/b]\n")
	if q := b24QuoteBlock(summary); q != "" {
		b.WriteString(q)
		b.WriteString("\n")
	}
	if len(artifactURLs) > 0 {
		b.WriteString("\n[b]Файловые артефакты:[/b]\n")
		for _, u := range artifactURLs {
			if link := b24LinkBB(u); link != "" {
				b.WriteString("• ")
				b.WriteString(link)
				b.WriteString("\n")
			}
		}
	}
	if len(containerImages) > 0 {
		b.WriteString("\n[b]Контейнерные образы:[/b]\n")
		for _, img := range containerImages {
			line := strings.TrimSpace(img)
			if line != "" {
				b.WriteString("• ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildB24Failed(projectTitle, branchName, commitHash, errorMessage string) string {
	var b strings.Builder
	b.WriteString("[COLOR=#cc0000][b]Сборка провалилась[/b][/COLOR]\n\n")
	b.WriteString(fmt.Sprintf("[b]Проект:[/b] %s\n", projectTitle))
	b.WriteString(fmt.Sprintf("[b]Ветка:[/b] %s\n", branchName))
	b.WriteString(fmt.Sprintf("[b]Коммит:[/b] %s\n\n", b24ShortCommit(commitHash)))
	b.WriteString("[b]Ошибка:[/b]\n")
	if q := b24QuoteBlock(errorMessage); q != "" {
		b.WriteString(q)
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildB24Test(projectTitle, branchName string) string {
	var b strings.Builder
	b.WriteString("[COLOR=#0066cc][b]Тест B24 (Bitrix)[/b][/COLOR]\n\n")
	b.WriteString(fmt.Sprintf("[b]Проект:[/b] %s\n", projectTitle))
	b.WriteString(fmt.Sprintf("[b]Ветка:[/b] %s\n\n", branchName))
	b.WriteString("Тестовое сообщение для проверки настроек уведомлений.")
	return b.String()
}

func (n *TelegramNotifier) SendTestB24Notification(ctx context.Context, project *db.Project, branch *db.Branch, typeKey string) error {
	if !n.b24Configured() {
		return fmt.Errorf("B24 is not configured (set B24_WEBHOOK_URL and B24_API_KEY)")
	}
	typeKey = strings.TrimSpace(typeKey)
	if typeKey == "" {
		return fmt.Errorf("type_key is required")
	}
	projectTitle := project.Title
	if projectTitle == "" {
		projectTitle = project.Name
	}
	msg := buildB24Test(projectTitle, branch.Name)
	return n.sendB24(ctx, typeKey, msg)
}
