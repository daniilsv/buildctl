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
	prompt.WriteString("Ты составляешь краткую сводку изменений по сообщениям git-коммитов для отчёта о сборке не больше 3 предложений.\n\n" +
		"Правила:\n" +
		"1. Опирайся строго на текст исходных сообщений коммитов, ничего не придумывай и не добавляй от себя.\n" +
		"2. Объединяй изменения по смыслу и убирай повторы: если несколько коммитов про одно и то же — опиши это один раз, не дублируя формулировки.\n" +
		"3. Обязательно сохраняй номера задач, если они указаны в сообщениях (например, 12345, SD-123, #456, JIRA-1024), и ставь их рядом с соответствующим изменением.\n" +
		"4. Пиши на русском языке, кратко и по делу, без воды и без повторного пересказа одного и того же.\n" +
		"5. Каждое отдельное изменение выводи отдельным пунктом с новой строки, начиная с «- ».\n" +
		"6. Не используй markdown, заголовки, стилизацию и слово «коммит». В ответе верни только готовый текст сводки для отчёта.\n\n" +
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

	return response.Choices[0].Message.Content, nil
}
