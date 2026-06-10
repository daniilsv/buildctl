package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	apiURL string
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIClient(apiURL, apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *OpenAIClient) SummarizeCommits(ctx context.Context, messages []string) (string, error) {
	var prompt strings.Builder
	prompt.WriteString("Ты составляешь краткую сводку изменений по сообщениям git-коммитов для отчёта о сборке.\n\n" +
		"Правила:\n" +
		"1. Опирайся СТРОГО на текст исходных сообщений коммитов. Ничего не придумывай, не добавляй и не дополняй от себя.\n" +
		"2. Объединяй изменения по смыслу и убирай повторы: если несколько коммитов про одно и то же — опиши это один раз, не дублируя формулировки.\n" +
		"3. Номера задач переноси в сводку ТОЛЬКО если они дословно присутствуют в тексте коммита (например в форматах вида ABC-000, #000). Никогда не выдумывай, не подставляй и не угадывай номера задач, которых нет в исходных сообщениях. Если номера нет — не упоминай никакую задачу.\n" +
		"4. Пиши на русском языке, кратко и по делу, без воды и без повторного пересказа одного и того же.\n" +
		"5. Каждое отдельное изменение выводи отдельным пунктом с новой строки, начиная с «- ».\n" +
		"6. Не используй markdown, заголовки, стилизацию и слово «коммит». Не добавляй вводных фраз и двоеточий перед списком. В ответе верни только готовый текст сводки для отчёта.\n\n" +
		"Сообщения коммитов:\n")
	for _, msg := range messages {
		prompt.WriteString(msg + "\n")
	}

	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt.String(),
			},
		},
		"max_tokens": 8096,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return sanitizeSummary(response.Choices[0].Message.Content), nil
}

// sanitizeSummary убирает мусорные ведущие двоеточия/пробелы, которые модель
// иногда добавляет перед списком, и обрезает пустые строки по краям.
func sanitizeSummary(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Срезаем ведущее двоеточие (артефакт вида ": - изменение").
		line = strings.TrimSpace(strings.TrimPrefix(line, ":"))
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}
