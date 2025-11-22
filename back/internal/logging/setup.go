package logging

import (
	"fmt"
	"log/slog"
	"strings"
)

// ParseLevel парсит строку уровня логирования в slog.Level
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseFormat парсит строку формата в OutputFormat
func ParseFormat(format string) OutputFormat {
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "text":
		return FormatText
	default:
		return FormatText
	}
}

// SetupFromEnv настраивает логгер из конфига приложения
// stdoutEnabled - включить вывод в stdout
// stdoutFormat - формат для stdout (text/json)
// stdoutLevel - уровень для stdout
// fileEnabled - включить вывод в файл
// filePath - путь к файлу
// fileFormat - формат для файла (text/json)
// fileLevel - уровень для файла
// addSource - добавлять ли информацию о файле/строке
// serviceName - имя сервиса (опционально)
func SetupFromEnv(
	stdoutEnabled bool,
	stdoutFormat string,
	stdoutLevel string,
	fileEnabled bool,
	filePath string,
	fileFormat string,
	fileLevel string,
	addSource bool,
	serviceName string,
) error {
	builder := NewLoggerBuilder()

	if stdoutEnabled {
		builder.WithStdout(
			ParseFormat(stdoutFormat),
			ParseLevel(stdoutLevel),
		)
	}

	if fileEnabled {
		_, err := builder.WithFile(
			filePath,
			ParseFormat(fileFormat),
			ParseLevel(fileLevel),
		)
		if err != nil {
			return fmt.Errorf("failed to setup file logging: %w", err)
		}
	}

	if serviceName != "" {
		builder.WithServiceName(serviceName)
	}

	builder.WithSource(addSource)
	builder.Build()

	return nil
}
